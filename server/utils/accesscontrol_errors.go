package utils

import "errors"

var (
	ErrAccessDenied = errors.New("access denied")
	ErrNotMember    = errors.New("user is not a project member")
)
