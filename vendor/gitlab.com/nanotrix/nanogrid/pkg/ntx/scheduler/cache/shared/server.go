package shared

import (
	"gitlab.com/nanotrix/nanogrid/pkg/ntx"
	"golang.org/x/net/context"

	"strings"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/scheduler/cache"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	Server struct {
		keeper  Keeper
		evictor cache.Evictor
		owner   string
		log     log.Logger
	}

	ServerOpt func(s *Server)
)

func NewServer(keeper Keeper, evictor cache.Evictor, owner string, log log.Logger) *Server {
	return &Server{keeper: keeper, evictor: evictor, owner: owner, log: log}
}

func (s *Server) Ping(ctx context.Context, t *Tasks) (*ntx.Empty, error) {
	if err := t.Valid(); err != nil {
		return nil, status.Errorf(codes.NotFound, errors.Wrap(err, "tasks").Error())
	}

	s.keeper.Ping(t.Ids...)

	return &ntx.Empty{}, nil
}

func (s *Server) Release(ctx context.Context, t *Tasks) (*ntx.Empty, error) {
	if err := t.Valid(); err != nil {
		return nil, status.Errorf(codes.NotFound, errors.Wrap(err, "tasks").Error())
	}

	s.log.Log("ev", "req_to_release", "tasks", strings.Join(t.Ids, ","))
	s.keeper.Release(t.Ids...)

	return &ntx.Empty{}, nil
}

func (s *Server) Kill(ctx context.Context, t *Tasks) (*ntx.Empty, error) {
	if err := t.Valid(); err != nil {
		return nil, status.Errorf(codes.NotFound, errors.Wrap(err, "tasks").Error())
	}

	s.log.Log("ev", "req_to_kill", "tasks", strings.Join(t.Ids, ","))
	s.keeper.Kill(t.Ids...)

	return &ntx.Empty{}, nil
}

func (s *Server) Request(ctx context.Context, t *TaskRequest) (*TaskResponse, error) {
	if err := t.Valid(); err != nil {
		return nil, status.Errorf(codes.NotFound, errors.Wrap(err, "req").Error())
	}

	key := cache.TaskConfigurationToCacheKey(t.Config)
	task, found := s.keeper.Get(key)
	if !found {
		if t.Evict {
			ok := s.evictor.Evict(cache.ResourcesFromTaskConfig(t.Config))
			s.log.Log("req", key, "found", false, "evict", ok, "debug", true)
		}

		s.log.Log("req", key, "found", false, "debug", true)
		return &TaskResponse{Found: false}, nil
	} else {
		snap := task.Snapshot()
		s.log.Log("req", key, "found", true, "tid", snap.Id, "debug", true)

		return &TaskResponse{
			Found:    true,
			Id:       snap.Id,
			Endpoint: snap.Endpoint,
			Owner:    s.owner,
		}, nil
	}
}
