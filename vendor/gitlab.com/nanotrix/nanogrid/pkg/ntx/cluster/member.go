package cluster

import (
	"github.com/pkg/errors"
)

type (
	Member struct {
		Id       string `json:"id"`
		Host     string `json:"host"`
		HttpPort uint32 `json:"httpPort"`
		GrpcPort uint32 `json:"grpcPort"`
	}

	Members []Member
)

func (m Members) Ids() []string {
	res := make([]string, len(m))
	for i := range m {
		res[i] = m[i].Id
	}

	return res
}

func (m Members) ContainsId(id string) bool {
	for i := range m {
		if m[i].Id == id {
			return true
		}
	}

	return false
}

func (m Members) Clone() Members {
	res := make(Members, len(m))

	for i, mb := range m {
		res[i] = mb
	}
	return res
}

func (m Member) Valid() error {
	if m.Id == "" {
		return errors.New("id: is blank")
	}

	if m.Host == "" {
		return errors.New("host: is blank")
	}

	if m.HttpPort == 0 {
		return errors.New("httpPort: not set")
	}

	if m.GrpcPort == 0 {
		return errors.New("grpcPort: not set")
	}

	return nil
}
