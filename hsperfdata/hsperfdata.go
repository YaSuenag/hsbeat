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

var globalbuf []byte


func GetHSPerfDataPath(pid string) (string, error) {
  user, err := user.Current()
  if err != nil {
    return "", err
  }

  return filepath.Join(os.TempDir(), "hsperfdata_" + user.Username, pid), nil
}

func ReadPrologue(f *os.File) (PerfDataPrologue, error) {
  var result PerfDataPrologue
  fileinfo, err := f.Stat()
  if err != nil {
      return result, err
  }

  globalbuf = make([]byte, fileinfo.Size())
  buf := make([]byte, 32)

  n, err := f.Read(buf)
  if err != nil {
    return result, err
  } else if n != 32 {
    return result, errors.New("Could not read all prologue data.")
  }

  reader := bytes.NewReader(buf)

  reader.Read(globalbuf[:4])
  result.Magic = binary.BigEndian.Uint32(globalbuf[:4])
  if result.Magic != 0xcafec0c0 {
    return result, errors.New("Invalid hsperfdata")
  }

  reader.Read(globalbuf[:4])
  result.ByteOrder = int8(globalbuf[0])
  var order binary.ByteOrder
  if result.ByteOrder == 0 {
    order = binary.BigEndian
  } else {
    order = binary.LittleEndian
  }

  result.MajorVersion = int8(globalbuf[1])
  result.MinorVersion = int8(globalbuf[2])
  result.Accessible = int8(globalbuf[3])

  reader.Read(globalbuf[:4])
  result.Used = int32(order.Uint32(globalbuf[:4]))

  reader.Read(globalbuf[:4])
  result.Overflow = int32(order.Uint32(globalbuf[:4]))

  reader.Read(globalbuf[:8])
  result.ModTimeStamp = int64(order.Uint32(globalbuf[:8]))

  reader.Read(globalbuf[:4])
  result.EntryOffset = int32(order.Uint32(globalbuf[:4]))

  reader.Read(globalbuf[:4])
  result.NumEntries = int32(order.Uint32(globalbuf[:4]))

  return result, nil
}

func readEntryName(reader *bytes.Reader, StartOfs int64, entry *PerfDataEntry) error {
  reader.Seek(StartOfs + int64(entry.NameOffset), os.SEEK_SET)

  NameLen := entry.DataOffset - entry.NameOffset
  n, err := reader.Read(globalbuf[:NameLen])
  if err != nil {
    return err
  } else if n != int(NameLen) {
    return errors.New("Could not read entry name.")
  }

  n = bytes.Index(globalbuf[:NameLen], []byte{0})
  entry.EntryName = string(globalbuf[:n])

  return nil
}

func readEntryValueAsString(reader *bytes.Reader, StartOfs int64, entry *PerfDataEntry) error {
  reader.Seek(StartOfs + int64(entry.DataOffset), os.SEEK_SET)

  DataLen := entry.EntryLength - entry.DataOffset
  n, err := reader.Read(globalbuf[:DataLen])
  if err != nil {
    return err
  } else if n != int(DataLen) {
    return errors.New("Could not read entry value.")
  }

  n = bytes.Index(globalbuf[:DataLen], []byte{0})
  entry.StringValue = string(globalbuf[:n])

  return nil
}

func readEntryValueAsLong(reader *bytes.Reader, StartOfs int64, prologue PerfDataPrologue, entry *PerfDataEntry) error {
  reader.Seek(StartOfs + int64(entry.DataOffset), os.SEEK_SET)

  var order binary.ByteOrder
  if prologue.ByteOrder == 0 {
    order = binary.BigEndian
  } else {
    order = binary.LittleEndian
  }

  reader.Read(globalbuf[:8])
  entry.LongValue = int64(order.Uint64(globalbuf[:8]))

  return nil
}

func ReadPerfEntry(f *os.File, prologue PerfDataPrologue) ([]PerfDataEntry, error){
  fileinfo, err := f.Stat()
  if err != nil {
      return nil, err
  }

  var buf []byte = make([]byte, fileinfo.Size() - 32)
  f.Read(buf)

  var result []PerfDataEntry = make([]PerfDataEntry, prologue.NumEntries)

  var order binary.ByteOrder
  if prologue.ByteOrder == 0 {
    order = binary.BigEndian
  } else {
    order = binary.LittleEndian
  }

  reader := bytes.NewReader(buf)
  for i := 0; i < int(prologue.NumEntries); i++ {
    StartOfs, err := reader.Seek(0, os.SEEK_CUR)
    if err != nil {
      return nil, err
    }

    reader.Read(globalbuf[:4])
    result[i].EntryLength = int32(order.Uint32(globalbuf[:4]))
    reader.Read(globalbuf[:4])
    result[i].NameOffset = int32(order.Uint32(globalbuf[:4]))
    reader.Read(globalbuf[:4])
    result[i].VectorLength = int32(order.Uint32(globalbuf[:4]))

    reader.Read(globalbuf[:4])
    result[i].DataType = int8(globalbuf[0])
    result[i].Flags = int8(globalbuf[1])
    result[i].DataUnits = int8(globalbuf[2])
    result[i].DataVariability = int8(globalbuf[3])

    reader.Read(globalbuf[:4])
    result[i].DataOffset = int32(order.Uint32(globalbuf[:4]))

    err = readEntryName(reader, StartOfs, &result[i])
    if err != nil {
      return nil, err
    }

    if result[i].DataType == 'B' {
      err := readEntryValueAsString(reader, StartOfs, &result[i])

      if err != nil {
        return nil, err
      }

    } else if result[i].DataType == 'J' {
      err := readEntryValueAsLong(reader, StartOfs, prologue, &result[i])

      if err != nil {
        return nil, err
      }

    }

    reader.Seek(StartOfs + int64(result[i].EntryLength), os.SEEK_SET)
  }

  return result, nil
}

