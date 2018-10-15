package license

import (
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/httputil"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

type PiracyChecker interface {
	Check() error
}

type MesosHostnamePiracyCheckFn func() error

func (a MesosHostnamePiracyCheckFn) Check() error {
	return a()
}

// check whether current master hostname is same as expected
type MesosHostnamePiracyChecker struct {
	master   string
	hostname string
}

func NewMesosHostnamePiracyChecker(endpoint string, hostname string) *MesosHostnamePiracyChecker {
	return &MesosHostnamePiracyChecker{
		master:   endpoint,
		hostname: hostname,
	}
}

func (p *MesosHostnamePiracyChecker) Check() error {
	resp, err := httputil.DefaultClient.Get(p.master)
	if err != nil {
		return errors.Wrap(err, "get")
	}

	respBody := make(map[string]interface{})
	if err := util.ReadCloserToJson(resp.Body, &respBody); err != nil {
		return errors.Wrap(err, "unmarshal body")
	}

	if _, ok := respBody["hostname"]; !ok {
		return errors.Errorf("invalid mesos state.json")
	}

	hostname, ok := respBody["hostname"].(string)
	if !ok {
		return errors.Errorf("invalid mesos state.json")
	}

	if hostname != p.hostname {
		return errors.Errorf("hostname mismatch, expected %v, got %v", p.hostname, hostname)
	}

	return nil
}
