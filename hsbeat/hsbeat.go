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

  "github.com/elastic/libbeat/beat"
  "github.com/elastic/libbeat/common"

  "github.com/YaSuenag/hsbeat/hsperfdata"
)


type HSBeat struct {
  Pid string
  Interval time.Duration
  HSPerfDataPath string
  ShouldTerminate bool
  PreviousData map[string]int64
  PerfData *hsperfdata.HSPerfData
}


func (this *HSBeat) Config(b *beat.Beat) error {
  return nil
}

func (this *HSBeat) Setup(b *beat.Beat) error {
  var err error

  this.PreviousData = make(map[string]int64)
  this.HSPerfDataPath, err = hsperfdata.GetHSPerfDataPath(this.Pid)
  this.PerfData = &hsperfdata.HSPerfData{}
  return err
}

func (this *HSBeat) publish(b *beat.Beat, entries []hsperfdata.PerfDataEntry) error {
  timestamp :=  common.Time(time.Now())
  event := common.MapStr{"@timestamp": timestamp,
                         "type": "hsbeat",
                         "pid": this.Pid}

  for _, entry := range entries {
    if entry.DataType == 'J' {
      event[entry.EntryName] = entry.LongValue
      prev, exists := this.PreviousData[entry.EntryName]

      if exists {
        event[entry.EntryName + ",diff"] = entry.LongValue - prev
      }

      this.PreviousData[entry.EntryName] = entry.LongValue
    } else {
      event[entry.EntryName] = entry.StringValue
    }
  }

  b.Events.PublishEvent(event)

  return nil
}

func (this *HSBeat) publishAll(b *beat.Beat) error {
  f, err := os.Open(this.HSPerfDataPath)
  if err != nil {
    return err
  }
  defer f.Close()

  err = this.PerfData.ReadPrologue(f)
  if err != nil {
    return err
  }

  f.Seek(int64(this.PerfData.Prologue.EntryOffset), os.SEEK_SET)
  result, err := this.PerfData.ReadAllEntry(f)
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
  f, err := os.Open(this.HSPerfDataPath)
  if err != nil {
    return err
  }
  defer f.Close()

  f.Seek(int64(this.PerfData.Prologue.EntryOffset), os.SEEK_SET)
  result, err := this.PerfData.ReadCachedEntry(f)
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
  err := this.publishAll(b)
  if err != nil {
    return err
  }

  for !this.ShouldTerminate {
    time.Sleep(this.Interval)

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
  this.ShouldTerminate = true
}

