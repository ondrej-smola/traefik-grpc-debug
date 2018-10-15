package gateway

import (
	"context"

	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/license"
)

func CreditLimiterProxyFn(acceptor license.RequestAcceptor, next DoFn) DoFn {
	return func(reqId string, from Src, to DestFn, tkn *auth.NtxToken, ctx context.Context) error {
		if !acceptor.Accept() {
			return errors.Errorf("credit limit reached")
		}

		return next(reqId, from, to, tkn, ctx)
	}
}
