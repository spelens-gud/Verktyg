package isql

import (
	"context"
	"fmt"
)

var (
	// 分页参默认值
	defaultPageSize = 15
	minPageSize     = 1
	maxPageSize     = 5000
	minPageNo       = 0
	maxPageNo       = 5000
)

type PageSQLParam struct {
	CountQuery SQLCommand
	DataQuery  SQLCommand
	Order      string
	PageNo     int
	PageSize   int
}

// GetPageResult 获取分页结果
func GetPageResult(ctx context.Context, param PageSQLParam, result interface{}) (count int, err error) {
	// 校验
	if param.CountQuery == nil || param.DataQuery == nil {
		return 0, fmt.Errorf("nil sql: %+v", param)
	}
	// 总数
	count, err = param.CountQuery.Count(ctx)
	if err != nil {
		return
	}
	if count == 0 {
		return
	}
	// Page 大于最大阈值抛error,防止使用者不知情导致业务问题
	var offset, size int
	if param.PageNo < minPageNo {
		param.PageNo = minPageNo
	}
	if param.PageNo > maxPageNo {
		err = fmt.Errorf("pageNo > maxPageNo, check maxPageNo")
		return
	}
	if param.PageSize < minPageSize {
		param.PageSize = defaultPageSize
	}
	if param.PageSize > maxPageSize {
		err = fmt.Errorf("pageSize > maxPageSize, check maxPageSize")
		return
	}
	size = param.PageSize
	offset = param.PageNo * param.PageSize

	// 数据
	err = param.DataQuery.Order(param.Order).Offset(offset).Limit(size).Scan(ctx, result)
	return
}
