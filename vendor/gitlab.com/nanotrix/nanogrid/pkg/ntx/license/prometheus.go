package license

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Prometheus struct {
	creditTotal   prometheus.Gauge
	creditUsed    prometheus.Gauge
	licenseExpire prometheus.Gauge
}

func NewPrometheus(namespace string, subsystem string) *Prometheus {
	p := &Prometheus{}

	p.creditTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "credit_total",
		Help:      "total credit for this month",
	})

	p.creditUsed = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "credit_used",
		Help:      "used credit for this month",
	})

	p.licenseExpire = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "license_expire",
		Help:      "license expire unix seconds",
	})

	return p
}

func (p *Prometheus) SetCreditUsed(used uint64) {
	p.creditUsed.Set(float64(used))
}

func (p *Prometheus) SetCreditTotal(total uint64) {
	p.creditTotal.Set(float64(total))
}

func (p *Prometheus) SetLicenseExpire(ts uint64) {
	p.licenseExpire.Set(float64(ts))
}

func (p *Prometheus) MustRegister(r *prometheus.Registry) {
	r.MustRegister(p.creditTotal, p.creditUsed, p.licenseExpire)
}
