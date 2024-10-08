package ddd

import (
	"context"
)

type QueryExecutor func(context.Context, Query) (QueryResponse, error)

type QueryService interface {
	SubscribedTo() []string
	Executor() QueryExecutor
}

type queryService struct {
	subscribedTo []string
	executor     QueryExecutor
}

func NewQueryService(executor QueryExecutor, subscribedTo ...interface{}) QueryService {
	var qt []string
	for _, q := range subscribedTo {
		qt = append(qt, NewQuery(q).Type())
	}
	return &queryService{
		subscribedTo: qt,
		executor:     executor,
	}
}

func (p *queryService) SubscribedTo() []string {
	return p.subscribedTo
}

func (p *queryService) Executor() QueryExecutor {
	return p.executor
}
