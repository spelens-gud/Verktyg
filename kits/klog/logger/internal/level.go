package internal

import "git.bestfulfill.tech/devops/go-core/interfaces/ilog"

type levelLog ilog.Level

func (l levelLog) Compare(level ilog.Level, _ string, _ ...interface{}) bool {
	return level >= ilog.Level(l)
}
