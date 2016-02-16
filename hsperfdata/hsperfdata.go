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
package hsperfdata

import(
  "os"
  "os/user"
  "path/filepath"
  "bytes"
  "encoding/binary"
  "errors"
)


type PerfDataPrologue struct{
  Magic uint32
  ByteOrder int8
  MajorVersion int8
  MinorVersion int8
  Accessible int8
  Used int32
  Overflow int32
  ModTimeStamp int64
  EntryOffset int32
  NumEntries int32
}

type PerfDataEntry struct{
  EntryLength int32
  NameOffset int32
  VectorLength int32
  DataType int8
  Flags int8
  DataUnits int8
  DataVariability int8
  DataOffset int32
  EntryName string
  StringValue string
  LongValue int64
}

type HSPerfData struct {
  globalbuf []byte
  Prologue PerfDataPrologue
  byteOrder binary.ByteOrder
  entryCache []PerfDataEntry
}


func GetHSPerfDataPath(pid string) (string, error) {
  user, err := user.Current()
  if err != nil {
    return "", err
  }

  return filepath.Join(os.TempDir(), "hsperfdata_" + user.Username, pid), nil
}

func (this *HSPerfData) ReadPrologue(f *os.File) error {
  fileinfo, err := f.Stat()
  if err != nil {
      return err
  }

  this.globalbuf = make([]byte, fileinfo.Size())
  buf := make([]byte, 32)

  n, err := f.Read(buf)
  if err != nil {
    return err
  } else if n != 32 {
    return errors.New("Could not read all prologue data.")
  }

  reader := bytes.NewReader(buf)

  reader.Read(this.globalbuf[:4])
  this.Prologue.Magic = binary.BigEndian.Uint32(this.globalbuf[:4])
  if this.Prologue.Magic != 0xcafec0c0 {
    return errors.New("Invalid hsperfdata")
  }

  reader.Read(this.globalbuf[:4])
  this.Prologue.ByteOrder = int8(this.globalbuf[0])
  if this.Prologue.ByteOrder == 0 {
    this.byteOrder = binary.BigEndian
  } else {
    this.byteOrder = binary.LittleEndian
  }

  this.Prologue.MajorVersion = int8(this.globalbuf[1])
  this.Prologue.MinorVersion = int8(this.globalbuf[2])
  this.Prologue.Accessible = int8(this.globalbuf[3])

  reader.Read(this.globalbuf[:4])
  this.Prologue.Used = int32(this.byteOrder.Uint32(this.globalbuf[:4]))

  reader.Read(this.globalbuf[:4])
  this.Prologue.Overflow = int32(this.byteOrder.Uint32(this.globalbuf[:4]))

  reader.Read(this.globalbuf[:8])
  this.Prologue.ModTimeStamp = int64(this.byteOrder.Uint32(this.globalbuf[:8]))

  reader.Read(this.globalbuf[:4])
  this.Prologue.EntryOffset = int32(this.byteOrder.Uint32(this.globalbuf[:4]))

  reader.Read(this.globalbuf[:4])
  this.Prologue.NumEntries = int32(this.byteOrder.Uint32(this.globalbuf[:4]))

  return nil
}

func (this *HSPerfData) readEntryName(reader *bytes.Reader, StartOfs int64, entry *PerfDataEntry) error {
  reader.Seek(StartOfs + int64(entry.NameOffset), os.SEEK_SET)

  NameLen := entry.DataOffset - entry.NameOffset
  n, err := reader.Read(this.globalbuf[:NameLen])
  if err != nil {
    return err
  } else if n != int(NameLen) {
    return errors.New("Could not read entry name.")
  }

  n = bytes.Index(this.globalbuf[:NameLen], []byte{0})
  for i := 0; i < n; i++ {  // Convert '.' to '/'
    if this.globalbuf[i] == '.' {
      this.globalbuf[i] = '/'
    }
  }
  entry.EntryName = string(this.globalbuf[:n])

  return nil
}

