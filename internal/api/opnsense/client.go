package opnsense

import (
	"opnsense-auto-dns/internal/api"
)

type Client struct {
	*api.Client
	Unbound *UnboundService
}

func NewClient(host, apiKey, apiSecret string, ignoreCert bool) *Client {
	baseClient := api.NewClient(host, apiKey, apiSecret, ignoreCert)

	client := &Client{
		Client: baseClient,
	}

	client.Unbound = NewUnboundService(client)
	return client
}
