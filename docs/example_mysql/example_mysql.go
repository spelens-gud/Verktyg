package main

import (
	"context"
	"errors"
	"math/rand"
	"strconv"

	"github.com/jinzhu/gorm"

	"github.com/spelens-gud/Verktyg/implements/gormctx"
	"github.com/spelens-gud/Verktyg/interfaces/isql"
	"github.com/spelens-gud/Verktyg/kits/klog/logger"
)

var sqlConfig = isql.SQLConfig{
	Host:     "10.43.3.13",
	Port:     3306,
	DB:       "book",
	User:     "devroot",
	Password: "ruVNBoRpboq3dXQn",
}

var ctx = context.Background()

type Book struct {
	ID    int    `gorm:"column:id"`
	Title string `gorm:"column:title"`
}

func (Book) TableName() string { return "book" }

type Service struct {
	DB isql.GormSQL
}

func (s *Service) Serve(ctx context.Context) (err error) {
	id := rand.Intn(100000)

	if err = s.DB.Transaction(ctx, func(txCtx context.Context, tx *gorm.DB) (err error) {
		// 使用txCtx传播的事务 插入指定ID记录
		if err = s.NewBook(txCtx, id); err != nil {
			return
		}

		// 查找插入的记录
		book, err := GetBookByID(tx, id)
		if err != nil {
			logger.FromContext(ctx).Errorf("look up book after insert error: %v", err)
			return
		}

		logger.FromContext(ctx).Infof("%v", book)

		// 模拟抛出错误 事务回滚
		return errors.New("some error happened")
	}); err != nil {
		logger.FromContext(ctx).Errorf("tx error: %v", err)
	}

	// 查找插入的记录
	if _, err = GetBookByID(s.DB.GetDB(ctx), id); err != nil {
		logger.FromContext(ctx).Errorf("look up book after tx rollback error: %v", err)
		// 找不到记录 证明事务已经回滚成功 txCtx传播的事务有效
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}
	return
}

func GetBookByID(db *gorm.DB, id int) (book Book, err error) {
	err = db.Where("id = ?", id).First(&book).Error
	return
}

func (s *Service) NewBook(ctx context.Context, id int) (err error) {
	// 插入一个记录
	if err = s.DB.GetDB(ctx).Table("book").Create(&Book{
		ID:    id,
		Title: "test_book_" + strconv.Itoa(id),
	}).Error; err != nil {
		return
	}
	return
}

func main() {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()

	// 初始化连接
	client, err := gormctx.NewSimpleGormSql(sqlConfig)
	if err != nil {
		return
	}
	// nolint
	defer client.Close()

	// 初始化服务模块
	service := &Service{DB: client}
	// 服务
	if err = service.Serve(ctx); err != nil {
		return
	}
}
