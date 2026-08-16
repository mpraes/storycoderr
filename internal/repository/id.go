package repository

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// NewID returns a 32-character hex identifier.
//
//	id, err := repository.NewID()
func NewID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("cannot generate id, expected 16 random bytes: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}
