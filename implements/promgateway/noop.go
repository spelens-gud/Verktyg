package promgateway

import "github.com/spelens-gud/Verktyg.git/interfaces/imetrics"

var _ imetrics.GatewayDaemon = NoopDaemon{}

type NoopDaemon struct{}

func (e NoopDaemon) Stop()        {}
func (e NoopDaemon) StartDaemon() {}
