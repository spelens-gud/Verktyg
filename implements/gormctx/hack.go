package gormctx

import (
	"reflect"
	"unsafe"

	"github.com/jinzhu/gorm"
)

func init() {
	f, ok := reflect.ValueOf(&gorm.DB{}).Type().Elem().FieldByName("db")
	if !ok {
		panic("invalid gorm DB,check gorm mod version")
	}
	gormDBOffset = f.Offset
}

func setGormCommonDB(db *gorm.DB, common gorm.SQLCommon) {
	db.Dialect().SetDB(common)
	*(*gorm.SQLCommon)(unsafe.Pointer(uintptr(unsafe.Pointer(db)) + gormDBOffset)) = common
}

func replaceTransactionHook(db *gorm.DB) {
	db = db.New()
	db.SetLogger(&gormLogger{db: &ctxConn{context: ctxBackground}})
	callbacks := db.Callback()
	callbacks.Update().Replace("gorm:begin_transaction", beginTx)
	callbacks.Create().Replace("gorm:begin_transaction", beginTx)
	callbacks.Delete().Replace("gorm:begin_transaction", beginTx)
}
