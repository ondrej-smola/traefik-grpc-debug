package license

import (
	"encoding/binary"
	"fmt"

	"encoding/base64"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/cipher"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

func clusterToKey(c *Cluster) string {
	return fmt.Sprintf("%v+%v++%v", c.Key, c.Id, c.Hostname)
}

func NewCipherForCluster(c *Cluster) cipher.Cipher {
	return cipher.WithGzip(cipher.NewGCM(clusterToKey(c), fmt.Sprint(c.Created)))
}

func NewKeyForCluster(c *Cluster) []byte {
	key := clusterToKey(c)
	salt := make([]byte, 8)
	binary.BigEndian.PutUint64(salt, c.Created)
	return cipher.CreateAES256Key([]byte(key), salt)
}

func TokenProcessorForCluster(cl *Cluster) auth.Processor {
	key := NewKeyForCluster(cl)
	return auth.NewTokenProcessor(auth.NewHS256TokenEncoder(key), auth.NewHS256TokenDecoder(key))
}

func ClusterToAuthProviderConfig(c *Cluster) *auth.AuthProviderConfig {
	key := NewKeyForCluster(c)

	return &auth.AuthProviderConfig{
		Issuer: c.Id,
		Hs256: &auth.AuthProviderConfig_HS256{
			Secret: string(key),
		},
	}
}

func DecodeLicense(token string, cl *Cluster) (*License, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, errors.Errorf("unexpected number of parts: %v", len(parts))
	}

	key := NewKeyForCluster(cl)
	jwt.SigningMethodHS512.Verify(parts[0], parts[1], key)

	lic := &License{}

	gzip, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}

	licBytes, err := util.GzipDecode(gzip)
	if err != nil {
		return nil, err
	}

	if err := util.ParseJsonPb(licBytes, lic); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	} else if err := lic.Valid(); err != nil {
		return nil, errors.Wrap(err, "license")
	}

	return lic, errors.Wrap(err, "license")
}

func EncodeLicense(lic *License) (string, error) {
	licBytes := util.MustJsonPb(lic)
	gzipBytes, err := util.GzipEncode(licBytes)
	if err != nil {
		return "", err
	}

	part1 := base64.StdEncoding.EncodeToString(gzipBytes)
	key := NewKeyForCluster(lic.Cluster)
	part2, err := jwt.SigningMethodHS512.Sign(part1, key)
	if err != nil {
		return "", err
	}

	return part1 + "." + part2, nil
}
