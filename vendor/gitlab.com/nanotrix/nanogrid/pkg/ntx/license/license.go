package license

import (
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

type Tasks []*Task
type Capabilities []Client_Capabilities_Capability
type Requirements []Client_Requirements_Requirement
type Clients []*Client

func (c *Cluster) Valid() error {
	if c == nil {
		return errors.Errorf("is nil")
	}

	if c.Id == "" {
		return errors.Errorf("id: is blank")
	}

	if c.Created == 0 {
		return errors.Errorf("created: is zero")
	}

	if c.Key == "" {
		return errors.Errorf("key: is blank")
	}

	if c.Hostname == "" {
		return errors.Errorf("hostname: is blank")
	}

	return nil
}

func (c *Cluster) Clone() *Cluster {
	cpy := *c
	return &cpy
}

func (c *Cluster) Equal(o *Cluster) bool {
	return c.Id == o.Id && c.Created == o.Created && c.Key == o.Key && c.Hostname == o.Hostname
}

func (l *License) Valid() error {
	if l == nil {
		return errors.Errorf("is nil")
	}

	if err := l.Cluster.Valid(); err != nil {
		return errors.Wrap(err, "cluster")
	}

	if err := l.Tasks.Valid(); err != nil {
		return errors.Wrap(err, "tasks")
	}

	if err := l.Limits.Valid(); err != nil {
		return errors.Wrap(err, "limits")
	}

	if l.CustomerId == "" {
		return errors.Errorf("customerId: is blank")
	}

	if l.Created == 0 {
		return errors.Errorf("created: is zero")
	}

	if l.Expire == 0 {
		return errors.Errorf("expire: is zero")
	}

	// do not check Capabilities and Requirements fro backwards compatibility

	return nil
}

func (l *License) Clone() *License {
	return &License{
		Limits:       l.Limits.Clone(),
		Cluster:      l.Cluster.Clone(),
		Tasks:        l.Tasks.Clone(),
		Capabilities: Capabilities(l.Capabilities).Clone(),
		Requirements: Requirements(l.Requirements).Clone(),
		CustomerId:   l.CustomerId,
		Created:      l.Created,
		Expire:       l.Expire,
	}
}

func (l *LicenseInfo) Valid() error {
	if l == nil {
		return errors.Errorf("is nil")
	}

	if err := l.Cluster.Valid(); err != nil {
		return errors.Wrap(err, "cluster")
	}

	if l.State != LicenseInfo_WAITING {
		if err := l.Limits.Valid(); err != nil {
			return errors.Wrap(err, "limits")
		}

		if err := l.State.Valid(); err != nil {
			return errors.Wrap(err, "state")
		}

		if l.Created == 0 {
			return errors.Errorf("created: is zero")
		}

		if l.Expire == 0 {
			return errors.Errorf("expire: is zero")
		}
	}

	return nil
}

func (l LicenseInfo_State) Valid() error {
	if _, ok := LicenseInfo_State_name[int32(l)]; !ok {
		return errors.Errorf("unknown license info state value: %v", l)
	}
	return nil
}

func (l *License) Expired() bool {
	return l.Expire < timeutil.NowUTCUnixSeconds()
}

func (l *Limits) Valid() error {
	if l == nil {
		return errors.Errorf("is nil")
	}

	if l.MonthlyCredit == 0 {
		return errors.Errorf("monthlyCredit: must be > 0")
	}

	return nil
}

func (l *Limits) Clone() *Limits {
	cpy := *l
	return &cpy
}

func (t Tasks) Valid() error {
	for i, c := range t {
		if err := c.Valid(); err != nil {
			return errors.Wrapf(err, "tasks[%v]", i)
		}
	}

	return nil
}

func (t Tasks) Clone() Tasks {
	clone := make([]*Task, len(t))
	for i, ct := range t {
		clone[i] = ct.Clone()
	}

	return clone
}

func (t *Task) Valid() error {
	if t == nil {
		return errors.Errorf("is nil")
	}

	if t.Service == "" {
		return errors.Errorf("service: is blank")
	}

	for i, p := range t.Profiles {
		if err := p.Valid(); err != nil {
			return errors.Wrapf(err, "Item: profile[%v]", i)
		}
	}

	return nil
}

func (t *Task) Clone() *Task {
	profClone := make([]*Task_Profile, len(t.Profiles))

	for i, p := range t.Profiles {
		profClone[i] = p.Clone()
	}

	return &Task{
		Service:  t.Service,
		Tags:     util.CloneStringSlice(t.Tags),
		Profiles: profClone,
	}
}

func (p *Task_Profile) Valid() error {
	if p == nil {
		return errors.Errorf("is nil")
	}

	if p.Id == "" {
		return errors.Errorf("id: is blank")
	}

	if err := p.Task.Valid(); err != nil {
		return errors.Wrap(err, "task")
	}

	return nil
}

func (p *Task_Profile) Clone() *Task_Profile {
	return &Task_Profile{
		Id:     p.Id,
		Labels: util.CloneStringSlice(p.Labels),
		Task:   p.Task.Clone(),
	}
}

func (c *Client) Valid() error {
	if c == nil {
		return errors.Errorf("is nil")
	}

	if c.CustomerId == "" {
		return errors.Errorf("customerId: is blank")
	}

	if c.ClusterId == "" {
		return errors.Errorf("clusterId: is blank")
	}

	if c.Tasks == "" {
		return errors.Errorf("tasks: is blank")
	}

	if err := c.Limits.Valid(); err != nil {
		return errors.Wrap(err, "limits")
	}

	if c.Expire == 0 {
		return errors.Errorf("expire: is zero")
	}

	if err := Requirements(c.Requirements).Valid(); err != nil {
		return errors.Wrap(err, "requirements")
	}

	if err := Capabilities(c.Capabilities).Valid(); err != nil {
		return errors.Wrap(err, "capabilities")
	}

	if c.Owner != nil {
		if err := c.Owner.Valid(); err != nil {
			return errors.Wrap(err, "owner")
		}
	}

	return nil
}

func (c Clients) Valid() error {
	for i, ic := range c {
		if err := ic.Valid(); err != nil {
			return errors.Wrapf(err, "clients[%v]", i)
		}
	}

	return nil
}

func (c Capabilities) Clone() Capabilities {
	caps := make([]Client_Capabilities_Capability, len(c))

	for i, ic := range c {
		caps[i] = ic
	}
	return caps
}

func (c Capabilities) Valid() error {
	for i, ic := range c {
		if _, ok := Client_Requirements_Requirement_name[int32(ic)]; !ok {
			return errors.Errorf("capability[%v]: unknown value %v", i, ic)
		}
	}

	return nil
}

func (r Requirements) Clone() Requirements {
	reqs := make([]Client_Requirements_Requirement, len(r))

	for i, ic := range r {
		reqs[i] = ic
	}
	return reqs
}

func (c Requirements) Valid() error {
	for i, ic := range c {
		if _, ok := Client_Requirements_Requirement_name[int32(ic)]; !ok {
			return errors.Errorf("requirement[%v]: unknown value %v", i, ic)
		}
	}

	return nil
}

func (c *Client) Clone() *Client {
	return &Client{
		ClusterId:          c.ClusterId,
		CustomerId:         c.CustomerId,
		Label:              c.Label,
		Tasks:              c.Tasks,
		Permissions:        util.CloneStringSlice(c.Permissions),
		Limits:             c.Limits.Clone(),
		Capabilities:       Capabilities(c.Capabilities).Clone(),
		Requirements:       Requirements(c.Requirements).Clone(),
		MaxLicenseDuration: c.MaxLicenseDuration,
		Expire:             c.Expire,
	}
}

func (c *Client) Expired() bool {
	return c.Expire < timeutil.NowUTCUnixSeconds()
}

func (c Clients) Clone() Clients {
	cpy := make([]*Client, len(c))

	for i, ic := range c {
		cpy[i] = ic.Clone()
	}

	return cpy
}

func (c Clients) Find(clusterId string) *Client {
	for _, c := range c {
		if c.ClusterId == clusterId {
			return c
		}
	}

	return nil
}

func (c *VersionedTasks) Valid() error {
	if c == nil {
		return errors.Errorf("is nil")
	}

	if c.Version == "" {
		return errors.Errorf("version: is blank")
	}

	if c.Created == 0 {
		return errors.Errorf("created: is zero")
	}

	return errors.Wrap(Tasks(c.Tasks).Valid(), "tasks")
}

func (c *VersionedTasks) Clone() *VersionedTasks {
	return &VersionedTasks{
		Version: c.Version,
		Created: c.Created,
		Tasks:   Tasks(c.Tasks).Clone(),
	}
}

func (caps Capabilities) Contains(tc Client_Capabilities_Capability) bool {
	for _, c := range caps {
		if c == tc {
			return true
		}
	}

	return false
}

func (r Requirements) Contains(cr Client_Requirements_Requirement) bool {
	for _, e := range r {
		if e == cr {
			return true
		}
	}

	return false
}

func (c *ChallengeRequest) Valid() error {
	if c == nil {
		return errors.Errorf("is nil")
	}

	if c.Challenge == "" {
		return errors.Errorf("challenge: is blank")
	}

	return nil
}

func (c *LicenseResponse) Valid() error {
	if c == nil {
		return errors.Errorf("is nil")
	}

	if c.License == "" {
		return errors.Errorf("license: is blank")
	}

	return nil
}

func (c *LicenseDecodeRequest) Valid() error {
	if c == nil {
		return errors.Errorf("is nil")
	}

	if c.Challenge == "" {
		return errors.Errorf("challenge: is blank")
	}

	if c.License == "" {
		return errors.Errorf("license: is blank")
	}

	return nil
}

func (l *CreditSnapshot) Valid() error {
	if l == nil {
		return errors.Errorf("is nil")
	}

	if l.Created == 0 {
		return errors.Errorf("created: is zero")
	}

	if l.ClusterId == "" {
		return errors.Errorf("clusterId: is blank")
	}

	return nil
}

func (l *CreditSnapshot) Clone() *CreditSnapshot {
	return &CreditSnapshot{
		ClusterId: l.ClusterId,
		Created:   l.Created,
		Credits:   util.CloneUint64Map(l.Credits),
	}
}

func (l *CreditSnapshot) Sum() uint64 {
	sum := uint64(0)
	for _, v := range l.Credits {
		sum += v
	}

	return sum
}

func (l *Challenge) Valid() error {
	if l == nil {
		return errors.New("is nil")
	}

	if err := l.Credit.Valid(); err != nil {
		return errors.Wrap(err, "credit")
	}

	if err := l.Cluster.Valid(); err != nil {
		return errors.Wrap(err, "cluster")
	}

	return nil
}

func (l *Challenge) Clone() *Challenge {
	return &Challenge{
		Cluster: l.Cluster.Clone(),
		Credit:  l.Credit.Clone(),
	}
}
