package storage

import (
	"errors"
)

var (
	ErrAliasNotFound = errors.New("alias doesnt exist")
	ErrAliasExists   = errors.New("alias exists")
)
