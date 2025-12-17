package internal

import "github.com/spelens-gud/Verktyg.git/interfaces/ilog"

type levelLog ilog.Level

func (l levelLog) Compare(level ilog.Level, _ string, _ ...interface{}) bool {
	return level >= ilog.Level(l)
}
