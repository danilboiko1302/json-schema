package jsonschema

import (
	"errors"
	"regexp"
)

var objectValidation map[string]rawValidation = map[string]rawValidation{
	"properties": {
		function: properties,
	},
	"required": {
		function: required,
	},
	"dependentRequired": {
		function: dependentRequired,
	},
	"minProperties": {
		function: minProperties,
	},
	"maxProperties": {
		function: maxProperties,
	},
	"propertyNames": {
		function: propertyNames,
	},
	"patternProperties": {
		function: patternProperties,
	},
}

func properties(value any) (func(a any) *Error, error) {
	props := make(map[string]ValueType)

	for name, value := range value.(map[string]any) {
		valueType, ok := getValueType(value.(map[string]any)["type"])
		if !ok {
			return nil, errors.New("properties requires valid type")
		}
		props[name] = valueType
	}

	return func(a any) *Error {
		v := a.(map[string]any)
		for name, value := range v {
			err := validate(value, &Schema{
				valueType: props[name],
			})
			if err != nil {
				return err
			}
		}
		return nil
	}, nil
}

func required(value any) (func(a any) *Error, error) {
	return func(a any) *Error {
		values := a.(map[string]any)
		for _, name := range value.([]any) {
			_, ok := values[name.(string)]
			if !ok {
				return NewError(name.(string), "")
			}
		}
		return nil
	}, nil
}

func dependentRequired(value any) (func(a any) *Error, error) {
	props := make(map[string][]string)

	for name, values := range value.(map[string]any) {
		temp := make([]string, 0)

		for _, value := range values.([]any) {
			temp = append(temp, value.(string))
		}
		props[name] = temp
	}

	return func(a any) *Error {
		v := a.(map[string]any)

		for name, values := range props {
			if _, ok := v[name]; !ok {
				continue
			}

			for _, value := range values {
				if _, ok := v[value]; !ok {
					return NewError(value, "")
				}
			}
		}

		return nil
	}, nil
}

func minProperties(value any) (func(a any) *Error, error) {
	v, ok := value.(float64)
	if !ok {
		return nil, errors.New("minProperties requires integer")
	}

	if float64(int(v)) != v {
		return nil, errors.New("minProperties requires integer")
	}

	return func(a any) *Error {
		values := a.(map[string]any)
		if len(values) >= int(v) {
			return nil
		}

		return NewError(int(v), len(values))
	}, nil
}

func maxProperties(value any) (func(a any) *Error, error) {
	v, ok := value.(float64)
	if !ok {
		return nil, errors.New("maxProperties requires integer")
	}

	if float64(int(v)) != v {
		return nil, errors.New("maxProperties requires integer")
	}

	return func(a any) *Error {
		values := a.(map[string]any)

		if len(values) <= int(v) {
			return nil
		}

		return NewError(int(v), len(values))
	}, nil
}

func propertyNames(value any) (func(a any) *Error, error) {
	values := value.(map[string]any)

	var schema *Schema = &Schema{
		valueType:    String,
		validateFunc: make(map[string]validateFunc),
	}

	for name, validation := range stringValidation {
		value, ok := values[name]
		if !ok {
			continue
		}

		validate, err := validation.function(value)
		if err != nil {
			return nil, err
		}

		schema.validateFunc[name] = validate
	}

	return func(a any) *Error {
		for name := range a.(map[string]any) {
			err := validate(name, schema)
			if err != nil {
				return err
			}
		}

		return nil
	}, nil
}

func patternProperties(value any) (func(a any) *Error, error) {
	values := value.(map[string]any)
	schemas := make(map[*regexp.Regexp]*Schema)

	for patter, values := range values {
		v := values.(map[string]any)

		valueType, ok := getValueType(v["type"])
		if !ok {
			return nil, errors.New("maxContains requires valid type")
		}

		var schema *Schema = &Schema{
			valueType:    valueType,
			validateFunc: make(map[string]validateFunc),
		}

		switch valueType {
		case Integer:
			for name, validation := range integerValidation {
				value, ok := v[name]
				if !ok {
					continue
				}

				validate, err := validation.function(value)
				if err != nil {
					return nil, err
				}

				schema.validateFunc[name] = validate
			}
		case String:
			for name, validation := range stringValidation {
				value, ok := v[name]
				if !ok {
					continue
				}

				validate, err := validation.function(value)
				if err != nil {
					return nil, err
				}

				schema.validateFunc[name] = validate
			}
		case Number:
			for name, validation := range numberValidation {
				value, ok := v[name]
				if !ok {
					continue
				}

				validate, err := validation.function(value)
				if err != nil {
					return nil, err
				}

				schema.validateFunc[name] = validate
			}
		}

		r, err := regexp.Compile(patter)
		if err != nil {
			return nil, err
		}

		schemas[r] = schema
	}

	return func(a any) *Error {
		for name, target := range a.(map[string]any) {
			for regex, schema := range schemas {
				if regex.MatchString(name) {
					err := validate(target, schema)
					if err != nil {
						return err
					}
				}
			}

		}

		return nil
	}, nil
}
