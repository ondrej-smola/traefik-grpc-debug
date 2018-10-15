package ntx

import ctx "golang.org/x/net/context"
import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type (
	ServiceNameWithContext struct {
		Service string
		Context context.Context
	}

	ServingStatusOrError struct {
		Status HealthCheckResponse_ServingStatus
		Err    error
	}

	HealthCheckServer struct {
	}
)

func NewHealthCheckServer() *HealthCheckServer {

	return &HealthCheckServer{}
}

func (s *HealthCheckServer) Check(ctx ctx.Context, req *HealthCheckRequest) (*HealthCheckResponse, error) {
	return &HealthCheckResponse{Status: HealthCheckResponse_SERVING}, nil
}

// return connection and headers returned during health check
func ConnectIfHealthy(endpoint string, serviceName string, ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, metadata.MD, error) {
	c, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Dial:")
	}

	cl := NewHealthClient(c)

	var header metadata.MD
	hResp, err := cl.Check(ctx, &HealthCheckRequest{Service: serviceName}, grpc.FailFast(true), grpc.Header(&header))

	if err != nil {
		c.Close()
		return nil, nil, err
	} else if hResp.Status != HealthCheckResponse_SERVING {
		c.Close()
		return nil, nil, errors.Errorf("Expected serving status, got %v", hResp.Status)
	} else {
		return c, header, nil
	}
}
