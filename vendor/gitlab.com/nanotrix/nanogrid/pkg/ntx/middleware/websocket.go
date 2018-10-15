package middleware

import (
	"net/url"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth"
)

type WsNtxTokenFn func(val url.Values) (*auth.NtxToken, error)

func NtxTokenFromWebsocketURLValues(ntxDec auth.Decoder) WsNtxTokenFn {
	return func(val url.Values) (*auth.NtxToken, error) {
		if ntxTokenString := val.Get(auth.NtxTokenHeaderName); ntxTokenString == "" {
			return nil, auth.ErrMissingNtxToken
		} else {
			return ntxTokenFromString(ntxTokenString, ntxDec)
		}
	}
}
