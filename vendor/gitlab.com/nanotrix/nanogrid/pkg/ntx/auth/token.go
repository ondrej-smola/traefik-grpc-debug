package auth

import (
	"context"
	"strings"

	"fmt"

	"github.com/pkg/errors"
)

const HttpContextNtxToken = "ntx-token"

type NtxTokenFn func() (string, error)

func SetNtxTokenForContext(ctx context.Context, tkn *NtxToken) context.Context {
	return context.WithValue(ctx, HttpContextNtxToken, tkn)
}

func GetNtxTokenFromContext(ctx context.Context) (*NtxToken, bool) {
	val := ctx.Value(HttpContextNtxToken)
	if tkn, ok := val.(*NtxToken); ok {
		return tkn, true
	} else {
		return nil, false
	}
}

func TokenStringFromAuthorizationHeader(header string) (string, error) {
	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || strings.ToLower(headerParts[0]) != "bearer" {
		return "", errors.New("Authorization header format must be Bearer {token}")
	}

	return headerParts[1], nil
}

func BearerAuthorizationHeaderName(val string) string {
	return fmt.Sprintf("Bearer %v", val)
}
