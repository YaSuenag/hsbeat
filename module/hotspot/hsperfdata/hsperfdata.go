package hsperfdata

import (
	"os"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/metricbeat/mb"

	"github.com/YaSuenag/hsbeat/utils/multierror"
)

const DEBUG_SELECTOR = "hsbeat"

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
	forceCachedEntries []string
	pid string
	procs map[string]ProcStats // PID to ProcStats map
}

// ProcStats type holds data for a given Java process (PID)
type ProcStats struct {
	pid string
	parser *HSPerfData
	previousData map[string]int64
	hsPerfDataPath string
	isFirst bool
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

	return &MetricSet{
		BaseMetricSet: base,
		pid: config.Pid,
		forceCachedEntries: config.ForceCachedEntries,
		procs: make(map[string]ProcStats, 0),
	}, nil
}

func (m *MetricSet) attachJavaProc(pid string) error {

	if _, exists := m.procs[pid]; exists {
		return nil // pid already attached
	}

	logp.Debug(DEBUG_SELECTOR, "Attaching java process: %v", pid)

	inst := &HSPerfData{}
	inst.ForceCachedEntryName = make(map[string]int)
	for _, entry := range m.forceCachedEntries {
		inst.ForceCachedEntryName[entry] = 1
	}

	perfDataPath, err := GetHSPerfDataPath(pid)
	if err != nil {
		return err
	}

	prevData := make(map[string]int64)

	procStats := ProcStats{
		pid: pid,
		parser: inst,
		previousData: prevData,
		hsPerfDataPath: perfDataPath,
		isFirst: true,
	}

	m.procs[pid] = procStats

	return nil
}

func (m *MetricSet) detachJavaProc(pid string) {
	logp.Debug(DEBUG_SELECTOR, "Detaching java process: %v", pid)
	delete(m.procs, pid)
}

// if configured pid equals 0 look for all running java processes
// else, only fetch for the configured pid
// the method updates MetricSet.procs map
func (m *MetricSet) findAndAttachJavaProcs() error {
	if m.pid != "0" {
		logp.Debug(DEBUG_SELECTOR, "Fetching data for only one pid: %v", m.pid)
		if err := m.attachJavaProc(m.pid); err != nil {
			return err
		}
	} else { // need to look for Java Processes
		logp.Debug(DEBUG_SELECTOR, "Fetching data for multiple java processes")
		runningPids, err := GetHSPerfPids()
		if err != nil {
			return err
		}
		logp.Debug(DEBUG_SELECTOR, "Found %v running java processes", len(runningPids))
		for _, pid := range runningPids {
			if err := m.attachJavaProc(pid); err != nil {
				logp.Err("Could not attach java process with pid: %v", pid, err)
				// continue with other processes
			}
		}
		// detach any proc that is no longer running
		for attachedPid, _ := range m.procs {
			found := false
			for _, runningPid := range runningPids {
				if runningPid == attachedPid {
					found = true
					break
				}
			}
			if !found {
				m.detachJavaProc(attachedPid)
			}
		}
	}
	return nil
}

func (p *ProcStats) buildMapStr(entries []PerfDataEntry) common.MapStr {
	event := common.MapStr{"pid": p.pid}

	for _, entry := range entries {
		if entry.DataType == 'J' {
			event[entry.EntryName] = entry.LongValue
			prev, exists := p.previousData[entry.EntryName]

			if exists {
				event[entry.EntryName + "/diff"] = entry.LongValue - prev
			}

			p.previousData[entry.EntryName] = entry.LongValue
		} else {
			event[entry.EntryName] = entry.StringValue
		}
	}

	return event
}

func (p *ProcStats) publishAll() (common.MapStr, error) {
	f, err := os.Open(p.hsPerfDataPath)
	if err != nil {
		if os.IsPermission(err) {
			logp.Debug(DEBUG_SELECTOR, "Could not open %v due to perimissions error, if you want to collect data from all users hsbeat needs to run as root (%v)", p.hsPerfDataPath, err)
		} else {
			logp.Debug(DEBUG_SELECTOR, "Could not open %v (%v)", p.hsPerfDataPath, err)
		}
		return nil, err
	}
	defer f.Close()

	err = p.parser.ReadPrologue(f)
	if err != nil {
		return nil, err
	}

	f.Seek(int64(p.parser.Prologue.EntryOffset), os.SEEK_SET)
	result, err := p.parser.ReadAllEntry(f)
	if err != nil {
		return nil, err
	}

	return p.buildMapStr(result), nil
}

func (p *ProcStats) publishCached() (common.MapStr, error) {
	f, err := os.Open(p.hsPerfDataPath)
	if err != nil {
		if os.IsPermission(err) {
			logp.Debug(DEBUG_SELECTOR, "Could not open %v due to perimissions error, if you want to collect data from all users hsbeat needs to run as root (%v)", p.hsPerfDataPath, err)
		} else {
			logp.Debug(DEBUG_SELECTOR, "Could not open %v (%v)", p.hsPerfDataPath, err)
		}
		return nil, err
	}
	defer f.Close()

	f.Seek(int64(p.parser.Prologue.EntryOffset), os.SEEK_SET)
	result, err := p.parser.ReadCachedEntry(f)
	if err != nil {
		return nil, err
	}

	return p.buildMapStr(result), nil
}

// Fetch methods implements the data gathering and data conversion to the right format
// It returns a list of events which is then forward to the output. In case of an error, a
// descriptive error must be returned.
// Map of java processes is updated at each fetch interval
// Errors are accumlated and returned only if no events were collected, otherwise events are returned and errors are just logged
func (m *MetricSet) Fetch() ([]common.MapStr, error) {

	errors := new(multierror.MultiError)

	if err := m.findAndAttachJavaProcs(); err != nil {
		errors.Append(err) // accumulate errors
	}

	events := make([]common.MapStr, 0, len(m.procs))
	for _, p := range m.procs {
		var event common.MapStr
		var err error
		if p.isFirst {
			event, err = p.publishAll()
			p.isFirst = false
		} else {
			event, err = p.publishCached()
		}

		if err != nil {
			errors.Append(err) // accumulate errors
		} else {
			events = append(events, event)
		}
	}

	if errors.HasErrors() {
		logp.Debug(DEBUG_SELECTOR, "Could not fetch metrics for all processes. Error(s) found: %v", errors.String())
		if len(events) == 0 {
			return nil, errors // return error only we didn't collect any event
		}
	}

	return events, nil
}
