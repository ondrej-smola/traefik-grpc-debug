package gateway

import (
	"github.com/pkg/errors"
)

func (g *Config) Valid() error {
	if g == nil {
		return errors.New("is nil")
	}

	if g.Id == "" {
		return errors.New("id: is blank")
	}

	if g.Host == "" {
		return errors.New("host: is blank")
	}

	if err := g.LocalCache.Valid(); err != nil {
		return errors.Wrap(err, "localCache")
	}

	if err := g.SharedCache.Valid(); err != nil {
		return errors.Wrap(err, "sharedCache")
	}

	if err := g.Scheduler.Valid(); err != nil {
		return errors.Wrap(err, "scheduler")
	}

	if err := g.Membership.Valid(); err != nil {
		return errors.Wrap(err, "membership")
	}

	if err := g.Zookeeper.Valid(); err != nil {
		return errors.Wrap(err, "zookeeper")
	}

	if err := g.History.Valid(); err != nil {
		return errors.Wrap(err, "history")
	}

	if err := g.Proxy.Valid(); err != nil {
		return errors.Wrap(err, "proxy")
	}

	return nil
}

func (p *Config_Proxy) Valid() error {
	if p == nil {
		return errors.New("is nil")
	}

	if p.AllocationAttempts == 0 {
		return errors.New("allocationAttempts must be > 0")
	}

	if p.AllocationTimeoutSeconds == 0 {
		return errors.New("allocationTimeoutSeconds must be > 0")
	}

	if p.InactivityTimeoutSeconds == 0 {
		return errors.New("inactivityTimeoutSeconds must be > 0")
	}

	return nil
}

func (g *Config_Membership) Valid() error {
	if g == nil {
		return errors.New("is nil")
	}

	if err := g.Type.Valid(); err != nil {
		return errors.Wrap(err, "type")
	}

	return nil
}

func (g Config_Membership_Type) Valid() error {
	if _, ok := Config_Membership_Type_name[int32(g)]; !ok {
		return errors.Errorf("unsupported value: %v", g)
	}
	return nil
}

func (z *Config_Zookeeper) Valid() error {
	if z == nil {
		return errors.New("is nil")
	}

	if len(z.Endpoints) == 0 {
		return errors.New("endpoints: empty")
	}

	if z.TimeoutSeconds == 0 {
		return errors.New("timeoutSeconds must be > 0")
	}

	return nil
}

func (h *Config_History) Valid() error {
	if h == nil {
		return errors.New("is nil")
	}

	if h.MaxSize == 0 {
		return errors.New("maxSize must be > 0")
	}

	return nil
}

func (c *Config_LocalCache) Valid() error {
	if c == nil {
		return errors.New("is nil")
	}

	if err := c.Type.Valid(); err != nil {
		return errors.Wrap(err, "type")
	}

	if c.Type == Config_LocalCache_DISABLED {
		return nil
	}

	if c.MaxLineSize == 0 {
		return errors.New("maxLineSize must be > 0")
	}

	if c.GcIntervalSeconds == 0 {
		return errors.New("gcIntervalSeconds must be > 0")
	}

	if c.EvictIntervalSeconds == 0 {
		return errors.New("evictIntervalSeconds must be > 0")
	}

	if c.TtlSeconds == 0 {
		return errors.New("ttlSeconds must be > 0")
	}

	return nil
}

func (g Config_LocalCache_Type) Valid() error {
	if _, ok := Config_LocalCache_Type_name[int32(g)]; !ok {
		return errors.Errorf("unsupported value: %v", g)
	}
	return nil
}

func (c *Config_SharedCache) Valid() error {
	if c == nil {
		return errors.New("is nil")
	}

	if err := c.Type.Valid(); err != nil {
		return errors.Wrap(err, "type")
	}

	if c.Type == Config_SharedCache_DISABLED {
		return nil
	}

	if c.TtlSeconds == 0 {
		return errors.New("ttlSeconds must be > 0")
	}

	if c.MembershipRefreshIntervalSeconds == 0 {
		return errors.New("membershipRefreshIntervalSeconds must be > 0")
	}

	if c.KeepaliveIntervalSeconds == 0 {
		return errors.New("keepaliveIntervalSeconds must be > 0")
	}

	if c.TtlSeconds <= c.KeepaliveIntervalSeconds {
		return errors.Errorf("ttlSeconds '%v' must be > keepaliveIntervalSeconds '%v'", c.TtlSeconds, c.KeepaliveIntervalSeconds)
	}

	if c.GcIntervalSeconds == 0 {
		return errors.New("gcIntervalSeconds must be > 0")
	}

	return nil
}

func (g Config_SharedCache_Type) Valid() error {
	if _, ok := Config_SharedCache_Type_name[int32(g)]; !ok {
		return errors.Errorf("unsupported value: %v", g)
	}
	return nil
}

func (s *Config_Scheduler) Valid() error {

	if s == nil {
		return errors.New("is nil")
	}

	if s.Mesos != nil {
		return errors.Wrap(s.Mesos.Valid(), "mesos")
	}

	return errors.New("no option set")
}

func (m *Config_Scheduler_Mesos) Valid() error {
	if m == nil {
		return errors.New("is nil")
	}

	if len(m.Masters) == 0 {
		return errors.New("masters: is empty")
	}

	if m.OfferRefuseSeconds == 0 {
		return errors.New("offerRefuseSeconds  must be > 0")
	}

	return nil
}
