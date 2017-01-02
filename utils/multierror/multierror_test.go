package multierror

import (
  "testing"
  "errors"
  "io"
)

func TestHasErrors(t *testing.T) {
  e := new(MultiError)
  assertFalse(t, e.HasErrors())
  e.Append(errors.New("first error"))
  assertTrue(t, e.HasErrors())
}

func TestCount(t *testing.T) {
  e := new(MultiError)
  assertEquals(t, 0, e.Count())
  e.Append(errors.New("first error"))
  assertEquals(t, 1, e.Count())
  e.Append(errors.New("second error"))
  assertEquals(t, 2, e.Count())

}

func TestAppend(t *testing.T) {
  e := new(MultiError)
  assertFalse(t, e.HasErrors())
  e.Append(errors.New("first error"))
  assertTrue(t, e.HasErrors())
  e.Append(errors.New("second error"))
  assertEquals(t, 2, e.Count())
  e.Append(io.EOF) // test with some standard error
  assertEquals(t, 3, e.Count())
}

func TestDontAppendNil(t *testing.T) {
  e := new(MultiError)
  assertFalse(t, e.HasErrors())
  e.Append(errors.New("first error"))
  assertEquals(t, 1, e.Count())
  e.Append(nil)
  assertEquals(t, 1, e.Count())
  e.Append(errors.New("second actual error"))
  assertEquals(t, 2, e.Count())

}

func TestError(t *testing.T) {
  e := new(MultiError)
  e.Append(errors.New("first error"))
  e.Append(errors.New("second error"))
  assertEquals(t, "multiple errors:\nfirst error\nsecond error", e.Error())
}

func TestStringer(t *testing.T) {
  e := new(MultiError)
  e.Append(errors.New("first error"))
  e.Append(errors.New("second error"))
  assertEquals(t, "multiple errors:\nfirst error\nsecond error", e.String())
}

func assertEquals(t *testing.T, expected interface{}, actual interface{}) {
  if expected != actual {
    t.Errorf("%v is not equal to %v", expected, actual)
  }
}

func assertTrue(t *testing.T, actual bool) {
  if ! actual {
    t.Errorf("expected true got false")
  }
}

func assertFalse(t *testing.T, actual bool) {
  if actual {
    t.Errorf("expected false got true")
  }
}
