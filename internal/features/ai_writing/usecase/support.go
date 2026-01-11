package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

type randomIDGenerator struct{}

func (randomIDGenerator) NewID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
