package repository

import (
	"errors"
)

//Nullable interface
type Nullable interface {
	IsNull() bool
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

//Get Nullable string with some value
func MakeNullableStr(s string) NullableStr {
	valid := true

	//If s is empty, string is NULL
	if s == "" {
		valid = false
	}

	return NullableStr{s, valid}
}

//Get Nullable int with some value
func MakeNullableInt(i int) NullableInt {
	return NullableInt{i, true}
}

//Checks if nullable is null
func (s NullableStr) IsNull() bool {
	return s.Valid
}

func (i NullableInt) IsNull() bool {
	return i.Valid
}
