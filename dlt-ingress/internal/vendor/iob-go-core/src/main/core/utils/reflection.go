package utils

import "reflect"

func GetType(i interface{}) reflect.Type {
	t := reflect.TypeOf(i)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func GetFieldValue(i interface{}, fieldName string) interface{} {
	v := reflect.ValueOf(i)
	v = reflect.Indirect(v)

	if v.Kind() != reflect.Struct {
		return nil
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return nil
	}
	return field.Interface()
}
