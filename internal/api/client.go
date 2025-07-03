package api

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"

	"opnsense-auto-dns/internal/logger"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	host      string
	apiKey    string
	apiSecret string
	resty     *resty.Client
}

func NewClient(host, apiKey, apiSecret string, ignoreCert bool) *Client {
	logger.Info("Creating new API client", "host", host, "ignoreCert", ignoreCert)

	restyClient := resty.New()

	if ignoreCert {
		logger.Warn("TLS certificate verification disabled")
		restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return &Client{
		host:      host,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		resty:     restyClient,
	}
}

func (c *Client) getAuthHeader() string {
	credentials := fmt.Sprintf("%s:%s", c.apiKey, c.apiSecret)
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	return fmt.Sprintf("Basic %s", encodedCredentials)
}

func (c *Client) GetRestyClient() *resty.Client {
	return c.resty
}

func (c *Client) GetHost() string {
	return c.host
}

func (c *Client) GetAuthHeader() string {
	return c.getAuthHeader()
}
