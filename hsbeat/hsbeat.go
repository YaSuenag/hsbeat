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
}


func (this *HSBeat) Config(b *beat.Beat) error {
  return nil
}

func (this *HSBeat) Setup(b *beat.Beat) error {
  var err error

  this.HSPerfDataPath, err = hsperfdata.GetHSPerfDataPath(this.Pid)
  return err
}

func (this *HSBeat) getAndPublish(b *beat.Beat) error {
  f, err := os.Open(this.HSPerfDataPath)
  if err != nil {
    return err
  }
  defer f.Close()

  timestamp :=  common.Time(time.Now())

  prologue, err := hsperfdata.ReadPrologue(f)
  if err != nil {
    return err
  }

  f.Seek(int64(prologue.EntryOffset), os.SEEK_SET)

  entries, err := hsperfdata.ReadPerfEntry(f, prologue)
  if err != nil {
    return err
  }

  for _, entry := range entries {
    event := common.MapStr{"@timestamp": timestamp,
                           "type": "hsbeat",
                           "pid": this.Pid}

    event["name"] = entry.EntryName
    event["data_type"] = entry.DataType

    if entry.DataType == 'J' {
      event["long_val"] = entry.LongValue
    } else {
      event["str_val"] = entry.StringValue
    }

    b.Events.PublishEvent(event)
  }


  return nil
}

func (this *HSBeat) Run(b *beat.Beat) error {

  for !this.ShouldTerminate {
    err := this.getAndPublish(b)
    if err != nil {
      return err
    }

    time.Sleep(this.Interval)
  }

  return nil
}

func (this *HSBeat) Cleanup(b *beat.Beat) error {
  return nil
}

func (this *HSBeat) Stop() {
  this.ShouldTerminate = true
}

