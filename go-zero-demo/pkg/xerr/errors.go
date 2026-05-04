package xerr

import "errors"

var (
    ErrInvalidParam = errors.New("invalid param")
    ErrNotFound     = errors.New("resource not found")
)

