package http

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/atakurt/messagingApp/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
)

func TestNewHttpClient(t *testing.T) {
	// given
	originalTimeout := config.Cfg.Http.Timeout
	originalMaxIdleConns := config.Cfg.Http.MaxIdleConns
	originalIdleConnTimeout := config.Cfg.Http.IdleConnTimeout
	originalTlsHandshakeTimeout := config.Cfg.Http.TlsHandshakeTimeout
	originalExpectContinueTimeout := config.Cfg.Http.ExpectContinueTimeout

	// Set test values
	config.Cfg.Http.Timeout = 10 * time.Second
	config.Cfg.Http.MaxIdleConns = 100
	config.Cfg.Http.IdleConnTimeout = 90 * time.Second
	config.Cfg.Http.TlsHandshakeTimeout = 10 * time.Second
	config.Cfg.Http.ExpectContinueTimeout = 1 * time.Second

	// Restore original config after test
	defer func() {
		config.Cfg.Http.Timeout = originalTimeout
		config.Cfg.Http.MaxIdleConns = originalMaxIdleConns
		config.Cfg.Http.IdleConnTimeout = originalIdleConnTimeout
		config.Cfg.Http.TlsHandshakeTimeout = originalTlsHandshakeTimeout
		config.Cfg.Http.ExpectContinueTimeout = originalExpectContinueTimeout
	}()

	// when
	client := NewHttpClient()

	// then
	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
	assert.Equal(t, config.Cfg.Http.Timeout, client.client.Timeout)

	transport, ok := client.client.Transport.(*http.Transport)
	assert.True(t, ok)
	assert.Equal(t, config.Cfg.Http.MaxIdleConns, transport.MaxIdleConns)
	assert.Equal(t, config.Cfg.Http.IdleConnTimeout, transport.IdleConnTimeout)
	assert.Equal(t, config.Cfg.Http.TlsHandshakeTimeout, transport.TLSHandshakeTimeout)
	assert.Equal(t, config.Cfg.Http.ExpectContinueTimeout, transport.ExpectContinueTimeout)
}

func TestHttpClient_Post(t *testing.T) {
	// given
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, []byte(`{"test":"data"}`), body)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	client := NewHttpClient()

	requestBody := bytes.NewBufferString(`{"test":"data"}`)

	// when
	resp, err := client.Post(server.URL, "application/json", requestBody)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"status":"success"}`), body)
}

func TestHttpClient_Post_Error(t *testing.T) {
	// given
	client := NewHttpClient()

	// when
	resp, err := client.Post("http://invalid-url-that-does-not-exist.example", "application/json", nil)

	// then
	assert.Error(t, err)
	assert.Nil(t, resp)
}
