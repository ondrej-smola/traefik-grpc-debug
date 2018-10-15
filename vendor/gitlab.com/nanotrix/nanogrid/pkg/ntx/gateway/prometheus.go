package gateway

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/scheduler"
)

type promClientEntry struct {
	task    string
	waiting bool
}

type PrometheusEvents struct {
	total   *prometheus.CounterVec
	waiting *prometheus.GaugeVec
	active  *prometheus.GaugeVec
	failed  *prometheus.CounterVec

	lock          *sync.Mutex
	activeClients map[string]*promClientEntry
}

var _ = ClientEvents(&PrometheusEvents{})

func NewPrometheusEvents(namespace string, subsystem string) *PrometheusEvents {
	p := &PrometheusEvents{
		lock:          &sync.Mutex{},
		activeClients: make(map[string]*promClientEntry),
	}

	p.total = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "requests_total",
		Help:      "total number of requests",
	}, []string{"task"})

	p.waiting = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "requests_waiting",
		Help:      "current number of requests waiting for resources",
	}, []string{"task"})

	p.active = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "requests_active",
		Help:      "current number of waiting clients",
	}, []string{"task"})

	p.failed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "requests_failed",
		Help:      "total number of failed requests",
	}, []string{"task"})

	return p
}

func (p *PrometheusEvents) New(id string, task *auth.NtxToken) {
	t := task.Task.Id + "/" + task.Task.Label
	p.lock.Lock()
	if _, ok := p.activeClients[id]; !ok {
		p.activeClients[id] = &promClientEntry{task: t, waiting: true}
		p.waiting.WithLabelValues(t).Add(1)
	}
	p.lock.Unlock()
}

func (p *PrometheusEvents) Scheduled(id string, snap *scheduler.TaskSnapshot) {}

func (p *PrometheusEvents) Started(id string) {
	p.lock.Lock()
	if c, ok := p.activeClients[id]; ok && c.waiting {
		c.waiting = false
		p.waiting.WithLabelValues(c.task).Sub(1)
		p.active.WithLabelValues(c.task).Add(1)
	}
	p.lock.Unlock()
}

func (p *PrometheusEvents) Completed(id string, err ...string) {
	p.lock.Lock()
	c, ok := p.activeClients[id]
	if ok {
		if c.waiting {
			p.waiting.WithLabelValues(c.task).Sub(1)
		} else {
			p.active.WithLabelValues(c.task).Sub(1)
		}

		p.total.WithLabelValues(c.task).Add(1)
		if len(err) > 0 {
			p.failed.WithLabelValues(c.task).Add(1)
		}

		delete(p.activeClients, id)
	}
	p.lock.Unlock()
}

func (p *PrometheusEvents) MsgIn(id string, size uint64)          {}
func (p *PrometheusEvents) MsgOut(id string, size uint64)         {}
func (p *PrometheusEvents) StreamLength(id string, offset uint64) {}

func (p *PrometheusEvents) MustRegister(r *prometheus.Registry) {
	r.MustRegister(p.total, p.waiting, p.active, p.failed)
}
