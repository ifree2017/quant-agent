package model

import "github.com/google/uuid"

// newID 生成新UUID
func newID() string {
	return uuid.New().String()
}
