package license

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
)

const (
	EngineLicenseGeneratorPrefix = "nanotrix:"
)

type (
	// Input is encrypted engine query and output is encrypted response
	EngineLicenseGenerator interface {
		Generate() string
		Valid(license string, blockId uint32) error
	}

	EngineLicenseGenerateFn func() string

	DefaultEngineLicenseGenerator struct {
		nowFn timeutil.NowFn
	}
)

var _ = EngineLicenseGenerator(&DefaultEngineLicenseGenerator{})

type DefaultEngineLicenseGeneratorOpt func(d *DefaultEngineLicenseGenerator)

func WithNowFnDefaultEngineLicenseCacheOpt(fn timeutil.NowFn) DefaultEngineLicenseGeneratorOpt {
	return func(d *DefaultEngineLicenseGenerator) {
		d.nowFn = fn
	}
}

func NewDefaultEngineLicenseGenerator(opts ...DefaultEngineLicenseGeneratorOpt) *DefaultEngineLicenseGenerator {
	s := &DefaultEngineLicenseGenerator{
		nowFn: timeutil.NowFn(timeutil.NowUTC),
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

func (p *DefaultEngineLicenseGenerator) Generate() string {
	secret := make([]byte, 8)
	rand.Read(secret[:4])
	blockId := p.nowFn().Unix()
	binary.LittleEndian.PutUint32(secret[4:], uint32(blockId))

	hash := sha1.New()
	hash.Write(secret)

	sum := hash.Sum(nil)
	sum = append(sum, secret[:4]...)

	return fmt.Sprintf("%v%v", EngineLicenseGeneratorPrefix, hex.EncodeToString(sum))
}

func (p *DefaultEngineLicenseGenerator) Valid(lic string, blockId uint32) error {
	if !strings.HasPrefix(lic, EngineLicenseGeneratorPrefix) {
		return errors.New("invalid prefix")
	}

	lic = strings.TrimPrefix(lic, EngineLicenseGeneratorPrefix)

	if len(lic) < 48 {
		return errors.New("invalid license")
	}

	sumWithRandom, err := hex.DecodeString(lic)

	if err != nil {
		return err
	}

	secret := make([]byte, 8)

	copy(secret, sumWithRandom[20:])
	binary.LittleEndian.PutUint32(secret[4:], blockId)

	hash := sha1.New()
	if _, err = hash.Write(secret); err != nil {
		return err
	}

	expected := hash.Sum(nil)
	if bytes.Compare(expected, sumWithRandom[:20]) != 0 {
		return errors.New("not equal")
	}

	return nil
}
