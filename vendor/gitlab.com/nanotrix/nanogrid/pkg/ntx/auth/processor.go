package auth

import (
	"crypto"

	"github.com/dgrijalva/jwt-go"
	"github.com/mendsley/gojwk"
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

type (
	Decoder interface {
		// decode and validate token
		Decode(token string) (*NtxToken, error)
	}

	DecoderFn func(token string) (*NtxToken, error)

	Encoder interface {
		Encode(*NtxToken) (string, error)
	}

	EncoderFn func(*NtxToken) (string, error)

	Processor interface {
		Decoder
		Encoder
	}

	HS256Decoder struct {
		key []byte
	}

	HS256Encoder struct {
		key []byte
	}

	RS256Decoder struct {
		kidToPubKey map[string]crypto.PublicKey
	}

	RS256TokenEncoder struct {
		key crypto.PrivateKey
	}

	IssuerListTokenDecoder struct {
		issuerToKeyFunc map[string]jwt.Keyfunc
	}

	TokenWithIssuer interface {
		Issuer() string
	}

	processor struct {
		dec Decoder
		enc Encoder
	}
)

var _ = Decoder(&HS256Decoder{})
var _ = Decoder(&RS256Decoder{})
var _ = Decoder(&IssuerListTokenDecoder{})

var _ = Encoder(&HS256Encoder{})

func (f DecoderFn) Decode(token string) (*NtxToken, error) {
	return f(token)
}

func (f EncoderFn) Encode(tkn *NtxToken) (string, error) {
	return f(tkn)
}

func NewHS256TokenProcessorFromKey(key []byte) Processor {
	return NewTokenProcessor(NewHS256TokenEncoder(key), NewHS256TokenDecoder(key))
}

func NewTokenProcessor(enc Encoder, dec Decoder) Processor {
	return &processor{enc: enc, dec: dec}
}

func (p *processor) Decode(token string) (*NtxToken, error) {
	return p.dec.Decode(token)
}

func (p *processor) Encode(tkn *NtxToken) (string, error) {
	return p.enc.Encode(tkn)
}

func NewTokenDecoderForAuthConfigs(authCfgs ...*AuthProviderConfig) (Decoder, error) {
	issuerToParser := make(map[string]jwt.Keyfunc)

	for i, authCfg := range authCfgs {
		if alg := authCfg.Rs256; alg != nil {
			key := &gojwk.Key{}
			if err := util.ReadAllFromInputStringToJson(alg.Input, &key); err != nil {
				return nil, errors.Wrapf(err, "ReadPacket input string for config[%v]", i)
			}

			keySet := key.Keys

			if key.Alg != "" {
				keySet = append(keySet, key)
			}

			parser, err := NewRS256TokenDecoder(keySet...)
			if err != nil {
				return nil, errors.Wrap(err, "RS256 decoder")
			}
			issuerToParser[authCfg.Issuer] = parser.KeyFunc
		} else if alg := authCfg.Hs256; alg != nil {

			parser := NewHS256TokenDecoder([]byte(alg.Secret))
			issuerToParser[authCfg.Issuer] = parser.KeyFunc
		} else {
			return nil, errors.Errorf("Unsupported algorithm %T for config[%v]", alg, i)
		}
	}

	return &IssuerListTokenDecoder{issuerToKeyFunc: issuerToParser}, nil
}

func (d *IssuerListTokenDecoder) Decode(token string) (*NtxToken, error) {
	return DecodeNtxToken(token, d.KeyFunc)
}

func (d *IssuerListTokenDecoder) KeyFunc(token *jwt.Token) (interface{}, error) {
	tkn, ok := token.Claims.(TokenWithIssuer)
	if !ok {
		return nil, errors.Errorf("Claims does not have Issuer() function, got %T", token.Claims)
	}

	if kf, ok := d.issuerToKeyFunc[tkn.Issuer()]; !ok {
		return nil, errors.Errorf("Get parser for issuer '%v' not found", tkn.Issuer())
	} else {
		return kf(token)
	}
}

func NewHS256TokenDecoder(key []byte) *HS256Decoder {
	keyCopy := make([]byte, len(key))

	copy(keyCopy, key)
	return &HS256Decoder{key: keyCopy}
}

func (d *HS256Decoder) Decode(token string) (*NtxToken, error) {
	return DecodeNtxToken(token, d.KeyFunc)
}

func (d *HS256Decoder) KeyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, errors.Errorf("Unexpected signing method: %v, expected HMAC", token.Header["alg"])
	}
	return d.key, nil
}

func NewRS256TokenDecoder(keys ...*gojwk.Key) (*RS256Decoder, error) {
	kidToKey := make(map[string]crypto.PublicKey)

	for i, k := range keys {
		pk, err := k.DecodePublicKey()
		if err != nil {
			return nil, errors.Wrapf(err, "Key[%v]", i)
		}

		kidToKey[k.Kid] = pk
	}

	return &RS256Decoder{kidToPubKey: kidToKey}, nil
}

func (d *RS256Decoder) Decode(token string) (*NtxToken, error) {
	return DecodeNtxToken(token, d.KeyFunc)
}

func (d *RS256Decoder) KeyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, errors.Errorf("Unexpected signing method: %v, expected RSA", token.Header["alg"])
	}

	kid := ""

	kidH := token.Header["kid"]
	switch k := kidH.(type) {
	case string:
		kid = k
	}

	if pk, ok := d.kidToPubKey[kid]; !ok {
		return nil, errors.Errorf("Publci key for kid '%v' not found", kid)
	} else {
		return pk, nil
	}
}

func NewHS256TokenEncoder(key []byte) *HS256Encoder {
	keyCopy := make([]byte, len(key))

	copy(keyCopy, key)
	return &HS256Encoder{key: keyCopy}
}

func NewRS256TokenEncoder(key crypto.PrivateKey) *RS256TokenEncoder {
	return &RS256TokenEncoder{key: key}
}

func (e *HS256Encoder) Encode(token *NtxToken) (string, error) {
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, token)
	return tkn.SignedString(e.key)
}

func (e *RS256TokenEncoder) Encode(token *NtxToken) (string, error) {
	tkn := jwt.NewWithClaims(jwt.SigningMethodRS256, token)
	return tkn.SignedString(e.key)
}

func DecodeNtxToken(token string, fn jwt.Keyfunc) (*NtxToken, error) {
	claims := &NtxToken{}
	tkn, err := jwt.ParseWithClaims(token, claims, fn)

	if err != nil {
		return nil, errors.Wrap(err, "JWT: parse")
	}

	if !tkn.Valid {
		return nil, errors.New("JWT: token not valid")
	}

	return claims, errors.Wrap(claims.Valid(), "JWT: invalid claims")
}
