package auth

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func (t *NtxToken) UnmarshalJSON(b []byte) error {
	type Token struct {
		// Issuer
		Iss string `json:"iss,omitempty"`
		// IssuedAt
		Iat int64 `json:"iat,omitempty"`
		// NotBefore
		Nbf int64 `json:"nbf,omitempty"`
		// Expires
		Exp int64 `json:"exp,omitempty"`
		// Audience
		Aud interface{} `json:"aud,omitempty"`
		// Subject
		Sub string `json:"sub,omitempty"`
		// Nonce
		Jti            string         `json:"jti,omitempty"`
		Permissions    []string       `json:"permissions,omitempty"`
		Task           *NtxToken_Task `json:"task,omitempty"`
		NtxPermissions []string       `json:"https://nanotrix.cz/permissions"`
		NtxEmail       string         `json:"https://nanotrix.cz/email"`
		Email          string         `json:"email,omitempty"`
	}

	resAud := []string{}

	tkn := &Token{}

	if err := json.Unmarshal(b, tkn); err != nil {
		return err
	}

	switch aud := tkn.Aud.(type) {
	case string:
		resAud = []string{aud}
	case []interface{}:
		resAud = make([]string, len(aud))
		for i, a := range aud {
			if s, ok := a.(string); !ok {
				return errors.Errorf("aud[%v]: expected string got '%T'", i, a)
			} else {
				resAud[i] = s
			}
		}
	case nil:
	default:
		return errors.Errorf("aud: expected string array, got %T", tkn.Aud)
	}

	t.Aud = resAud
	t.Nbf = tkn.Nbf
	t.Exp = tkn.Exp
	t.Iat = tkn.Iat
	t.Sub = tkn.Sub
	t.Iss = tkn.Iss
	t.Task = tkn.Task
	t.Jti = tkn.Jti

	if tkn.NtxEmail != "" {
		t.Email = tkn.NtxEmail
	}

	// overwrite NtxEmail if Email specified
	if tkn.Email != "" {
		t.Email = tkn.Email
	}

	t.Permissions = append(tkn.Permissions, tkn.NtxPermissions...)

	return nil
}

type Token struct {
	// Issuer
	Iss string `json:"iss,omitempty"`
	// IssuedAt
	Iat int64 `json:"iat,omitempty"`
	// NotBefore
	Nbf int64 `json:"nbf,omitempty"`
	// Expires
	Exp int64 `json:"exp,omitempty"`
	// Audience
	Aud interface{} `json:"aud,omitempty"`
	// Subject
	Sub string `json:"sub,omitempty"`
	// Nonce
	Jti            string         `json:"jti,omitempty"`
	Permissions    []string       `json:"permissions,omitempty"`
	Task           *NtxToken_Task `json:"task,omitempty"`
	NtxPermissions []string       `json:"https://nanotrix.cz/permissions"`
	NtxEmail       string         `json:"https://nanotrix.cz/email"`
	Email          string         `json:"email,omitempty"`
}
