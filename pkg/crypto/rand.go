package crypto

import (
	"crypto/rand"

	"github.com/pkg/errors"
)

// GenerateRandomBytes returns a bytes slice with size n that contains
// cryptographically secure random bytes.
func GenerateRandomBytes(n uint) ([]byte, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
		return nil, errors.Wrap(err, "error reading random bytes")
	}

	return b, nil
}
