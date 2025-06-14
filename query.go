package ddd

import (
	"math"
)

type QueryFiler = func(ctx *Context) (QueryResponse, error)

type Query interface {
	Filter(ctx *Context) (QueryResponse, error)
	PageSize() int
	PageIndex() int
}

type query struct {
	filter    QueryFiler
	pageIndex int
	pageSize  int
}

func NewQuery(filter QueryFiler) Query {
	return &query{filter, 0, 1}
}

func NewPagedQuery(filter QueryFiler, pageIndex int, pageSize int) Query {
	return &query{filter, pageIndex, pageSize}
}

func (c *query) Filter(ctx *Context) (QueryResponse, error) {
	return c.filter(ctx)
}

func (c *query) PageSize() int {
	return c.pageSize
}

func (c *query) PageIndex() int {
	return c.pageIndex
}

type QueryResponse interface {
	Items() any
	TotalPages() int
	PageNumber() int
	HasPrev() bool
	Prev() int
	HasNext() bool
	Next() int
}

type queryResponse struct {
	items     any
	count     int
	pageIndex int
	pageSize  int
}

func NewQueryResponse(items any) QueryResponse {
	return &queryResponse{items, 1, 0, 1}
}

func NewPagedQueryResponse(items any, count int, pageIndex int, pageSize int) QueryResponse {
	return &queryResponse{items, count, pageIndex, pageSize}
}

func (qr *queryResponse) Items() any {
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