func (this *HSPerfData) readEntryValueAsString(reader *bytes.Reader, StartOfs int64, entry *PerfDataEntry) error {
  reader.Seek(StartOfs + int64(entry.DataOffset), os.SEEK_SET)

  DataLen := entry.EntryLength - entry.DataOffset
  n, err := reader.Read(this.globalbuf[:DataLen])
  if err != nil {
    return err
  } else if n != int(DataLen) {
    return errors.New("Could not read entry value.")
  }

  n = bytes.Index(this.globalbuf[:DataLen], []byte{0})
  entry.StringValue = string(this.globalbuf[:n])

  return nil
}

func (this *HSPerfData) readEntryValueAsLong(reader *bytes.Reader, StartOfs int64, entry *PerfDataEntry) error {
  reader.Seek(StartOfs + int64(entry.DataOffset), os.SEEK_SET)
  reader.Read(this.globalbuf[:8])
  entry.LongValue = int64(this.byteOrder.Uint64(this.globalbuf[:8]))

  return nil
}

func (this *HSPerfData) ReadAllEntry(f *os.File) ([]PerfDataEntry, error){
  fileinfo, err := f.Stat()
  if err != nil {
      return nil, err
  }

  var buf []byte = make([]byte, fileinfo.Size() - 32)
  f.Read(buf)

  var result []PerfDataEntry = make([]PerfDataEntry, this.Prologue.NumEntries)
  this.entryCache = make([]PerfDataEntry, 0, this.Prologue.NumEntries)

  reader := bytes.NewReader(buf)
  for i := 0; i < int(this.Prologue.NumEntries); i++ {
    StartOfs, err := reader.Seek(0, os.SEEK_CUR)
    if err != nil {
      return nil, err
    }

    reader.Read(this.globalbuf[:4])
    result[i].EntryLength = int32(this.byteOrder.Uint32(this.globalbuf[:4]))
    reader.Read(this.globalbuf[:4])
    result[i].NameOffset = int32(this.byteOrder.Uint32(this.globalbuf[:4]))
    reader.Read(this.globalbuf[:4])
    result[i].VectorLength = int32(this.byteOrder.Uint32(this.globalbuf[:4]))

    reader.Read(this.globalbuf[:4])
    result[i].DataType = int8(this.globalbuf[0])
    result[i].Flags = int8(this.globalbuf[1])
    result[i].DataUnits = int8(this.globalbuf[2])
    result[i].DataVariability = int8(this.globalbuf[3])

    reader.Read(this.globalbuf[:4])
    result[i].DataOffset = int32(this.byteOrder.Uint32(this.globalbuf[:4]))

    err = this.readEntryName(reader, StartOfs, &result[i])
    if err != nil {
      return nil, err
    }

    if result[i].DataType == 'B' {
      err := this.readEntryValueAsString(reader, StartOfs, &result[i])

      if err != nil {
        return nil, err
      }

    } else if result[i].DataType == 'J' {
      err := this.readEntryValueAsLong(reader, StartOfs, &result[i])

      if err != nil {
        return nil, err
      }

    }

    if result[i].DataVariability != 1 {  // Modifiable value
      this.entryCache = append(this.entryCache, result[i])
    }

    reader.Seek(StartOfs + int64(result[i].EntryLength), os.SEEK_SET)
  }

  return result, nil
}

func (this *HSPerfData) ReadCachedEntry(f *os.File) ([]PerfDataEntry, error){
  fileinfo, err := f.Stat()
  if err != nil {
      return nil, err
  }

  var buf []byte = make([]byte, fileinfo.Size() - 32)
  f.Read(buf)

  var result []PerfDataEntry = make([]PerfDataEntry, len(this.entryCache))

  reader := bytes.NewReader(buf)
  for i, entry := range this.entryCache {
    StartOfs, err := reader.Seek(0, os.SEEK_CUR)
    if err != nil {
      return nil, err
    }

    result[i] = entry

    if result[i].DataType == 'B' {
      err := this.readEntryValueAsString(reader, StartOfs, &result[i])

      if err != nil {
        return nil, err
      }

    } else if result[i].DataType == 'J' {
      err := this.readEntryValueAsLong(reader, StartOfs, &result[i])

      if err != nil {
        return nil, err
      }

    }

    reader.Seek(StartOfs + int64(result[i].EntryLength), os.SEEK_SET)
  }

  return result, nil
}

