package data

import "errors"

var (
	ErrFileNotFound = errors.New("data file not found")
	ErrUnknownMarket = errors.New("unknown market type")
	ErrInvalidFormat = errors.New("invalid data format")
)
