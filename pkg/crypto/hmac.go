package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
)

// HmacSha256 generates a hash-based message authentication code for
// the given data.
func HmacSha256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)

	// mac.Write() never returns an error
	// Ref: https://golang.org/pkg/hash/#Hash
	mac.Write(data)

	return mac.Sum(nil)
}

// VerifyHmacSha256 reports whether the given hash is valid.
func VerifyHmacSha256(key, givenMAC, data []byte) bool {
	expectedMAC := HmacSha256(key, data)

	return hmac.Equal(givenMAC, expectedMAC)
}
