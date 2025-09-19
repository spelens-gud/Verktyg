package httpreq

import (
	"git.bestfulfill.tech/devops/go-core/internal/incontext"
)

const (
	contextKeyHttpOption  = incontext.Key("httpreq.option")
	contextKeyRetriedTime = incontext.Key("httpreq.retried_times")
	contextKeyInTransport = incontext.Key("httpreq.in_transport")
)
