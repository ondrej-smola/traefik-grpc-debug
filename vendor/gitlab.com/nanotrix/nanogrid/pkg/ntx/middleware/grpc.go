package middleware

import (
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/grpcutil"
	"google.golang.org/grpc/metadata"
)

type GrpcNtxTokenFn func(md metadata.MD) (*auth.NtxToken, error)

func NtxTokenFromGrpcMetadata(dec auth.Decoder) GrpcNtxTokenFn {
	return func(md metadata.MD) (*auth.NtxToken, error) {
		if ntxTokenString := grpcutil.GetFirstFromMD(md, auth.NtxTokenHeaderName); ntxTokenString == "" {
			return nil, auth.ErrMissingNtxToken
		} else {
			return ntxTokenFromString(ntxTokenString, dec)
		}
	}
}
