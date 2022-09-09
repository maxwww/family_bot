package units

import "errors"

var (
	ErrDuplicateID   = errors.New("duplicate ID")
	ErrNotFound      = errors.New("record not found")
	ErrInternal      = errors.New("internal error")
	DataParsingError = errors.New("data parsing error")
)
