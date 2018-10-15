package gateway

import (
	"sync"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/license"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/scheduler"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
)

type (
	ClientEvents interface {
		New(id string, task *auth.NtxToken)
		Scheduled(id string, snap *scheduler.TaskSnapshot)
		Started(id string)
		Completed(id string, err ...string)
		MsgIn(id string, size uint64)
		MsgOut(id string, size uint64)
		StreamLength(id string, offset uint64)
	}

	ClientsSnapshot interface {
		Snapshot() *Snapshot_Clients
	}

	ClientEventsStore struct {
		completedMaxSize int

		sync.Mutex // for
		counter    uint64
		completed  []*Client
		active     map[string]*Client
	}

	BillingsEventsStore struct {
		store  license.CreditStore
		lock   *sync.Mutex
		active map[string]uint64
	}

	ClientEventsSeq []ClientEvents

	ClientEventsStoreOpt func(s *ClientEventsStore)
)

func WithClientEventsStoreCompletedMaxSize(max int) ClientEventsStoreOpt {
	return func(s *ClientEventsStore) {
		s.completedMaxSize = max
	}
}

var _ = ClientEvents(&ClientEventsStore{})
var _ = ClientsSnapshot(&ClientEventsStore{})

func NewClientEventsStore(opts ...ClientEventsStoreOpt) *ClientEventsStore {
	s := &ClientEventsStore{
		active:           map[string]*Client{},
		completedMaxSize: 128,
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

func (s *ClientEventsStore) New(reqId string, token *auth.NtxToken) {
	s.Lock()
	s.active[reqId] = &Client{
		Id:      reqId,
		Token:   token.Clone(),
		State:   Client_WAITING,
		Created: uint64(timeutil.NowUTC().Unix()),
	}
	s.Unlock()
}

func (s *ClientEventsStore) MsgIn(id string, size uint64) {
	s.Lock()
	if c, ok := s.active[id]; ok {
		c.MsgsIn += 1
		c.BytesIn += size
	}
	s.Unlock()
}

func (s *ClientEventsStore) MsgOut(id string, size uint64) {
	s.Lock()
	if c, ok := s.active[id]; ok {
		c.MsgsOut += 1
		c.BytesOut += size
	}
	s.Unlock()
}

func (s *ClientEventsStore) StreamLength(id string, offset uint64) {
	s.Lock()
	if c, ok := s.active[id]; ok {
		c.StreamLength = offset
	}
	s.Unlock()
}

func (s *ClientEventsStore) Scheduled(reqId string, snap *scheduler.TaskSnapshot) {
	s.Lock()
	if cl, ok := s.active[reqId]; ok {
		cl.State = Client_SCHEDULED
		cl.TaskId = snap.Id
		cl.TaskEndpoint = snap.Endpoint
		cl.Scheduled = uint64(timeutil.NowUTC().Unix())
	}
	s.Unlock()
}

func (s *ClientEventsStore) Started(reqId string) {
	s.Lock()
	if cl, ok := s.active[reqId]; ok {
		cl.State = Client_RUNNING
		cl.Started = uint64(timeutil.NowUTC().Unix())
	}
	s.Unlock()
}

func (s *ClientEventsStore) Completed(id string, optErr ...string) {
	err := ""
	state := Client_COMPLETED
	if len(optErr) > 0 {
		err = optErr[0]
		state = Client_FAILED
	}

	s.Lock()
	if cl, ok := s.active[id]; ok {
		cl.State = state
		cl.Completed = uint64(timeutil.NowUTC().Unix())
		cl.Error = err
		cl.Seqn = s.counter
		s.completed = append(s.completed, cl)
		if len(s.completed) > s.completedMaxSize {
			s.completed = s.completed[len(s.completed)-s.completedMaxSize:]
		}
	}
	delete(s.active, id)
	s.counter += 1
	s.Unlock()
}

func (s *ClientEventsStore) Snapshot() *Snapshot_Clients {
	cls := &Snapshot_Clients{}

	s.Lock()
	cls.Active = make([]*Client, len(s.active))
	i := 0
	for _, a := range s.active {
		// make shallow copy (we do not need to copy token as it cannot change after it is set
		cpy := *a
		cls.Active[i] = &cpy
		i++
	}

	// completed clients are immutable
	cls.Completed = make([]*Client, len(s.completed))
	copy(cls.Completed, s.completed)
	s.Unlock()

	return cls
}

func NewBillingEventsStore(store license.CreditStore) *BillingsEventsStore {
	return &BillingsEventsStore{
		store:  store,
		lock:   &sync.Mutex{},
		active: make(map[string]uint64),
	}
}

var _ = ClientEvents(&BillingsEventsStore{})

func (s *BillingsEventsStore) New(reqId string, token *auth.NtxToken) {
	s.lock.Lock()
	s.active[reqId] = 0
	s.lock.Unlock()
}

func (s *BillingsEventsStore) MsgIn(id string, size uint64)  {}
func (s *BillingsEventsStore) MsgOut(id string, size uint64) {}

func (s *BillingsEventsStore) StreamLength(id string, offset uint64) {
	diff := uint64(0)
	s.lock.Lock()
	if prev, ok := s.active[id]; ok {
		now := offset / timeutil.TicksPerSecond
		if now > prev {
			diff = now - prev
			s.active[id] = now
		}
	}
	s.lock.Unlock()
	if diff > 0 {
		s.store.Use(diff)
	}
}

func (s *BillingsEventsStore) Scheduled(reqId string, snap *scheduler.TaskSnapshot) {}
func (s *BillingsEventsStore) Started(reqId string)                                 {}

func (s *BillingsEventsStore) Completed(id string, err ...string) {
	s.lock.Lock()
	delete(s.active, id)
	s.lock.Unlock()
}

var _ = ClientEvents(ClientEventsSeq{})

func NewClientEventsSeq(evs ...ClientEvents) ClientEventsSeq {
	return evs
}

func (s ClientEventsSeq) New(id string, token *auth.NtxToken) {
	for _, c := range s {
		c.New(id, token)
	}
}

func (s ClientEventsSeq) MsgIn(id string, size uint64) {
	for _, c := range s {
		c.MsgIn(id, size)
	}
}

func (s ClientEventsSeq) MsgOut(id string, size uint64) {
	for _, c := range s {
		c.MsgOut(id, size)
	}
}

func (s ClientEventsSeq) StreamLength(id string, offset uint64) {
	for _, c := range s {
		c.StreamLength(id, offset)
	}
}

func (s ClientEventsSeq) Scheduled(reqId string, snap *scheduler.TaskSnapshot) {
	for _, c := range s {
		c.Scheduled(reqId, snap)
	}

}
func (s ClientEventsSeq) Started(reqId string) {
	for _, c := range s {
		c.Started(reqId)
	}
}

func (s ClientEventsSeq) Completed(id string, err ...string) {
	for _, c := range s {
		c.Completed(id, err...)
	}
}
