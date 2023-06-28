package crypto

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/pkg/errors"
)

// Aes256CbcEncrypt encrypts the given plaintext with AES-256 in CBC mode.
func Aes256CbcEncrypt(plaintext, key []byte) ([]byte, error) {
	// Make sure key is valid length (256 bits)
	if len(key) != 32 {
		return nil, ErrInvalidKeyLength
	}
	// Initialize the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "aes cipher creation failed")
	}
	// Pad the plaintext so that its size is multiple of the default AES
	// block size
	plaintext, err = PKCS7Pad(plaintext, aes.BlockSize)
	if err != nil {
		return nil, errors.WithMessage(err, "padding failed")
	}
	// Generate a random initialization vector.
	iv, err := GenerateRandomBytes(aes.BlockSize)
	if err != nil {
		return nil, errors.WithMessage(err, "iv generation failed")
	}
	// CreateInvitation a CBC Encrypter
	mode := cipher.NewCBCEncrypter(block, iv)

	// Encrypt the plaintext
	ciphertext := plaintext // just reference plaintext; cryptBlocks will work in-place
	mode.CryptBlocks(plaintext, ciphertext)

	// Return the ciphertext with the initialization vector prepended
	return append(iv, ciphertext...), nil
}

// Aes256CbcDecrypt decrypts a message that was encrypted with AES-256-CBC.
func Aes256CbcDecrypt(ciphertext, key []byte) ([]byte, error) {
	// Make sure key is valid length (256 bits)
	if len(key) != 32 {
		return nil, ErrInvalidKeyLength
	}
	// Check if the ciphertext is smaller than AES's default blocksize.
	// We multiply by two because the IV will be prepended to the ciphertext
	if len(ciphertext) < aes.BlockSize*2 {
		return nil, ErrInvalidCiphertext
	}
	// Initialize the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "aes cipher creation failed")
	}
	// Split the initialization vector from the ciphertext
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	// Return an error if the ciphertext is not multiple of AES's blocksize
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, ErrInvalidCiphertext
	}
	// CreateInvitation a CBC Decrypter
	mode := cipher.NewCBCDecrypter(block, iv)

	// Decrypt the ciphertext
	plaintext := ciphertext // work in-place
	mode.CryptBlocks(plaintext, ciphertext)

	// Unpad the plaintext
	plaintext, err = PKCS7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return nil, errors.WithMessage(err, "pkcs7unpad failed")
	}
	// Return the plaintext
	return plaintext, nil
}
