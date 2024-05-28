package jsonschema

import (
	"errors"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var stringValidation map[string]rawValidation = map[string]rawValidation{
	"minLength": {
		function: minLength,
	},
	"maxLength": {
		function: maxLength,
	},
	"pattern": {
		function: pattern,
	},
	"format": {
		function: format,
	},
}

func format(format any) (func(a any) *Error, error) {
	format, ok := format.(string)
	if !ok {
		return nil, errors.New("format requires string")
	}

	switch format {
	case "date-time":
		return func(a any) *Error {
			v := a.(string)
			_, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return NewError(format, a.(string))
			}
			return nil
		}, nil
	case "date":
		return func(a any) *Error {
			v := a.(string)
			_, err := time.Parse(time.DateOnly, v)
			if err != nil {
				return NewError(format, a.(string))
			}
			return nil
		}, nil
	case "time":
		return func(a any) *Error {
			v := a.(string)
			_, err := time.Parse(time.TimeOnly, v)
			if err != nil {
				return NewError(format, a.(string))
			}
			return nil
		}, nil
	case "duration":
		return func(a any) *Error {
			_, err := time.ParseDuration(a.(string))
			if err != nil {
				return NewError(format, a.(string))
			}
			return nil
		}, nil
	case "regex":
		return func(a any) *Error {
			_, err := regexp.Compile(a.(string))
			if err != nil {
				return NewError(format, a.(string))
			}
			return nil
		}, nil
	case "email":
		return func(a any) *Error {
			_, err := mail.ParseAddress(a.(string))
			if err != nil {
				return NewError(format, a.(string))
			}
			return nil
		}, nil
	case "hostname", "uri":
		return func(a any) *Error {
			_, err := url.Parse(a.(string))
			if err != nil {
				return NewError(format, a.(string))
			}
			return nil
		}, nil
	case "ipv4", "ipv6":
		return func(a any) *Error {
			ip := net.ParseIP(a.(string))
			if ip == nil {
				return NewError(format, a.(string))
			}
			return nil
		}, nil
	case "uuid":
		return func(a any) *Error {
			_, err := uuid.Parse(a.(string))
			if err != nil {
				return NewError(format, a.(string))
			}
			return nil
		}, nil
	}

	return nil, errors.New("unknown format for str")
}

func pattern(pattern any) (func(a any) *Error, error) {
	v, ok := pattern.(string)
	if !ok {
		return nil, errors.New("pattern requires string")
	}

	r, err := regexp.Compile(v)
	if err != nil {
		return nil, err
	}

	return func(a any) *Error {
		ok := r.MatchString(a.(string))
		if !ok {
			return NewError(r, a.(string))
		}
		return nil
	}, nil
}

func maxLength(max any) (func(a any) *Error, error) {
	v, ok := max.(float64)
	if !ok {
		return nil, errors.New("maxLength requires integer")
	}

	if float64(int(v)) != v {
		return nil, errors.New("maxLength requires integer")
	}

	return func(a any) *Error {
		if int(v) >= len(a.(string)) {
			return nil
		}

		return NewError(int(v), len(a.(string)))
	}, nil
}

func minLength(min any) (func(a any) *Error, error) {
	v, ok := min.(float64)
	if !ok {
		return nil, errors.New("minLength requires integer")
	}

	if float64(int(v)) != v {
		return nil, errors.New("minLength requires integer")
	}

	return func(a any) *Error {
		if int(v) <= len(a.(string)) {
			return nil
		}

		return NewError(int(v), len(a.(string)))
	}, nil
}
