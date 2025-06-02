package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/paulvitic/ddd-go"
)

type IdProviderEndpoint struct {
	ddd.Endpoint
	*ddd.EventBus
}

func NewIdProviderEndpoint(logger *ddd.Logger, router *mux.Router, eventBus *ddd.EventBus) *IdProviderEndpoint {
	paths := []string{"/users/idProvider"}
	return &IdProviderEndpoint{
		ddd.NewEndpoint(&UsersEndpoint{}, paths, logger, router),
		eventBus,
	}
}

func (t *IdProviderEndpoint) Post(w http.ResponseWriter, r *http.Request) {
	event := ToIdentityIdProviderEvent(r)
	if err := t.EventBus.Dispatch(event); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
