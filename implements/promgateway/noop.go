package promgateway

import "git.bestfulfill.tech/devops/go-core/interfaces/imetrics"

var _ imetrics.GatewayDaemon = NoopDaemon{}

type NoopDaemon struct{}

func (e NoopDaemon) Stop()        {}
func (e NoopDaemon) StartDaemon() {}
