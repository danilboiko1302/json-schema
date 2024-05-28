package jsonschema

import (
	"errors"
	"fmt"
	"reflect"
)

var sliceValidation map[string]rawValidation = map[string]rawValidation{
	"minItems": {
		function: minItems,
	},
	"maxItems": {
		function: maxItems,
	},
	"uniqueItems": {
		function: uniqueItems,
	},
	"contains": {
		function: contains,
	},
	"minContains": {
		function: minContains,
		requires: []string{"contains"},
	},
	"maxContains": {
		function: maxContains,
		requires: []string{"contains"},
	},
	"items": {
		function: items,
	},
}

// TODO: supports only slice
func items(value any) (func(a any) *Error, error) {
	slice, ok := value.([]any)
	if ok {
		return itemsSlice(slice)
	}

	v, ok := value.(map[string]any)
	if ok {
		return itemsMap(v)
	}

	return nil, errors.New("items supports map or slice only")
}

func itemsSlice(value []any) (func(a any) *Error, error) {
	var res []ValueType = make([]ValueType, 0, len(value))

	for _, value := range value {
		valueType, ok := getValueType(value.(map[string]any)["type"])
		if !ok {
			return nil, errors.New("maxContains requires valid type")
		}

		res = append(res, valueType)
	}

	return func(a any) *Error {
		for i, valueType := range res {
			err := validate(a.([]any)[i], &Schema{
				valueType: valueType,
			})
			if err != nil {
				return NewError(valueType, a.([]any)[i])
			}
		}

		return nil
	}, nil
}

func itemsMap(values map[string]any) (func(a any) *Error, error) {
	valueType, ok := getValueType(values["type"])
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
	case String:
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
	case Number:
		for name, validation := range numberValidation {
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
	}

	return func(a any) *Error {
		for _, target := range a.([]any) {
			err := validate(target, schema)
			if err != nil {
				return err
			}
		}

		return nil
	}, nil
}

func maxContains(value any) (func(a any) *Error, error) {
	if len(value.([]any)) < 2 {
		return nil, errors.New("maxContains requires 2 value")
	}

	v := value.([]any)

	checkType, ok := v[1].(map[string]any)
	if !ok {
		return nil, errors.New("maxContains requires type")
	}

	valueType, ok := getValueType(checkType["type"])
	if !ok {
		return nil, errors.New("maxContains requires valid type")
	}

	checkMax, ok := v[0].(float64)
	if !ok {
		return nil, errors.New("maxContains requires number")
	}

	return func(a any) *Error {
		var correct int

		for _, elem := range a.([]any) {
			err := validate(elem, &Schema{
				valueType: valueType,
			})
			if err == nil {
				correct++
			}
		}

		if correct <= int(checkMax) {
			return nil
		}

		return NewError(int(checkMax), correct)
	}, nil
}

func minContains(value any) (func(a any) *Error, error) {
	if len(value.([]any)) < 2 {
		return nil, errors.New("minContains requires 2 value")
	}

	v := value.([]any)

	checkType, ok := v[1].(map[string]any)
	if !ok {
		return nil, errors.New("minContains requires type")
	}

	valueType, ok := getValueType(checkType["type"])
	if !ok {
		return nil, errors.New("minContains requires valid type")
	}

	checkMin, ok := v[0].(float64)
	if !ok {
		return nil, errors.New("minContains requires number")
	}

	return func(a any) *Error {
		var correct int

		for _, elem := range a.([]any) {
			err := validate(elem, &Schema{
				valueType: valueType,
			})
			if err == nil {
				correct++
			}
		}

		if correct >= int(checkMin) {
			return nil
		}

		return NewError(int(checkMin), correct)
	}, nil
}

func contains(value any) (func(a any) *Error, error) {
	v, ok := value.(map[string]any)
	if !ok {
		return nil, errors.New("contains requires type")
	}

	valueType, ok := getValueType(v["type"])
	if !ok {
		return nil, errors.New("contains requires valid type")
	}

	return func(a any) *Error {
		types := make(map[string]struct{})
		for _, elem := range a.([]any) {
			types[reflect.ValueOf(elem).Kind().String()] = struct{}{}

			err := validate(elem, &Schema{
				valueType: valueType,
			})
			if err == nil {
				return nil
			}

		}

		keys := make([]string, 0, len(types))
		for vt := range types {
			keys = append(keys, vt)
		}

		return NewError(valueType, keys)
	}, nil
}

func uniqueItems(value any) (func(a any) *Error, error) {
	v, ok := value.(bool)
	if !ok {
		return nil, errors.New("uniqueItems requires *Errorean")
	}

	return func(a any) *Error {
		if !v {
			return nil
		}

		check := map[any]struct{}{}

		for _, elem := range a.([]any) {
			if _, ok := check[elem]; ok {
				return NewError("unique", fmt.Sprintf("elem: %v duplicated", elem))
			}

			check[elem] = struct{}{}
		}

		return nil
	}, nil
}

func maxItems(value any) (func(a any) *Error, error) {
	v, ok := value.(float64)
	if !ok {
		return nil, errors.New("maxItems requires integer")
	}

	if float64(int(v)) != v {
		return nil, errors.New("maxItems requires integer")
	}

	return func(a any) *Error {
		if len(a.([]any)) <= int(v) {
			return nil
		}

		return NewError(int(v), len(a.([]any)))
	}, nil
}

func minItems(value any) (func(a any) *Error, error) {
	v, ok := value.(float64)
	if !ok {
		return nil, errors.New("minItems requires integer")
	}

	if float64(int(v)) != v {
		return nil, errors.New("minItems requires integer")
	}

	return func(a any) *Error {
		if len(a.([]any)) >= int(v) {
			return nil
		}

		return NewError(int(v), len(a.([]any)))
	}, nil
}
