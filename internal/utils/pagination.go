package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Limit  int    `json:"limit"`
	Page   int    `json:"page"`
	Sort   string `json:"sort"`
	Offset int    `json:"-"`
}

func GetPagination(c *gin.Context) Pagination {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	sort := c.DefaultQuery("sort", "created_at desc")

	if limit > 100 {
		limit = 100
	}
	if page < 1 {
		page = 1
	}

	return Pagination{
		Limit:  limit,
		Page:   page,
		Sort:   sort,
		Offset: (page - 1) * limit,
	}
}

type PaginatedMeta struct {
	TotalCount int64 `json:"total_count"`
	TotalPage  int   `json:"total_page"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
}

func SuccessPaginatedResponse(c *gin.Context, message string, totalCount int64, page int, limit int, data interface{}) {
	totalPage := int(totalCount) / limit
	if int(totalCount)%limit > 0 {
		totalPage++
	}

	meta := PaginatedMeta{
		TotalCount: totalCount,
		TotalPage:  totalPage,
		Page:       page,
		Limit:      limit,
	}

	GeneralResponse(c, 200, true, message, data, meta, nil)
}
