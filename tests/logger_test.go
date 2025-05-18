package ddd

import (
	"testing"

	"github.com/paulvitic/ddd-go"
)

func TestLoggerInit(t *testing.T) {
	log := &ddd.Logger{}
	log.OnInit()
	log.Info("testing logger info")
	log.Warn("testing logger warning")
	log.Error("testing logger %s", "error")
}
