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

import(
  "fmt"
  "os"
  "log"

  "github.com/YaSuenag/hsbeat/hsperfdata"
)


func main() {
  hsperfdata_path, err := hsperfdata.GetHSPerfDataPath(os.Args[1])
  if err != nil {
    log.Fatal(err)
  }

  fmt.Printf("perfdata path: %s\n", hsperfdata_path)

  f, err := os.Open(hsperfdata_path)
  if err != nil {
    log.Fatal(err)
  }
  defer f.Close()

  prologue, err := hsperfdata.ReadPrologue(f)
  if err != nil {
    log.Fatal(err)
  }

  fmt.Printf("Magic: 0x%x\n", prologue.Magic)
  fmt.Printf("ByteOrder: %d\n", prologue.ByteOrder)
  fmt.Printf("MajorVersion: %d\n", prologue.MajorVersion)
  fmt.Printf("MinorVersion: %d\n", prologue.MinorVersion)
  fmt.Printf("Accessible: %d\n", prologue.Accessible)
  fmt.Printf("Used: %d\n", prologue.Used)
  fmt.Printf("Overflow: %d\n", prologue.Overflow)
  fmt.Printf("ModTimeStamp: %d\n", prologue.ModTimeStamp)
  fmt.Printf("EntryOffset: %d\n", prologue.EntryOffset)
  fmt.Printf("NumEntries: %d\n", prologue.NumEntries)

  f.Seek(int64(prologue.EntryOffset), os.SEEK_SET)

  entries, err := hsperfdata.ReadPerfEntry(f, prologue)
  if err != nil {
    log.Fatal(err)
  }

  var i int32
  for i = 0; i < prologue.NumEntries; i++ {
    fmt.Printf("[%d]: %c %s\n", i, entries[i].DataType, entries[i].EntryName)

    if entries[i].DataType == 'B' {
      fmt.Printf("  -> \"%s\"\n", entries[i].StringValue)
    } else {
      fmt.Printf("  -> %d\n", entries[i].LongValue)
    }

  }

}

