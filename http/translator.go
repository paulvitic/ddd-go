package http

import (
	"github.com/paulvitic/ddd-go"
	"net/http"
)

type CommandTranslator func(from *http.Request) (go_ddd.Command, error)

type QueryTranslator func(from *http.Request) (go_ddd.Query, error)
