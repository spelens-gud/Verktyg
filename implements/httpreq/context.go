package httpreq

import (
	"github.com/spelens-gud/Verktyg.git/internal/incontext"
)

const (
	contextKeyHttpOption  = incontext.Key("httpreq.option")
	contextKeyRetriedTime = incontext.Key("httpreq.retried_times")
	contextKeyInTransport = incontext.Key("httpreq.in_transport")
)
