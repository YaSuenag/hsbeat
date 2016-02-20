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
package main

import (
  "os"
  "log"
  "strconv"
  "time"

  "github.com/elastic/beats/libbeat/beat"

  hsbeat "github.com/YaSuenag/hsbeat/hsbeat"

  //"runtime/pprof"
)


func main() {
  interval, err := strconv.Atoi(os.Args[2])
  if err != nil {
    log.Fatal(err)
  }

/*
  prof, err := os.Create("hsbeat.pprof")
  if err != nil {
    log.Fatal(err)
  }
  pprof.StartCPUProfile(prof)
  defer pprof.StopCPUProfile()

  mprof, err := os.Create("hsbeat.mprof")
  if err != nil {
    log.Fatal(err)
  }
  defer pprof.WriteHeapProfile(mprof)
*/

  hb :=&hsbeat.HSBeat{os.Args[1], time.Duration(interval), "",
                                                       false, nil, nil, nil}
  b := beat.NewBeat("hsbeat", "0.1.0", hb)
  b.CommandLineSetup()
  b.LoadConfig()
  b.Run()
}

