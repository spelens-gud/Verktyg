package isql

import (
	"github.com/spelens-gud/Verktyg/kits/kstruct"
)

type (
	// deprecated: 废弃 不建议使用
	MultiSqlFactory func(cfg MultiWRSQLConfig, multiDB interface{}) (cf func(), err error)
	// deprecated: 废弃 不建议使用
	WRSqlConstructor func(config WRSQLConfig) (GormSQL, error)
)

// deprecated: 废弃 不建议使用
func NewMultiSqlFactory(f WRSqlConstructor) MultiSqlFactory {
	return func(cfg MultiWRSQLConfig, multiDB interface{}) (cf func(), err error) {
		cf, err = kstruct.MultiFieldsBuilder(cfg, multiDB, "db", func(i interface{}) (ret interface{}, err error) {
			return f(i.(WRSQLConfig))
		}, func(i interface{}) func() {
			return func() {
				_ = i.(GormSQL).Close()
			}
		})
		return
	}
}
