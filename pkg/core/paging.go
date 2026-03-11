package core

import (
	"fmt"
	"math"
)

const PagingLimitDefault = 20

type Paging struct {
	List      interface{} `json:"list"`
	Page      int64       `json:"page"`
	Limit     int64       `json:"limit"`
	TotalPage int64       `json:"total_page"`
	Total     int64       `json:"total"`
}

func Offset(pageNo int64, limitNo int64) int64 {
	return (limitNo * pageNo) - limitNo
}

func Pagination(pageNo int64, limitNo int64, getCount func() int64, getData func(limit int64, offset int64) interface{}) Paging {
	fmt.Println("2 Pagination")
	total := getCount()
	var pageCount = math.Ceil(float64(total) / float64(limitNo))
	pageCountInt := int64(pageCount)
	if pageNo <= 0 {
		pageNo = 1
	}
	offset := (limitNo * pageNo) - limitNo

	data := getData(limitNo, offset)

	return Paging{
		List:      data,
		Page:      pageNo,
		Limit:     limitNo,
		TotalPage: pageCountInt,
		Total:     total,
	}
}
