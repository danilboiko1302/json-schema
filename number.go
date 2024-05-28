package jsonschema

import (
	"errors"
)

var numberValidation map[string]rawValidation = map[string]rawValidation{
	"minimum": {
		function: minimum,
	},
	"exclusiveMinimum": {
		function: exclusiveMinimum,
	},
	"maximum": {
		function: maximum,
	},
	"exclusiveMaximum": {
		function: exclusiveMaximum,
	},
	"multipleOf": {
		function: multipleOf,
	},
}

func multipleOf(value any) (func(a any) *Error, error) {
	v, ok := value.(float64)
	if !ok {
		return nil, errors.New("multipleOf requires number")
	}

	return func(a any) *Error {
		check := a.(float64) / v
		if float64(int(check)) == check {
			return nil
		}

		return NewError(v, a.(float64))
	}, nil
}

func exclusiveMaximum(max any) (func(a any) *Error, error) {
	v, ok := max.(float64)
	if !ok {
		return nil, errors.New("exclusiveMaximum requires number")
	}

	return func(a any) *Error {
		if v > a.(float64) {
			return nil
		}

		return NewError(v, a.(float64))
	}, nil
}

func maximum(max any) (func(a any) *Error, error) {
	v, ok := max.(float64)
	if !ok {
		return nil, errors.New("maximum requires number")
	}

	return func(a any) *Error {
		if v >= a.(float64) {
			return nil
		}

		return NewError(v, a.(float64))
	}, nil
}

func exclusiveMinimum(min any) (func(a any) *Error, error) {
	v, ok := min.(float64)
	if !ok {
		return nil, errors.New("minimum requires number")
	}

	return func(a any) *Error {
		if v < a.(float64) {
			return nil
		}

		return NewError(v, a.(float64))
	}, nil
}

func minimum(min any) (func(a any) *Error, error) {
	v, ok := min.(float64)
	if !ok {
		return nil, errors.New("minimum requires number")
	}

	return func(a any) *Error {
		if v <= a.(float64) {
			return nil
		}

		return NewError(v, a.(float64))
	}, nil
}
