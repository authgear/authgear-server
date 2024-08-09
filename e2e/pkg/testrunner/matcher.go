package testrunner

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type MatchViolation struct {
	Path     string `json:"path"`
	Message  string `json:"message"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
}

/*
*

	 MatchJSON compares a json string to a schema json string, for example
	 {
			"id": "[[number]]",
			"title": "[[string]]",
			"publish": "[[boolean]]",
			"type": "articles",
			"tags": ["[[arrayof]]", "[[object]]"],
			"error": "[[null]]",
			"ignoreme": "[[ignore]]",
	 }
	 *
*/
func MatchJSON(jsonStr string, schema string) ([]MatchViolation, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
	}

	var schemaData interface{}
	if err := json.Unmarshal([]byte(schema), &schemaData); err != nil {
		return nil, err
	}

	violations := []MatchViolation{}
	violations = append(violations, matchValue("", data, schemaData)...)

	return violations, nil
}

func matchMap(path string, data, schema map[string]interface{}) (violations []MatchViolation) {
	for key, schemaValue := range schema {
		if key == "[[...rest]]" {
			continue
		}

		if schemaValue == "[[ignore]]" {
			continue
		}

		dataValue, ok := data[key]
		if schemaValue == "[[never]]" {
			if ok {
				violations = append(violations, typeMismatchViolation(path+"/"+key, "[[never]]", dataValue))
			}
			continue
		}
		if !ok {
			violations = append(violations, missingFieldViolation(path+"/"+key, schemaValue))
			continue
		}

		violations = append(violations, matchValue(path+"/"+key, dataValue, schemaValue)...)
	}

	// Check for extra fields
	restSchemaValue, ok := schema["[[...rest]]"]
	if !ok {
		// Allow extra fields by default
		restSchemaValue = "[[ignore]]"
	}
	for dataKey := range data {
		if _, ok := schema[dataKey]; !ok {
			violations = append(violations, matchValue(path+"/"+dataKey, data[dataKey], restSchemaValue)...)
		}
	}

	return violations
}

func matchValue(path string, dataValue, schemaValue interface{}) (violations []MatchViolation) {
	switch schemaValue := schemaValue.(type) {
	case string:
		violations = append(violations, matchScalar(path, dataValue, schemaValue)...)
	case map[string]interface{}:
		violations = append(violations, matchMap(path, dataValue.(map[string]interface{}), schemaValue)...)
	case []interface{}:
		violations = append(violations, matchArray(path, dataValue.([]interface{}), schemaValue)...)
	default:
		violations = append(violations, matchConstant(path, dataValue, schemaValue)...)
	}

	return violations
}

func matchArray(path string, data []interface{}, schemaValue interface{}) (violations []MatchViolation) {
	schemaArray, ok := schemaValue.([]interface{})
	if !ok {
		return []MatchViolation{unknownSchemaTypeViolation(path, schemaValue)}
	}

	for i, item := range data {
		var itemType interface{}
		if schemaArray[0] == "[[arrayof]]" {
			itemType = schemaArray[1]
		} else {
			itemType = schemaArray[i]
		}

		violations = append(violations, matchValue(fmt.Sprintf("%s/%d", path, i), item, itemType)...)
	}

	if len(data) != len(schemaArray) && schemaArray[0] != "[[arrayof]]" {
		for i := len(schemaArray); i < len(data); i++ {
			violations = append(violations, matchValue(fmt.Sprintf("%s/%d", path, i), data[i], schemaArray[len(schemaArray)-1])...)
		}
		for i := len(data); i < len(schemaArray); i++ {
			violations = append(violations, missingFieldViolation(fmt.Sprintf("%s/%d", path, i), schemaArray[i]))
		}
	}

	return violations
}

func matchScalar(path string, data interface{}, schema string) (violations []MatchViolation) {
	ok := false

	switch schema {
	case "[[number]]":
		ok = reflect.TypeOf(data).Kind() == reflect.Float64
	case "[[string]]":
		ok = reflect.TypeOf(data).Kind() == reflect.String
	case "[[boolean]]":
		ok = reflect.TypeOf(data).Kind() == reflect.Bool
	case "[[object]]":
		ok = reflect.TypeOf(data).Kind() == reflect.Map
	case "[[array]]":
		ok = reflect.TypeOf(data).Kind() == reflect.Slice
	case "[[null]]":
		ok = data == nil
	case "[[ignore]]":
		ok = true
	case "[[never]]":
		ok = false
	default:
		// Normal string
		ok = true
		violations = append(violations, matchConstant(path, data, schema)...)
	}

	if !ok {
		violations = append(violations, typeMismatchViolation(path, schema, data))
	}

	return violations
}

func matchConstant(path string, data interface{}, schema interface{}) (violations []MatchViolation) {
	if !reflect.DeepEqual(data, schema) {
		violations = append(violations, valueMismatchViolation(path, schema, data))
	}

	return violations
}

func unknownSchemaTypeViolation(path string, schema interface{}) MatchViolation {
	return MatchViolation{
		Path:    path,
		Message: "unknown schema type " + fmt.Sprintf("%v", schema),
	}
}

func missingFieldViolation(path string, expected interface{}) MatchViolation {
	return MatchViolation{
		Path:     path,
		Message:  "missing field",
		Expected: fmt.Sprintf("%v", expected),
		Actual:   "<missing>",
	}
}

func typeMismatchViolation(path string, expected interface{}, actual interface{}) MatchViolation {
	var guess = actual

	if actual == nil {
		guess = "[[null]]"
	} else {
		switch reflect.TypeOf(actual).Kind() {
		case reflect.Map:
			guess = "[[object]]"
		case reflect.Slice:
			guess = "[[array]]"
		case reflect.Float64:
			guess = "[[number]]"
		case reflect.String:
			guess = "[[string]]"
		case reflect.Bool:
			guess = "[[boolean]]"
		case reflect.Invalid:
			if guess == nil {
				guess = "[[null]]"
			}
		default:
			guess = fmt.Sprintf("%T", actual)
		}
	}

	return MatchViolation{
		Path:     path,
		Message:  "type mismatch",
		Expected: fmt.Sprintf("%v", expected),
		Actual:   guess.(string),
	}
}

func valueMismatchViolation(path string, expected interface{}, actual interface{}) MatchViolation {
	return MatchViolation{
		Path:     path,
		Message:  "value mismatch",
		Expected: fmt.Sprintf("%v", expected),
		Actual:   fmt.Sprintf("%v", actual),
	}
}
