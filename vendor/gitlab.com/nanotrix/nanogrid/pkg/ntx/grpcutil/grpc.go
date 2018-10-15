package grpcutil

import (
	"context"

	"net"
	"net/url"

	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	cause := errors.Cause(err)

	if s, ok := status.FromError(err); ok {
		return s.Err()
	}

	switch {
	case util.IsNotFound(cause):
		return status.Errorf(codes.NotFound, err.Error())
	default:
		return status.Errorf(codes.Internal, err.Error())
	}
}

func ErrorDesc(err error) string {
	if s, ok := status.FromError(err); ok {
		return s.Message()
	}
	return err.Error()
}

func ErrorCode(err error) codes.Code {
	if s, ok := status.FromError(err); ok {
		return s.Code()
	}
	return codes.Unknown
}

func GetFirstFromMD(md metadata.MD, key string) string {
	res, ok := md[key]
	if !ok || len(res) == 0 {
		return ""
	} else {
		return res[0]
	}
}

func IsContextCancelled(err error) bool {
	st, ok := status.FromError(errors.Cause(err))
	if !ok {
		return err == context.Canceled
	} else {
		return st.Code() == codes.Canceled
	}
}

func HttpEndpointToGrpcEndpoint(in string) (string, bool, error) {
	u, err := url.Parse(in)
	if err != nil {
		return "", false, err
	}

	if !u.IsAbs() {
		return "", false, errors.New("not and absolute URL")
	}

	secure := u.Scheme == "https"

	res := u.Host
	if u.Port() == "" {
		if secure {
			res = net.JoinHostPort(u.Host, "443")
		} else {
			res = net.JoinHostPort(u.Host, "80")
		}
	}

	return res, secure, nil
}
