package domain

import "errors"

// ErrNotFound — запись не найдена в хранилище.
var ErrNotFound = errors.New("entity not found")

// ErrSlotBusy — выбранный интервал уже занят другой арендой.
var ErrSlotBusy = errors.New("слот аренды занят")
