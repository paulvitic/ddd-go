package http

import (
	"github.com/paulvitic/ddd-go"
	"github.com/stretchr/testify/assert"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockEndpoint struct {
	path    string
	methods []string
	handler func(w http.ResponseWriter, r *http.Request)
}

func (m *mockEndpoint) Path() string {
	return m.path
}

func (m *mockEndpoint) Methods() []string {
	return m.methods
}

func (m *mockEndpoint) Handler() func(w http.ResponseWriter, r *http.Request) {
	return m.handler
}

func (m *mockEndpoint) WithCommandTranslator(translator CommandTranslator) Endpoint {
	return nil
}

func (m *mockEndpoint) WithQueryTranslator(translator QueryTranslator) Endpoint {
	return nil
}

func (m *mockEndpoint) RegisterCommandBus(bus go_ddd.CommandBus) {
	return
}

func (m *mockEndpoint) RegisterQueryBus(bus go_ddd.QueryBus) {
	return
}

func TestServer(t *testing.T) {
	s := NewServer(":8080")

	endpoint := &mockEndpoint{
		path:    "test",
		methods: []string{"GET"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("test passed"))
			if err != nil {
				t.Error(err)
			}
		},
	}

	s.RegisterEndpoint(endpoint)

	req, err := http.NewRequest("GET", "test", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(endpoint.Handler())

	handler.ServeHTTP(rr, req)

	status := rr.Code
	assert.Equal(t, http.StatusOK, status)

	expected := `test passed`
	assert.Equal(t, expected, rr.Body.String())
}

func TestServerStartAndStop(t *testing.T) {
	s := NewServer(":8080")
	s.Start()

	// Check if the server is running
	conn, err := net.Dial("tcp", "localhost:8080")
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	// Close the connection
	err = conn.Close()
	assert.NoError(t, err)

	// Stop the server
	err = s.Stop()
	assert.NoError(t, err)

	// Give the server some time to stop
	time.Sleep(1 * time.Second)

	// Check if the server has stopped
	conn, err = net.Dial("tcp", "localhost:8080")
	assert.Error(t, err)
	assert.Nil(t, conn)
}
