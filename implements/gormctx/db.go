package gormctx

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/jinzhu/gorm"
	"go.uber.org/multierr"

	"github.com/spelens-gud/Verktyg/interfaces/isql"
)

var (
	_ isql.GormSQL = &wrdb{}
	_ isql.GormSQL = &wdb{}

	poolMutex sync.Mutex
	dbPool    = make(map[*gorm.DB]*sync.Pool)
)

func newDBPool(db *gorm.DB) *sync.Pool {
	detailMode, _ := db.Get("detailMode")
	detailModeOn, _ := detailMode.(bool)

	p := &sync.Pool{}
	p.New = func() interface{} {
		newDb := db.New()
		dbConn := &ctxConn{}
		setGormCommonDB(newDb, dbConn)
		newDb.SetLogger(&gormLogger{db: dbConn, detailMode: detailModeOn})
		runtime.SetFinalizer(newDb, func(newDb *gorm.DB) {
			if _, ok := newDb.CommonDB().(*ctxConn); ok {
				p.Put(newDb)
			}
		})
		return newDb
	}
	return p
}

func newCtxGormDB(ctx context.Context, db *gorm.DB, write, read connDB) (ndb *gorm.DB) {
	poolMutex.Lock()
	pool, ok := dbPool[db]
	if !ok {
		pool = newDBPool(db)
		dbPool[db] = pool
	}
	poolMutex.Unlock()

	ndb = pool.Get().(*gorm.DB)
	conn := ndb.CommonDB().(*ctxConn)
	conn.context = ctx
	conn.read = read
	conn.write = write
	return ndb
}

type wrdb struct {
	wdb
	read []connDB
	idx  int64
}

type wdb struct {
	write  connDB
	gormDB *gorm.DB
}

// 获取根据操作自动区分读写库的gorm.DB
func (db *wrdb) GetDB(ctx context.Context) *gorm.DB {
	if tx, ok := isql.ExtractTransaction(ctx, isql.GormSQL(db)).(*gorm.DB); ok {
		return tx
	}
	return newCtxGormDB(ctx, db.gormDB, db.write, db.getRead())
}

// 获取读库 使用原子增进行分片递增均衡
func (db *wrdb) getRead() connDB {
	if len(db.read) == 0 {
		return db.write
	}
	if len(db.read) == 1 {
		return db.read[0]
	}
	return db.read[int(atomic.AddInt64(&db.idx, 1))%len(db.read)]
}

// 关闭所有连接池
func (db *wrdb) Close() error {
	var errs []error

	if e := db.write.Close(); e != nil {
		errs = append(errs, e)
	}

	for _, r := range db.read {
		if e := r.Close(); e != nil {
			errs = append(errs, e)
		}
	}
	return multierr.Combine(errs...)
}

func (db *wdb) GetDB(ctx context.Context) *gorm.DB {
	if tx, ok := isql.ExtractTransaction(ctx, isql.GormSQL(db)).(*gorm.DB); ok {
		return tx
	}
	return newCtxGormDB(ctx, db.gormDB, db.write, db.write)
}

// Deprecated: 使用 Transaction
func (db *wdb) WithTransaction(ctx context.Context, f func(tx *gorm.DB) error) (err error) {
	return withTransaction(ctx, db, func(_ context.Context, tx *gorm.DB) error {
		return f(tx)
	})
}

func (db *wdb) Transaction(ctx context.Context, f func(ctx context.Context, tx *gorm.DB) (err error)) (err error) {
	return withTransaction(ctx, db, f)
}

func (db *wdb) Close() error {
	return db.write.Close()
}
