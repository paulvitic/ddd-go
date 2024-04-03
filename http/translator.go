package http

import (
	ddd "github.com/paulvitic/ddd-go"
	"net/http"
)

type CommandTranslator func(from *http.Request) (ddd.Command, error)

type QueryTranslator func(from *http.Request) (ddd.Query, error)
