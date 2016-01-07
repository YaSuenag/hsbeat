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


func GetHSPerfDataPath(pid string) (string, error) {
  user, err := user.Current()
  if err != nil {
    return "", err
  }

  return filepath.Join(os.TempDir(), "hsperfdata_" + user.Username, pid), nil
}

func ReadPrologue(f *os.File) (PerfDataPrologue, error) {
  var result PerfDataPrologue
  var buf []byte = make([]byte, 32)

  n, err := f.Read(buf)
  if err != nil {
    return result, err
  } else if n != 32 {
    return result, errors.New("Could not read all prologue data.")
  }

  reader := bytes.NewReader(buf)

  binary.Read(reader, binary.BigEndian, &result.Magic)
  if result.Magic != 0xcafec0c0 {
    return result, errors.New("Invalid hsperfdata")
  }

  binary.Read(reader, binary.BigEndian, &result.ByteOrder)

  var order binary.ByteOrder
  if result.ByteOrder == 0 {
    order = binary.BigEndian
  } else {
    order = binary.LittleEndian
  }

  binary.Read(reader, order, &result.MajorVersion)
  binary.Read(reader, order, &result.MinorVersion)
  binary.Read(reader, order, &result.Accessible)
  binary.Read(reader, order, &result.Used)
  binary.Read(reader, order, &result.Overflow)
  binary.Read(reader, order, &result.ModTimeStamp)
  binary.Read(reader, order, &result.EntryOffset)
  binary.Read(reader, order, &result.NumEntries)

  return result, nil
}

func ReadEntryName(f *os.File, StartOfs int64, entry *PerfDataEntry) error {
  f.Seek(StartOfs + int64(entry.NameOffset), os.SEEK_SET)

  NameLen := entry.DataOffset - entry.NameOffset
  var buf []byte = make([]byte, NameLen)
  n, err := f.Read(buf)
  if err != nil {
    return err
  } else if n != int(NameLen) {
    return errors.New("Could not read entry name.")
  }

  n = bytes.Index(buf, []byte{0})
  entry.EntryName = string(buf[:n])

  return nil
}

func ReadEntryValueAsString(f *os.File, StartOfs int64, entry *PerfDataEntry) error {
  f.Seek(StartOfs + int64(entry.DataOffset), os.SEEK_SET)

  DataLen := entry.EntryLength - entry.DataOffset
  var buf []byte = make([]byte, DataLen)
  n, err := f.Read(buf)
  if err != nil {
    return err
  } else if n != int(DataLen) {
    return errors.New("Could not read entry value.")
  }

  n = bytes.Index(buf, []byte{0})
  entry.StringValue = string(buf[:n])

  return nil
}

func ReadEntryValueAsLong(f *os.File, StartOfs int64, prologue PerfDataPrologue, entry *PerfDataEntry) error {
  f.Seek(StartOfs + int64(entry.DataOffset), os.SEEK_SET)

  var buf []byte = make([]byte, 8)
  n, err := f.Read(buf)
  if err != nil {
    return err
  } else if n != 8 {
    return errors.New("Could not read entry value.")
  }

  reader := bytes.NewReader(buf)

  var order binary.ByteOrder
  if prologue.ByteOrder == 0 {
    order = binary.BigEndian
  } else {
    order = binary.LittleEndian
  }

  binary.Read(reader, order, &entry.LongValue)

  return nil
}

func ReadPerfEntry(f *os.File, prologue PerfDataPrologue) ([]PerfDataEntry, error){
  var buf []byte = make([]byte, 20)
  var result []PerfDataEntry = make([]PerfDataEntry, prologue.NumEntries)

  var order binary.ByteOrder
  if prologue.ByteOrder == 0 {
    order = binary.BigEndian
  } else {
    order = binary.LittleEndian
  }

  var i int32
  for i = 0; i < prologue.NumEntries; i++ {
    StartOfs, err := f.Seek(0, os.SEEK_CUR)
    if err != nil {
      return nil, err
    }

    n, err := f.Read(buf)
    if err != nil {
      return nil, err
    } else if n != 20 {
      return nil, errors.New("Could not read PerfDataEntry.")
    }

    reader := bytes.NewReader(buf)

    binary.Read(reader, order, &result[i].EntryLength)
    binary.Read(reader, order, &result[i].NameOffset)
    binary.Read(reader, order, &result[i].VectorLength)
    binary.Read(reader, order, &result[i].DataType)
    binary.Read(reader, order, &result[i].Flags)
    binary.Read(reader, order, &result[i].DataUnits)
    binary.Read(reader, order, &result[i].DataVariability)
    binary.Read(reader, order, &result[i].DataOffset)

    err = ReadEntryName(f, StartOfs, &result[i])
    if err != nil {
      return nil, err
    }

    if result[i].DataType == 'B' {
      err := ReadEntryValueAsString(f, StartOfs, &result[i])

      if err != nil {
        return nil, err
      }

    } else if result[i].DataType == 'J' {
      err := ReadEntryValueAsLong(f, StartOfs, prologue, &result[i])

      if err != nil {
        return nil, err
      }

    }

    f.Seek(StartOfs + int64(result[i].EntryLength), os.SEEK_SET)
  }

  return result, nil
}

