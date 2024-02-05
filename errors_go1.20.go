//go:build !go1.21

package pgxgeos

import "errors"

var errUnsupported = errors.New("unsupported")
