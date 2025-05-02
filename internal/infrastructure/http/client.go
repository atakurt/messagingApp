package http

import (
	"io"
	"net/http"

	"github.com/atakurt/messagingApp/internal/infrastructure/config"
)

type Client interface {
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}

type HttpClient struct {
	client *http.Client
}

func NewHttpClient() *HttpClient {
	return &HttpClient{
		client: &http.Client{
			Timeout: config.Cfg.Http.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:          config.Cfg.Http.MaxIdleConns,
				IdleConnTimeout:       config.Cfg.Http.IdleConnTimeout,
				TLSHandshakeTimeout:   config.Cfg.Http.TlsHandshakeTimeout,
				ExpectContinueTimeout: config.Cfg.Http.ExpectContinueTimeout,
			},
		},
	}
}

func (c *HttpClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	return c.client.Post(url, contentType, body)
}
