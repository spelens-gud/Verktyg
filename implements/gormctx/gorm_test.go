package gormctx

import (
	"context"
	"math/rand"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"gotest.tools/assert"

	"git.bestfulfill.tech/devops/go-core/interfaces/ihttp"
	"git.bestfulfill.tech/devops/go-core/interfaces/isql"
	"git.bestfulfill.tech/devops/go-core/kits/kcontext"
)

func testInit() (g isql.GormSQL, cf func(), err error) {
	var d = struct {
		DB isql.GormSQL `db:"db"`
	}{}
	cf, err = NewMultiGormSQL(map[string]isql.WRSQLConfig{
		"db": {
			Write: isql.SQLConfig{
				Host:         "localhost",
				Port:         3306,
				DB:           "config_manager",
				User:         "root",
				Password:     "Aa123456",
				Location:     "",
				MaxIdleConns: 1000,
			},
			Reads: []isql.SQLConfig{
				{
					Host:         "127.0.0.1",
					Port:         3306,
					DB:           "config_manager",
					User:         "root",
					Password:     "Aa123456",
					Location:     "",
					MaxIdleConns: 1000,
				},
				{
					Host:         "0.0.0.0",
					Port:         3306,
					DB:           "config_manager",
					User:         "root",
					Password:     "Aa123456",
					Location:     "",
					MaxIdleConns: 1000,
				},
			},
		},
	}, &d)
	if err != nil {
		return
	}
	g = d.DB
	return
}

func TestTx(t *testing.T) {
	g, cf, err := testInit()
	if err != nil {
		t.Fatal(err)
	}
	defer cf()
	testTx(t, g)
}

func TestServer(t *testing.T) {
	g, cf, err := testInit()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		time.Sleep(time.Second)
		cf()
	}()
	testServer(t, g)
}

func testUpdateError(db *gorm.DB) (err error) {
	err = db.Table("config_ref").Where("id=1").
		Updates(map[string]interface{}{
			"data_error": rand.Intn(100),
		}).Error
	return
}

func testQuery(db *gorm.DB) (err error) {
	err = db.Table("config_ref").Where("id=1").
		Find(&struct{}{}).Error
	return
}

func testUpdate(db *gorm.DB) (err error) {
	err = db.Table("config_ref").Where("id=1").
		Updates(map[string]interface{}{
			"data_id": rand.Intn(100),
		}).Error
	return
}

func testRow(db *gorm.DB) (err error) {
	r := db.Table("config_ref").Where("id=1").Select("data_id").Row()
	s := ""
	return r.Scan(&s)
}

func testExec(db *gorm.DB) (err error) {
	return db.Exec("UPDATE `config_ref` SET `data_id` = ?  WHERE (id=1)", rand.Intn(100)).Error
}

func testPrepare(db *gorm.DB) (err error) {
	stmt, err := db.CommonDB().Prepare("UPDATE `config_ref` SET `data_id` = ?  WHERE (id=1)")
	if err != nil {
		return
	}
	_, err = stmt.Exec(rand.Intn(100))
	return
}

var testFuncList = []func(db *gorm.DB) (err error){
	testPrepare, testRow, testExec, testQuery, testUpdate,
}

func testServer(t *testing.T, g isql.GormSQL) {
	var err error
	e := gin.New()
	e.Use(gin.Logger())
	e.GET("/", func(c *gin.Context) {
		ctx := c.Request.Context()
		if err = g.GetDB(ctx).Transaction(testUpdate); err != nil {
			panic(err)
		}
		go func(ctx context.Context) {
			_ = g.WithTransaction(ctx, testUpdate)
		}(ctx)

		ctx = kcontext.Detach(ctx)
		go func() {
			db := g.GetDB(ctx)
			for _, f := range testFuncList {
				if err = f(db); err != nil {
					panic(err)
				}
			}
			_ = testUpdateError(db)

			for _, f := range testFuncList {
				if err = g.WithTransaction(ctx, f); err != nil {
					panic(err)
				}
			}
			_ = g.WithTransaction(ctx, testUpdateError)
		}()
	})
	go func() {
		pprof.RouteRegister(e.Group("/"))
		_ = e.Run(":80")
	}()
	http.DefaultClient.Transport = ihttp.NewDefaultTransport()
	wg := new(sync.WaitGroup)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := http.Get("http://localhost")
			if err != nil {
				panic(err)
			}
			_ = r.Body.Close()
		}()
		time.Sleep(time.Millisecond)
	}
	wg.Wait()
}

func testTx(t *testing.T, g isql.GormSQL) {
	ctx := context.Background()
	db := g.GetDB(ctx)
	var ret int
	err := db.Table("config_ref").Select("data_id").Where("id=1").Row().Scan(&ret)
	if err != nil {
		t.Fatal(err)
	}

	err = g.WithTransaction(ctx, func(tx *gorm.DB) (err error) {
		if err = tx.Table("config_ref").Where("id=1").
			Updates(map[string]interface{}{
				"data_id": ret + 1,
			}).Error; err != nil {
			return
		}
		return testUpdateError(tx)
	})

	func() {
		p := "test panic"
		defer func() {
			assert.Equal(t, recover(), p)
		}()
		err = g.WithTransaction(ctx, func(tx *gorm.DB) (err error) {
			if err = tx.Table("config_ref").Where("id=1").
				Updates(map[string]interface{}{
					"data_id": ret + 1,
				}).Error; err != nil {
				return
			}
			panic(p)
		})
	}()

	if err == nil {
		t.Fatal("should be nil")
	}

	var ret2 int
	if err = db.Table("config_ref").Select("data_id").Where("id=1").Row().Scan(&ret2); err != nil {
		t.Fatal(err)
	}
	if ret != ret2 {
		assert.Equal(t, ret, ret2)
	}
}
