package jsonschema

import "fmt"

func NewError(expect, got any) *Error {
	return &Error{
		expect: expect,
		got:    got,
	}
}

type Error struct {
	expect any
	got    any
	name   string
}

func (e *Error) Error() string {
	return fmt.Sprintf("failed to validate %s; got: %v, expected: %v", e.name, e.got, e.expect)
}

func (e *Error) SetName(name string) *Error {
	if e.name != "" {
		return e
	}

	e.name = name
	return e
}
