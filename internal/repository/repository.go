package repository

import (
	"errors"
)

//Nullable interface
type Nullable interface {
	NullableValue() any
}

//Nullable Values
type NullableStr struct {
	Value string
	Valid bool
}

type NullableInt struct {
	Value int
	Valid bool
}

var NULLStr = NullableStr{"", false} //Default NULL string
var NULLInt = NullableInt{-1, false} //Default NULL int

var StringIsNullErr error = errors.New("string is null")
var IntIsNullErr error = errors.New("int is null")

//Get Nullable string with some value
func MakeNullableStr(s any) NullableStr {
	//Make sure s is a string
	str, ok := s.(string)
	if !ok {
		return NULLStr
	}

	//If str is empty, string is NULL
	if str == "" {
		return NULLStr
	}

	return NullableStr{str, true}
}

//Get Nullable int with some value
func MakeNullableInt(i any) NullableInt {
	//Make sure i is int
	switch v := i.(type) {
	case int:
		return NullableInt{v, true}
	case int64:
		return NullableInt{int(v), true}
	default:
		return NULLInt
	}
}

//Gets value. Can be NULL or nullable.Value
func (s NullableStr) NullableValue() any {
	if s.Valid {
		return s.Value
	} else {
		return nil
	}
}

func (i NullableInt) NullableValue() any {
	if i.Valid {
		return i.Value
	} else {
		return nil
	}
}
