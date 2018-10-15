package hello

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"golang.org/x/net/context"
)

type HelloService struct {
	log log.Logger
}

func NewHelloService(log log.Logger) *HelloService {
	return &HelloService{log: log}
}

func (h *HelloService) SayHello(ctx context.Context, req *Hello) (*Hello, error) {
	return &Hello{Msg: fmt.Sprintf("Hello %v", req.Msg)}, nil

}
func (h *HelloService) SayHelloNTimes(r *Hello, s HelloService_SayHelloNTimesServer) error {
	for i := 0; i < int(r.Times); i++ {
		if err := s.Send(&Hello{Msg: fmt.Sprintf("Hello %v (%v)", r.Msg, i)}); err != nil {
			h.log.Log("err", err)
			return err
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func (h *HelloService) SayHelloBidi(s HelloService_SayHelloBidiServer) error {
	i := 0
	for {
		r, err := s.Recv()
		if err != nil {
			h.log.Log("recv_err", err)
			return err
		}

		err = s.Send(&Hello{Msg: fmt.Sprintf("Hello %v (%v)", r.Msg, i)})
		if err != nil {
			h.log.Log("send_err", err)
			return err
		}
		i++
	}
}
