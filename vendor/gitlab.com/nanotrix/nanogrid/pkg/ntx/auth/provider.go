package auth

import (
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

const (
	authProviderPartsSeparator = "|"
	ProvidersSeparator         = " "
)

func ParseProviderConfigs(in string) (ProviderConfigs, error) {
	var cfgs ProviderConfigs

	parts := strings.Split(in, ProvidersSeparator)

	for i, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			if cfg, err := ParseProviderConfig(s); err != nil {
				return nil, errors.Wrapf(err, "parse config[%v]", i)
			} else {
				cfgs = append(cfgs, cfg)
			}
		}
	}

	return cfgs, errors.Wrap(cfgs.Valid(), "auth-configs")
}

/// alg|issuer|secret or jwks as input string
func ParseProviderConfig(in string) (*AuthProviderConfig, error) {
	parts := strings.Split(in, authProviderPartsSeparator)

	if len(parts) != 3 {
		return nil, errors.Errorf("unexpected number of parts, expected 3 got %v", len(parts))
	}

	alg := strings.ToLower(parts[0])
	secretInputString := parts[2]

	cfg := &AuthProviderConfig{Issuer: parts[1]}

	switch alg {
	case "hs256":
		secret, err := util.ReadStringFromInputString(secretInputString)
		if err != nil {
			return nil, errors.Wrapf(err, "read secret input string: %v", secretInputString)
		}

		cfg.Hs256 = &AuthProviderConfig_HS256{Secret: secret}
	case "rs256":
		cfg.Rs256 = &AuthProviderConfig_RS256{Input: secretInputString}
	default:
		return nil, errors.Errorf("unsupported alg: '%v'", alg)
	}

	return cfg, nil
}
