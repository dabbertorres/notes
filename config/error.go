package config

import "fmt"

type fieldError struct {
	Key string
	Msg string
}

func (e fieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Key, e.Msg)
}

func (e fieldError) qualify(parent string) fieldError {
	e.Key = parent + e.Key
	return e
}

type fieldErrorList []fieldError

func (l fieldErrorList) qualify(parent string) fieldErrorList {
	for i, e := range l {
		l[i] = e.qualify(parent)
	}

	return l
}

func (l fieldErrorList) asErrors() (out []error) {
	out = make([]error, len(l))
	for i, e := range l {
		out[i] = e
	}
	return out
}
