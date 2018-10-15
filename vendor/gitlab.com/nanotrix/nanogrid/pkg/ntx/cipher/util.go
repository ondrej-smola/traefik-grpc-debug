package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"crypto/sha512"

	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
)

const Pbkdf2Iterations = 10123

func CreateAES256Key(secret, salt []byte) []byte {
	return CreateKey(secret, salt, 32)
}

func CreateKey(secret, salt []byte, keySize int) []byte {
	return pbkdf2.Key(secret, salt, Pbkdf2Iterations, keySize, sha512.New)
}

func EncryptGCM(key, message []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	nonce, err := GenerateNonce(gcm.NonceSize())
	if err != nil {
		return nil, err
	}

	var buf []byte
	buf = append(buf, nonce...)
	buf = gcm.Seal(buf, nonce, message, nil)
	return buf, nil
}

func DecryptGCM(key, message []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	if len(message) <= gcm.NonceSize() {
		return nil, errors.Errorf("message too short")
	}

	nonce := make([]byte, gcm.NonceSize())
	copy(nonce, message)

	out, err := gcm.Open(nil, nonce, message[gcm.NonceSize():], nil)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GenerateNonce creates a new random nonce.
func GenerateNonce(size int) ([]byte, error) {
	nonce := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}
	return nonce, nil
}
