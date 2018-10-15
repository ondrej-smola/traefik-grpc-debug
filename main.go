package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/ondrej-smola/traefik-grpc-debug/event"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"os"
	"time"
)

var logger = log.With(log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)), "ts", log.DefaultTimestampUTC)

func exit(msg interface{}) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

var isServer bool
var endpoint string
var bind string

func main() {
	flag.BoolVar(&isServer, "server", false, "start server")
	flag.StringVar(&endpoint, "endpoint", "localhost:443", "server endpoint")
	flag.StringVar(&bind, "bind", ":8080", "server bind address")

	flag.Parse()

	if isServer {
		runServer()
		return
	}

	c, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})))
	if err != nil {
		exit(errors.Wrapf(err, "dial %v", endpoint))
	}

	defer c.Close()

	cl := event.NewEventServiceClient(c)

	stream, err := cl.Events(context.Background())
	if err != nil {
		exit(errors.Wrap(err, "events"))
	}

	smallMsg := make([]byte, 64)
	largeMsg := make([]byte, 8096)

	for i := 0; i < 50; i++ {
		ev := &event.Event{Id: uint32(i)}

		if i%2 == 0 {
			ev.Payload = smallMsg
		} else {
			ev.Payload = largeMsg
		}

		if err := stream.Send(ev); err != nil {
			exit(errors.Wrap(err, "send"))
		}

		logger.Log("msg_out", ev.Id, "len", len(ev.Payload))
		ev, err := stream.Recv()
		if err != nil {
			exit(errors.Wrap(err, "recv"))
		}

		logger.Log("msg_in", ev.Id, "len", len(ev.Payload))
	}
}

type server struct {
}

var _ = event.EventServiceServer(&server{})

func (s *server) Events(srv event.EventService_EventsServer) error {
	for {
		ev, err := srv.Recv()
		if err != nil {
			logger.Log("recv_err", err)
			return errors.Wrap(err, "recv")
		}

		logger.Log("msg_in", ev.Id, "len", len(ev.Payload))

		time.Sleep(50 * time.Millisecond) // simulate some processing

		// send the same message back
		if err := srv.Send(ev); err != nil {
			logger.Log("send_err", err)
			return errors.Wrap(err, "send")
		}

		logger.Log("msg_out", ev.Id)
	}
}

func runServer() {
	l, err := net.Listen("tcp", bind)
	if err != nil {
		exit(errors.Wrap(err, "net_listen"))
	}

	grpcServer := grpc.NewServer()
	event.RegisterEventServiceServer(grpcServer, &server{})

	logger.Log("listening", bind)
	if err := grpcServer.Serve(l); err != nil {
		exit(errors.Wrap(err, "grpc_server"))
	}
}
