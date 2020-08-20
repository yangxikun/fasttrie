package fasttrie

import (
	"fmt"
)

const (
	errSetValue         = "a value is already registered for path '%s'"
	errSetWildcardValue = "a wildcard value is already registered for path '%s'"
	errWildPathConflict = "'%s' in new path '%s' conflicts with existing wild path '%s' in existing prefix '%s'"
	errWildcardConflict = "'%s' in new path '%s' conflicts with existing wildcard '%s' in existing prefix '%s'"
	errWildcardSlash    = "no / before wildcard in path '%s'"
	errWildcardNotAtEnd = "wildcard routes are only allowed at the end of the path in path '%s'"
)

type radixError struct {
	msg    string
	params []interface{}
}

func (err *radixError) Error() string {
	return fmt.Sprintf(err.msg, err.params...)
}

func newRadixError(msg string, params ...interface{}) *radixError {
	return &radixError{msg, params}
}
