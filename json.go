package jsonschema

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"reflect"

	"github.com/go-resty/resty/v2"
)

type ValueType string

const (
	String  ValueType = "string"
	Integer ValueType = "integer"
	Number  ValueType = "number"
	Object  ValueType = "object"
	Array   ValueType = "array"
	Boolean ValueType = "boolean"
	Null    ValueType = "null"
)

func validateValueType(value ValueType) bool {
	switch value {
	case String, Integer, Number, Object, Array, Boolean, Null:
		return true
	default:
		return false
	}
}

type Schema struct {
	valueType    ValueType
	validateFunc map[string]validateFunc
	properties   map[string]Schema
}

type validateFunc func(any) *Error
type rawValidateFunc func(any) (func(any) *Error, error)

func Validate(target any, schema any) error {
	validatedTarget, err := validateTarget(reflect.ValueOf(target))
	if err != nil {
		return err
	}

	validatedSchema, err := validateSchema(reflect.ValueOf(schema))
	if err != nil {
		return err
	}

	err = validate(validatedTarget, validatedSchema)
	if err != (*Error)(nil) {
		return err
	}

	return nil
}

func validate(target any, schema *Schema) *Error {
	switch schema.valueType {
	case String:
		if _, ok := target.(string); !ok {
			return NewError("string", reflect.ValueOf(target).Kind().String())
		}

		return validateValue(target.(string), schema)
	case Integer:
		if _, ok := target.(float64); !ok {
			return NewError("int", reflect.ValueOf(target).Kind().String())
		}

		v := target.(float64)

		if float64(int(v)) != v {
			return NewError("int", "float64")
		}

		return validateValue(int(v), schema)
	case Number:
		if _, ok := target.(float64); !ok {
			return NewError("float64", reflect.ValueOf(target).Kind().String())
		}

		return validateValue(target.(float64), schema)
	case Boolean:
		if _, ok := target.(bool); !ok {
			return NewError("bool", reflect.ValueOf(target).Kind().String())
		}

		return validateValue(target.(bool), schema)
	case Null:
		if target != nil {
			return NewError("null", reflect.ValueOf(target).Kind().String())
		}

		return nil
	case Array:
		if _, ok := target.([]any); !ok {
			return NewError("slice", reflect.ValueOf(target).Kind().String())
		}

		return validateValue(target.([]any), schema)
	case Object:
		if _, ok := target.(map[string]any); !ok {
			return NewError("object", reflect.ValueOf(target).Kind().String())
		}

		return validateValue(target, schema)
	}
	return nil
}

func validateValue(target any, schema *Schema) *Error {
	for name, value := range schema.validateFunc {
		err := value(target)
		if err != nil {
			return err.SetName(name)
		}
	}

	return nil
}

func validateSchema(value reflect.Value) (*Schema, error) {
	switch value.Kind() {
	//json/path/url
	case reflect.String:
		return schemaFromString(value.String())
	//json bytes
	case reflect.Slice:
		bytes, ok := value.Interface().([]byte)
		if !ok {
			return nil, errors.New("unknown target, supports only slice of bytes")
		}

		return schemaFromString(string(bytes))
	case reflect.Map:
		schema, ok := value.Interface().(Schema)
		if !ok {
			return nil, errors.New("unknown schema, wrong map")
		}

		return &schema, nil
	case reflect.Pointer:
		schema, ok := value.Interface().(*Schema)
		if !ok {
			return validateSchema(value.Elem())
		}

		return schema, nil
	default:
		return nil, errors.New("unknown schema")
	}
}

func schemaFromString(str string) (*Schema, error) {
	data, err := JSONFromString(str)
	if err != nil {
		return nil, err
	}

	var values map[string]interface{} = make(map[string]interface{})

	err = json.Unmarshal([]byte(data), &values)
	if err != nil {
		return nil, err
	}

	return createSchemaFromJSON(values)
}

