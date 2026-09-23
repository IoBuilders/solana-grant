package utils

import "github.com/google/uuid"

// Helper function to check if a value exists in a list
func Contains[E string | int | int8 | int16 | int64 | uint | uint8 | uint16 | uint32 | uint64 | rune | float32 | float64 | bool | uuid.UUID](list []E, value E) (E, bool) {
	for _, item := range list {
		if item == value {
			return item, true
		}
	}
	var zeroValue E
	return zeroValue, false
}

// Transform each element using fn and return a new slice.
func MapSlice[In any, Out any](items []In, fn func(In) Out) []Out {
	if items == nil {
		return nil
	}
	out := make([]Out, len(items))
	for i := range items {
		out[i] = fn(items[i])
	}
	return out
}

// MapSlice function, but simplifies pointer handling uses on fn input
func MapSlicePtr[In any, Out any](items []In, fn func(*In) Out) []Out {
	if items == nil {
		return nil
	}
	out := make([]Out, len(items))
	for i := range items {
		out[i] = fn(&items[i])
	}
	return out
}

// MapSlicePtr function, but simplifies pointer handling uses on fon output
func MapSlicePtrValue[In any, Out any](items []In, fn func(*In) *Out) []Out {
	if items == nil {
		return nil
	}
	out := make([]Out, len(items))
	for i := range items {
		res := fn(&items[i])
		if res != nil {
			out[i] = *res
		}
	}
	return out
}
