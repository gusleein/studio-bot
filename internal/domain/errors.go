package domain

import "errors"

// ErrNotFound — запись не найдена в хранилище.
var ErrNotFound = errors.New("entity not found")
