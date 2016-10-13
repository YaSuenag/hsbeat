package hsperfdata

import (
	"os"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/metricbeat/mb"
)

// init registers the MetricSet with the central registry.
// The New method will be called after the setup of the module and before starting to fetch data
func init() {
	if err := mb.Registry.AddMetricSet("hotspot", "hsperfdata", New); err != nil {
		panic(err)
	}
}

// MetricSet type defines all fields of the MetricSet
// As a minimum it must inherit the mb.BaseMetricSet fields, but can be extended with
// additional entries. These variables can be used to persist data or configuration between
// multiple fetch calls.
type MetricSet struct {
	mb.BaseMetricSet
	parser *HSPerfData
	previousData map[string]int64
	hsPerfDataPath string
	isFirst bool
	pid string
}

// New create a new instance of the MetricSet
// Part of new is also setting up the configuration by processing additional
// configuration entries if needed.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {

	config := struct{
		ForceCachedEntries []string `config:"force_collect"`
		Pid string `config:"pid"`
	}{
		ForceCachedEntries: []string{},
		Pid: "0",
	}

	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}

	inst := &HSPerfData{}
	inst.ForceCachedEntryName = make(map[string]int)
	for _, entry := range config.ForceCachedEntries {
		inst.ForceCachedEntryName[entry] = 1
	}

	perfDataPath, err := GetHSPerfDataPath(config.Pid)
	if err != nil {
		return nil, err
	}

	prevData := make(map[string]int64)

	return &MetricSet{
		BaseMetricSet: base,
		parser: inst,
		previousData: prevData,
		hsPerfDataPath: perfDataPath,
		isFirst: true,
		pid: config.Pid,
	}, nil
}

func (m *MetricSet) buildMapStr(entries []PerfDataEntry) common.MapStr {
	event := common.MapStr{"pid": m.pid}

	for _, entry := range entries {
		if entry.DataType == 'J' {
			event[entry.EntryName] = entry.LongValue
			prev, exists := m.previousData[entry.EntryName]

			if exists {
				event[entry.EntryName + "/diff"] = entry.LongValue - prev
			}

			m.previousData[entry.EntryName] = entry.LongValue
		} else {
			event[entry.EntryName] = entry.StringValue
		}
	}

	return event
}

func (m *MetricSet) publishAll() (common.MapStr, error) {
	f, err := os.Open(m.hsPerfDataPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = m.parser.ReadPrologue(f)
	if err != nil {
		return nil, err
	}

	f.Seek(int64(m.parser.Prologue.EntryOffset), os.SEEK_SET)
	result, err := m.parser.ReadAllEntry(f)
	if err != nil {
		return nil, err
	}

	return m.buildMapStr(result), nil
}

func (m *MetricSet) publishCached() (common.MapStr, error) {
	f, err := os.Open(m.hsPerfDataPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	f.Seek(int64(m.parser.Prologue.EntryOffset), os.SEEK_SET)
	result, err := m.parser.ReadCachedEntry(f)
	if err != nil {
		return nil, err
	}

	return m.buildMapStr(result), nil
}

// Fetch methods implements the data gathering and data conversion to the right format
// It returns the event which is then forward to the output. In case of an error, a
// descriptive error must be returned.
func (m *MetricSet) Fetch() (common.MapStr, error) {
	var event common.MapStr
	var err error
	if m.isFirst {
		event, err = m.publishAll()
		m.isFirst = false
	} else {
		event, err = m.publishCached()
	}

	if err != nil {
		return nil, err
	}

	return event, nil
}
