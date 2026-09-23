package utils

import "reflect"

func NilInitializeStruct(ptr interface{}) {
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return // Must be a non-nil pointer to a struct
	}
	v = v.Elem()

	if v.Kind() != reflect.Struct {
		return // Must point to a struct
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		switch field.Kind() {
		case reflect.Pointer:
			field.Set(reflect.Zero(field.Type()))
		case reflect.Struct:
			// Recurse if embedded struct (not pointer to struct)
			if field.CanAddr() {
				NilInitializeStruct(field.Addr().Interface())
			}
		}
	}
}
