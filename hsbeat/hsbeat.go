/*
 * Copyright (C) 2016 Yasumasa Suenaga
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation; either version 2
 * of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston,
 * MA 02110-1301, USA.
 */
package hsbeat

import(
  "os"
  "time"

  "github.com/elastic/beats/libbeat/beat"
  "github.com/elastic/beats/libbeat/common"

  "github.com/YaSuenag/hsbeat/hsperfdata"
)


type HSBeat struct {
  Pid string
  Interval time.Duration
  hsPerfDataPath string
  previousData map[string]int64
  perfData *hsperfdata.HSPerfData
  cachedEvent common.MapStr
  ch chan struct{}
}


func (this *HSBeat) Config(b *beat.Beat) error {
  return nil
}

func (this *HSBeat) Setup(b *beat.Beat) error {
  var err error

  this.previousData = make(map[string]int64)
  this.hsPerfDataPath, err = hsperfdata.GetHSPerfDataPath(this.Pid)
  this.perfData = &hsperfdata.HSPerfData{}
  this.ch = make(chan struct{})
  return err
}

func (this *HSBeat) publish(b *beat.Beat, entries []hsperfdata.PerfDataEntry) error {
  var event common.MapStr
  if this.cachedEvent == nil {
    event = common.MapStr{"type": "hsbeat", "pid": this.Pid}
  } else {
    event = this.cachedEvent
  }

  event["@timestamp"] = common.Time(time.Now())

  for _, entry := range entries {
    if entry.DataType == 'J' {
      event[entry.EntryName] = entry.LongValue
      prev, exists := this.previousData[entry.EntryName]

      if exists {
        event[entry.EntryName + "/diff"] = entry.LongValue - prev
      }

      this.previousData[entry.EntryName] = entry.LongValue
    } else {
      event[entry.EntryName] = entry.StringValue
    }
  }

  b.Events.PublishEvent(event)

  return nil
}

func (this *HSBeat) publishAll(b *beat.Beat) error {
  this.cachedEvent = nil

  f, err := os.Open(this.hsPerfDataPath)
  if err != nil {
    return err
  }
  defer f.Close()

  err = this.perfData.ReadPrologue(f)
  if err != nil {
    return err
  }

  f.Seek(int64(this.perfData.Prologue.EntryOffset), os.SEEK_SET)
  result, err := this.perfData.ReadAllEntry(f)
  if err != nil {
    return err
  }

  err = this.publish(b, result)
  if err != nil {
    return err
  }

  return nil
}

func (this *HSBeat) publishCached(b *beat.Beat) error {
  this.cachedEvent = common.MapStr{"type": "hsbeat", "pid": this.Pid}

  f, err := os.Open(this.hsPerfDataPath)
  if err != nil {
    return err
  }
  defer f.Close()

  f.Seek(int64(this.perfData.Prologue.EntryOffset), os.SEEK_SET)
  result, err := this.perfData.ReadCachedEntry(f)
  if err != nil {
    return err
  }

  err = this.publish(b, result)
  if err != nil {
    return err
  }

  return nil
}

func (this *HSBeat) Run(b *beat.Beat) error {
  if b.Events == nil {
    panic("Beat.Events is nil")
  }

  ticker := time.NewTicker(this.Interval)
  defer ticker.Stop()

  err := this.publishAll(b)
  if err != nil {
    return err
  }

  for {

    select {
      case <- this.ch:
        return nil
      case <- ticker.C:
    }

    err := this.publishCached(b)
    if err != nil {
      return err
    }

  }

  return nil
}

func (this *HSBeat) Cleanup(b *beat.Beat) error {
  return nil
}

func (this *HSBeat) Stop() {
  close(this.ch)
}

