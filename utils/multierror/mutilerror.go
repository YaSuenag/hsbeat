package multierror

type MultiError struct {
	errors []error
}


func (e *MultiError) Error() string {
	return e.format()
}

func (e *MultiError) String() string {
	return e.format()
}

func (e *MultiError) Append(err error) {
	if err != nil {
		e.errors = append(e.errors, err)
	}
}

func (e *MultiError) HasErrors() bool {
	return len(e.errors) > 0
}

func (e *MultiError) Count() int {
  return len(e.errors)
}

func (e *MultiError) format() string {
	if len(e.errors) == 0 {
		return "No errors"
	}

	if len(e.errors) == 1 {
		return e.errors[0].Error()
	}

	msg := "multiple errors:"
	for _, err := range e.errors {
		msg += "\n" + err.Error()
	}
	return msg
}
