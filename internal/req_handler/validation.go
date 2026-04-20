package req_handler

import (
	"fmt"
	"time"
)

/*
* This file holds types and methods for Validation purposes
 */
type DateField struct { Value any }
type RaceCategoryField struct { Value any }
type RaceStatusField struct { Value any }

type FieldEmptyError struct { }
type FieldWrongFormatError struct { }
type FieldInvalidInputError struct { }

var FieldIsEmptyErr error = FieldEmptyError{} //Default instantiation of error
var FieldIsWrongFormatErr error = FieldWrongFormatError{} //Default instantiation of error
var FieldIsInvalidErr error = FieldInvalidInputError{}

/*
* Errors
*/

func (e FieldEmptyError) Error() string {
	return fmt.Sprintln("Validation Error: Field is empty.")
}

func (e FieldWrongFormatError) Error() string {
	return fmt.Sprintln("Validation Error: Field is incorrectly formatted.")
}

func (e FieldInvalidInputError) Error() string {
	return fmt.Sprintln("Validation Error: Field is invalid value")
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

func (f *RaceCategoryField) Validate() error {
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

func (f *RaceStatusField) Validate() error {
	if f.Value == nil {
		return FieldIsEmptyErr
	}

	fStr, ok := f.Value.(string)
	if !ok {
		return FieldIsWrongFormatErr
	}
	
	if fStr == "" {
		return FieldIsEmptyErr
	}

	//Check valid category cache 
	if (fStr != "upcoming" && fStr != "in_progress" && fStr != "completed") {
		return FieldIsInvalidErr
	}

	return nil
}
