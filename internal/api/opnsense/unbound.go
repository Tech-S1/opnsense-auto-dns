package opnsense

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"opnsense-auto-dns/internal/logger"

	"github.com/go-resty/resty/v2"
)

type UnboundService struct {
	client *Client
}

func NewUnboundService(client *Client) *UnboundService {
	return &UnboundService{
		client: client,
	}
}

func (s *UnboundService) makeAPIRequest(method, endpoint string, payload any) ([]byte, error) {
	url := fmt.Sprintf("https://%s%s", s.client.GetHost(), endpoint)
	logger.Debug("Making API request", "method", method, "url", url)

	req := s.client.GetRestyClient().R().
		SetHeader("Authorization", s.client.GetAuthHeader())

	if payload != nil {
		req.SetHeader("Content-Type", "application/json").
			SetBody(payload)
	}

	var resp *resty.Response
	var err error

	switch method {
	case "GET":
		resp, err = req.Get(url)
	case "POST":
		resp, err = req.Post(url)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return nil, err
	}

	body := resp.Body()
	logger.Debug("Received API response", "status", resp.StatusCode(), "body_length", len(body))

	if resp.StatusCode() != http.StatusOK {
		return body, fmt.Errorf("API request failed, status: %d, response: %s", resp.StatusCode(), string(body))
	}

	return body, nil
}

func (s *UnboundService) parseAPIResponse(body []byte, operation string) error {
	var apiResponse Response
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		logger.Error("Failed to parse API response", "error", err, "response_body", string(body), "operation", operation)
		return fmt.Errorf("failed to parse API response: %v", err)
	}

	if apiResponse.Result == "failed" {
		logger.Error("API operation failed", "result", apiResponse.Result, "response", string(body), "operation", operation)
		return fmt.Errorf("API operation failed: %s", string(body))
	}

	return nil
}

func (s *UnboundService) createHostPayload(hostname, domain, ip string, enabled string) map[string]any {
	payload := map[string]any{
		"host": map[string]any{
			"hostname":    hostname,
			"domain":      domain,
			"rr":          "A",
			"server":      ip,
			"description": fmt.Sprintf("Auto-updated by opnsense-auto-dns at %s", time.Now().Format("2006-01-02 15:04:05")),
		},
	}

	if enabled != "" {
		payload["host"].(map[string]any)["enabled"] = enabled
	}

	return payload
}

func (s *UnboundService) GetExistingDNSRecord(hostname, domain string) (*HostOverride, error) {
	logger.Info("Searching for existing DNS record", "hostname", hostname, "domain", domain)

	body, err := s.makeAPIRequest("GET", "/api/unbound/settings/search_host_override", nil)
	if err != nil {
		logger.Error("Failed to fetch existing DNS records", "error", err, "hostname", hostname, "domain", domain)
		return nil, fmt.Errorf("failed to fetch existing DNS records: %v", err)
	}

	var searchResponse SearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		logger.Error("Failed to parse DNS records response", "error", err, "response_body", string(body))
		return nil, fmt.Errorf("failed to parse DNS records response: %v", err)
	}

	logger.Debug("Parsed search response", "status", searchResponse.Status, "record_count", len(searchResponse.Rows))

	for _, record := range searchResponse.Rows {
		if strings.EqualFold(record.Hostname, hostname) && strings.EqualFold(record.Domain, domain) {
			logger.Info("Found existing DNS record", "uuid", record.UUID, "hostname", record.Hostname, "domain", record.Domain, "server", record.Server)
			return &record, nil
		}
	}

	logger.Info("No existing DNS record found", "hostname", hostname, "domain", domain)
	return nil, nil
}

func (s *UnboundService) UpdateDNSRecord(record *HostOverride, hostname, domain, ip string) error {
	logger.Info("Updating existing DNS record", "uuid", record.UUID, "hostname", hostname, "domain", domain, "ip", ip)

	endpoint := fmt.Sprintf("/api/unbound/settings/setHostOverride/%s", record.UUID)
	payload := s.createHostPayload(hostname, domain, ip, "1")

	logger.Debug("Request payload", "payload", payload)

	body, err := s.makeAPIRequest("POST", endpoint, payload)
	if err != nil {
		logger.Error("Failed to update DNS record", "error", err, "uuid", record.UUID, "hostname", hostname, "domain", domain, "ip", ip)
		return fmt.Errorf("error updating DNS: %v", err)
	}

	if err := s.parseAPIResponse(body, "update DNS record"); err != nil {
		return err
	}

	logger.Info("Successfully updated DNS record", "uuid", record.UUID, "hostname", hostname, "domain", domain, "ip", ip)

	if err := s.ReconfigureService(); err != nil {
		logger.Error("Failed to reconfigure unbound service after DNS update", "error", err)
		return fmt.Errorf("DNS record updated but failed to reconfigure service: %v", err)
	}

	return nil
}

func (s *UnboundService) CreateDNSRecord(hostname, domain, ip string) error {
	logger.Info("Creating new DNS record", "hostname", hostname, "domain", domain, "ip", ip)

	payload := s.createHostPayload(hostname, domain, ip, "")

	logger.Debug("Request payload", "payload", payload)

	body, err := s.makeAPIRequest("POST", "/api/unbound/settings/addHostOverride", payload)
	if err != nil {
		logger.Error("Failed to create DNS record", "error", err, "hostname", hostname, "domain", domain, "ip", ip)
		return fmt.Errorf("error creating DNS: %v", err)
	}

	if err := s.parseAPIResponse(body, "create DNS record"); err != nil {
		return err
	}

	logger.Info("Successfully created DNS record", "hostname", hostname, "domain", domain, "ip", ip)

	if err := s.ReconfigureService(); err != nil {
		logger.Error("Failed to reconfigure unbound service after DNS creation", "error", err)
		return fmt.Errorf("DNS record created but failed to reconfigure service: %v", err)
	}

	return nil
}

func (s *UnboundService) ReconfigureService() error {
	logger.Info("Reconfiguring unbound DNS service")

	body, err := s.makeAPIRequest("POST", "/api/unbound/service/reconfigure", map[string]any{})
	if err != nil {
		logger.Error("Failed to reconfigure unbound service", "error", err)
		return fmt.Errorf("failed to reconfigure unbound service: %v", err)
	}

	if err := s.parseAPIResponse(body, "reconfigure service"); err != nil {
		return err
	}

	logger.Info("Successfully reconfigured unbound DNS service")
	return nil
}
