package initializers

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	CountRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "Number_of_requests",
		Help: "The total number of processed requests",
	})
	CountStatusCodes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "Number_of_status_codes",
		Help: "The total number of status codes returned",
	}, []string{"status_code"})
)