func createSchemaFromJSON(values map[string]interface{}) (*Schema, error) {
	var res *Schema = &Schema{}

	valueType, ok := getValueType(values["type"])
	if !ok {
		return nil, errors.New("schema has wrong type")
	}

	res.valueType = valueType

	validation, err := getValidation(valueType, values)
	if err != nil {
		return nil, err
	}

	res.validateFunc = validation

	if valueType == Object {
		raw, ok := values["properties"]
		//can be empty - skip
		if !ok {
			return res, nil
		}

		properties, ok := raw.(map[string]interface{})
		if !ok {
			return nil, errors.New("schema has wrong properties")
		}

		for key, value := range properties {
			prop, ok := value.(map[string]interface{})
			if !ok {
				continue
			}

			newSchema, err := createSchemaFromJSON(prop)
			if err != nil {
				return nil, err
			}

			if res.properties == nil {
				res.properties = make(map[string]Schema)
			}

			res.properties[key] = *newSchema
		}
	}

	return res, nil
}

func getValidation(valueType ValueType, values map[string]interface{}) (map[string]validateFunc, error) {
	validation, ok := validations[valueType]
	if !ok {
		return nil, nil
	}

	var res map[string]validateFunc = make(map[string]validateFunc)

	for name, validation := range validation {
		value, ok := values[name]
		if !ok {
			continue
		}

		if len(validation.requires) == 0 {
			validate, err := validation.function(value)
			if err != nil {
				return nil, err
			}

			res[name] = validate
		} else {
			var input []any = []any{value}

			for _, requires := range validation.requires {
				value, ok = values[requires]
				if !ok {
					return nil, errors.New(name + " requires " + requires)
				}

				input = append(input, value)
			}

			validate, err := validation.function(input)
			if err != nil {
				return nil, err
			}

			res[name] = validate
		}

	}

	return res, nil
}

type rawValidation struct {
	function rawValidateFunc
	requires []string
}

var validations map[ValueType]map[string]rawValidation = map[ValueType]map[string]rawValidation{
	String:  stringValidation,
	Integer: integerValidation,
	Number:  numberValidation,
	Array:   sliceValidation,
	Object:  objectValidation,
}

func getValueType(value any) (ValueType, bool) {
	if value == nil {
		return "", false
	}

	valueType, ok := value.(string)
	if !ok {
		return "", false
	}

	if !validateValueType(ValueType(valueType)) {
		return "", false
	}

	return ValueType(valueType), true
}

func validateTarget(target reflect.Value) (any, error) {
	switch target.Kind() {
	//json/path/url
	case reflect.String:
		return targetFromString(target.String())
	//json bytes
	case reflect.Slice:
		bytes, ok := target.Interface().([]byte)
		if !ok {
			return nil, errors.New("unknown target, supports only slice of bytes")
		}

		return targetFromString(string(bytes))
	case reflect.Struct, reflect.Pointer, reflect.Map:
		data, err := json.Marshal(target.Interface())
		if err != nil {
			return nil, err
		}

		res := map[string]any{}
		err = json.Unmarshal(data, &res)
		if err != nil {
			return nil, err
		}

		return res, nil
	default:
		return nil, errors.New("unknown target")
	}
}

func targetFromString(str string) (any, error) {
	data, err := JSONFromString(str)
	if err != nil {
		return nil, err
	}

	var res any
	err = json.Unmarshal([]byte(data), &res)
	if err != nil {
		return nil, err
	}

	return res, err
}

func JSONFromString(str string) (string, error) {
	if _, err := url.ParseRequestURI(str); err == nil {
		return getJSONFromUrl(str)
	}

	f, err := os.Open(str)
	if err == nil {
		f.Close()

		return getJSONFromFile(str)
	}

	return str, nil
}

func getJSONFromUrl(url string) (string, error) {
	response, err := resty.New().R().EnableTrace().Get(url)
	if err != nil {
		return "", err
	}

	if response.IsError() {
		return "", errors.New(string(response.Body()))
	}

	return string(response.Body()), err
}

func getJSONFromFile(str string) (string, error) {
	dat, err := os.ReadFile(str)
	if err != nil {
		return "", err
	}

	return string(dat), nil
}
