package req_handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

/*
* This file holds types and methods for Validation purposes
 */

//Field types
type DateField struct { Value any }
type TextField struct { Value any }
type TimeField struct { Value any }
type DurationField struct { Value any }
type IntField struct { Value any }
type ArrayField struct { Value any }

//Errors
type FieldEmptyError struct { }
type FieldWrongFormatError struct { }

//Default instantiation of errora
var FieldIsEmptyErr error = FieldEmptyError{}
var FieldIsWrongFormatErr error = FieldWrongFormatError{}

/*
* Errors
*/

func (e FieldEmptyError) Error() string {
	return fmt.Sprintln("Validation Error: Field is empty.")
}

func (e FieldWrongFormatError) Error() string {
	return fmt.Sprintln("Validation Error: Field is incorrectly formatted.")
}

/*
* f.Validate() error
* Checks if passed in value is the correct format.
*/

func (f *DateField) Validate() error {
	if f.Value == nil {
		return FieldIsEmptyErr
	} //If field is nil, return FieldEmptyError

	fStr, ok := f.Value.(string)
	if !ok {
		return FieldIsWrongFormatErr
	} //If field is not nil, but is the wrong datatype, return WrongFormatError

	if fStr == "" {
		return FieldIsEmptyErr
	} //Field is the right datatype, but the field is empty

	_, err := time.Parse(time.DateOnly, fStr) //Convert input to proper RFC3339
	if err != nil {
		return FieldIsWrongFormatErr
	} //If unable to parse data, this means data is wrong format

	return nil //Field is valid Date
}

func (f *TextField) Validate() error {
	if f.Value == nil {
		return FieldIsEmptyErr
	} //Field is empty

	fStr, ok := f.Value.(string)
	if !ok {
		return FieldIsWrongFormatErr
	}

	if fStr == "" {
		return FieldIsEmptyErr
	}

	return nil
}

func (f *TimeField) Validate() error {
	if f.Value == nil {
		return FieldIsEmptyErr
	} //If field is nil, return FieldEmptyError

	fStr, ok := f.Value.(string)
	if !ok {
		return FieldIsWrongFormatErr
	} //If field is not nil, but is the wrong datatype, return WrongFormatError

	if fStr == "" {
		return FieldIsEmptyErr
	} //Field is the right datatype, but the field is empty

	_, err := time.Parse(time.TimeOnly, fStr)
	if err != nil {
		return FieldIsWrongFormatErr
	} //If unable to parse data, this means data is wrong format

	return nil //Field is valid Time
}

func (f *DurationField) Validate() error {
	if f.Value == nil {
		return FieldIsEmptyErr
	} //If field is nil, return FieldEmptyError

	fStr, ok := f.Value.(string)
	if !ok {
		return FieldIsWrongFormatErr
	} //If field is not nil, but is the wrong datatype, return WrongFormatError

	if fStr == "" {
		return FieldIsEmptyErr
	} //Field is the right datatype, but the field is empty

	//Parse duration string
	parts := strings.Split(fStr, ":")
	if len(parts) != 3 {
		return FieldIsWrongFormatErr
	}
	for i, p := range parts {
		val, err := strconv.Atoi(p)
		if err != nil {
			return FieldIsWrongFormatErr
		}

		if i == 0 && val < 0 {
			return FieldIsWrongFormatErr
		} else if i > 0 && (val < 0 || val > 59) {
			return FieldIsWrongFormatErr
		}
	}

	return nil //Field is valid Duration
}

func (f *IntField) Validate() error {
	if f.Value == nil {
		return FieldIsEmptyErr
	} //Field is empty

	switch f.Value.(type) {
	case int, int64, float64:
		return nil
	default:
		return FieldIsWrongFormatErr
	}
}

func (f *ArrayField) Validate() error {
	if f.Value == nil {
		return FieldIsEmptyErr
	} //Field is empty

	arr, ok := f.Value.([]any)
	if !ok {
		return FieldIsWrongFormatErr
	}

	if len(arr) == 0 {
		return FieldIsEmptyErr
	}

	return nil
}