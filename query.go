package ddd

import (
	"math"
	"reflect"
)

type Query interface {
	Type() string
	Filter() interface{}
	PageSize() int
	PageIndex() int
}

type query struct {
	filter    interface{}
	pageIndex int
	pageSize  int
}

func NewQuery(filter interface{}) Query {
	return &query{filter, 0, 1}
}

func NewPagedQuery(filter interface{}, pageIndex int, pageSize int) Query {
	return &query{filter, pageIndex, pageSize}
}

func (c *query) Type() string {
	return reflect.TypeOf(c.Filter()).PkgPath() + "." + reflect.TypeOf(c.Filter()).Name()
}

func (c *query) Filter() interface{} {
	return c.filter
}

func (c *query) PageSize() int {
	return c.pageSize
}

func (c *query) PageIndex() int {
	return c.pageIndex
}

type QueryResponse interface {
	Items() interface{}
	TotalPages() int
	PageNumber() int
	HasPrev() bool
	Prev() int
	HasNext() bool
	Next() int
}

type queryResponse struct {
	items     interface{}
	count     int
	pageIndex int
	pageSize  int
}

func NewQueryResponse(items interface{}) QueryResponse {
	return &queryResponse{items, 1, 0, 1}
}

func NewPagedQueryResponse(items interface{}, count int, pageIndex int, pageSize int) QueryResponse {
	return &queryResponse{items, count, pageIndex, pageSize}
}

func (qr *queryResponse) Items() interface{} {
	return qr.items
}

func (qr *queryResponse) TotalPages() int {
	return int(math.Ceil(float64(qr.count) / float64(qr.pageSize)))
}

func (qr *queryResponse) PageNumber() int {
	return qr.pageIndex + 1
}

func (qr *queryResponse) HasPrev() bool {
	return qr.pageIndex > 0
}

func (qr *queryResponse) Prev() int {
	if qr.HasPrev() {
		return qr.PageNumber() - 1
	}
	return 0
}

func (qr *queryResponse) HasNext() bool {
	return (qr.PageNumber() * qr.pageSize) < qr.count
}

func (qr *queryResponse) Next() int {
	if qr.HasNext() {
		return qr.PageNumber() + 1
	}
	return 0
}

func QueryType(filter interface{}) string {
	return reflect.TypeOf(filter).PkgPath() + "." + reflect.TypeOf(filter).Name()
}
