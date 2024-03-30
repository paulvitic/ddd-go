package application

import (
	"context"
	ddd "github.com/paulvitic/ddd-go"
	"github.com/paulvitic/ddd-go/example/hotel/domain"
)

type guestsService struct {
	ddd.QueryService
}

func GuestsQueryExecutor(view domain.Guests) ddd.QueryExecutor {
	return func(ctx context.Context, query ddd.Query) (ddd.QueryResponse, error) {
		switch query.Type() {
		case ddd.QueryType(domain.GuestInRoom{}):
			guests := view.GuestInRoom(query.Filter().(domain.GuestInRoom).Number)
			return ddd.NewQueryResponse(guests), nil
		}
		return nil, nil
	}
}

func GuestsService(view domain.Guests) ddd.QueryService {
	return &guestsService{ddd.NewQueryService(GuestsQueryExecutor(view))}
}
