package jsonschema

import "errors"

var integerValidation map[string]rawValidation = map[string]rawValidation{
	"minimum": {
		function: minimumInteger,
	},
	"exclusiveMinimum": {
		function: exclusiveMinimumInteger,
	},
	"maximum": {
		function: maximumInteger,
	},
	"exclusiveMaximum": {
		function: exclusiveMaximumInteger,
	},
	"multipleOf": {
		function: multipleOfInteger,
	},
}

func multipleOfInteger(value any) (func(a any) *Error, error) {
	v, ok := value.(float64)
	if !ok {
		return nil, errors.New("multipleOf requires integer")
	}

	if float64(int(v)) != v {
		return nil, errors.New("multipleOf requires integer")
	}

	return func(a any) *Error {
		check := float64(a.(int)) / v
		if float64(int(check)) != check {
			return NewError(v, a)
		}

		return nil
	}, nil
}

func exclusiveMaximumInteger(max any) (func(a any) *Error, error) {
	v, ok := max.(float64)
	if !ok {
		return nil, errors.New("exclusiveMaximum requires integer")
	}

	if float64(int(v)) != v {
		return nil, errors.New("exclusiveMaximum requires integer")
	}

	return func(a any) *Error {
		if v > float64(a.(int)) {
			return nil
		}

		return NewError(float64(a.(int)), v)
	}, nil
}

func maximumInteger(max any) (func(a any) *Error, error) {
	v, ok := max.(float64)
	if !ok {
		return nil, errors.New("maximum requires integer")
	}

	if float64(int(v)) != v {
		return nil, errors.New("maximum requires integer")
	}

	return func(a any) *Error {
		if v >= float64(a.(int)) {
			return nil
		}

		return NewError(float64(a.(int)), v)
	}, nil
}

func exclusiveMinimumInteger(min any) (func(a any) *Error, error) {
	v, ok := min.(float64)
	if !ok {
		return nil, errors.New("minimum requires integer")
	}

	if float64(int(v)) != v {
		return nil, errors.New("minimum requires integer")
	}

	return func(a any) *Error {
		if v < float64(a.(int)) {
			return nil
		}

		return NewError(float64(a.(int)), v)
	}, nil
}

func minimumInteger(min any) (func(a any) *Error, error) {
	v, ok := min.(float64)
	if !ok {
		return nil, errors.New("minimum requires integer")
	}

	if float64(int(v)) != v {
		return nil, errors.New("minimum requires integer")
	}

	return func(a any) *Error {
		if v <= float64(a.(int)) {
			return nil
		}

		return NewError(float64(a.(int)), v)
	}, nil
}
