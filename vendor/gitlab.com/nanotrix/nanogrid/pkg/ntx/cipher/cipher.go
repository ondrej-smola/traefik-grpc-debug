package cipher

import (
	"crypto/sha512"

	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"io"

	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

type (
	Cipher interface {
		Encrypt(msg []byte) ([]byte, error)
		Decrypt(msg []byte) ([]byte, error)
	}

	FromAuthTokenFn func(tkb *auth.NtxToken) Cipher

	gcm struct {
		key []byte
	}

	gz struct {
		next Cipher
	}
)

func NewGCM(key, salt string) Cipher {
	keySum := sha512.Sum512([]byte(key))
	saltSum := sha512.Sum512([]byte(salt))
	return &gcm{key: CreateAES256Key(keySum[:], saltSum[:])}
}

func (g *gcm) Encrypt(msg []byte) ([]byte, error) {
	return EncryptGCM(g.key, msg)
}

func (g *gcm) Decrypt(msg []byte) ([]byte, error) {
	return DecryptGCM(g.key, msg)
}

func WithGzip(c Cipher) Cipher {
	return &gz{next: c}
}

func (g *gz) Encrypt(msg []byte) ([]byte, error) {
	if out, err := util.GzipEncode(msg); err != nil {
		return nil, err
	} else {
		return g.next.Encrypt(out)
	}
}

func (g *gz) Decrypt(msg []byte) ([]byte, error) {
	out, err := g.next.Decrypt(msg)
	if err != nil {
		return nil, err
	}

	return util.GzipDecode(out)
}

const (
	gcmNonceSize = 12
)

type GcmProcessor struct {
	block cipher.Block
}

var _ = Cipher(&GcmProcessor{})

func NewGcmProcessorForSecret(secret []byte) *GcmProcessor {
	sum := md5.Sum([]byte(secret))
	block, _ := aes.NewCipher(sum[:])
	return NewGcmProcessor(block)
}

func NewGcmProcessor(block cipher.Block) *GcmProcessor {
	return &GcmProcessor{block: block}
}

func (a *GcmProcessor) Decrypt(src []byte) ([]byte, error) {
	if len(src) < gcmNonceSize {
		return nil, errors.Errorf("token too short: got %v bytes", len(src))
	}

	nonce := src[:gcmNonceSize]
	gcm, err := cipher.NewGCM(a.block)

	if err != nil {
		return nil, errors.Wrap(err, "cipher init error")
	}

	if decrypted, err := gcm.Open(nil, nonce, src[gcmNonceSize:], nil); err != nil {
		return nil, errors.Wrap(err, "open")
	} else {
		return decrypted, nil
	}
}

func (a *GcmProcessor) Encrypt(src []byte) ([]byte, error) {
	res := make([]byte, gcmNonceSize+len(src))

	nonce := res[:gcmNonceSize]
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}

	gcm, err := cipher.NewGCM(a.block)
	if err != nil {
		return nil, errors.Wrap(err, "cipher init error")
	}
	res = gcm.Seal(nil, nonce, src, nil)
	return append(nonce, res...), nil
}
