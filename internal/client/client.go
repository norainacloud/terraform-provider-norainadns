// Copyright (c) Noraina
// SPDX-License-Identifier: MPL-2.0

// Package client provides the API client for NorainaDNS
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the NorainaDNS API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// ClientConfig contains configuration for creating a new client
type ClientConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// NewClient creates a new NorainaDNS API client
func NewClient(config ClientConfig) *Client {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// APIResponse is the standard API response wrapper
type APIResponse[T any] struct {
	Success bool      `json:"success"`
	Data    T         `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}

// APIError represents an API error
type APIError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// doRequest performs an HTTP request and decodes the response
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse as generic API response first to check for errors
	var apiResp APIResponse[json.RawMessage]
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success || apiResp.Error != nil {
		if apiResp.Error != nil {
			return apiResp.Error
		}
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse the actual data if result is provided
	if result != nil && apiResp.Data != nil {
		if err := json.Unmarshal(apiResp.Data, result); err != nil {
			return fmt.Errorf("failed to parse response data: %w", err)
		}
	}

	return nil
}

// =============================================================================
// Zone Types
// =============================================================================

// Zone represents a DNS zone
type Zone struct {
	ID            string  `json:"id"`
	TenantID      string  `json:"tenantId"`
	DomainName    string  `json:"domainName"`
	Description   *string `json:"description"`
	IsActive      bool    `json:"isActive"`
	SyncStatus    string  `json:"syncStatus"`
	LastSyncAt    *string `json:"lastSyncAt"`
	LastSyncError *string `json:"lastSyncError"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
}

// CreateZoneRequest is the request to create a zone
type CreateZoneRequest struct {
	DomainName  string `json:"domainName"`
	Description string `json:"description,omitempty"`
}

// UpdateZoneRequest is the request to update a zone
type UpdateZoneRequest struct {
	Description string `json:"description,omitempty"`
	IsActive    *bool  `json:"isActive,omitempty"`
}

// ZoneNameservers contains the nameserver information for a zone
type ZoneNameservers struct {
	Nameservers []string `json:"nameservers"`
	AWS         []string `json:"aws,omitempty"`
	GCP         []string `json:"gcp,omitempty"`
}

// =============================================================================
// Record Types
// =============================================================================

// Record represents a DNS record
type Record struct {
	ID            string        `json:"id"`
	ZoneID        string        `json:"zoneId"`
	Name          string        `json:"name"`
	RecordType    string        `json:"recordType"`
	TTL           int64         `json:"ttl"`
	IsDynamic     bool          `json:"isDynamic"`
	RoutingPolicy string        `json:"routingPolicy"`
	SyncStatus    string        `json:"syncStatus"`
	LastSyncAt    *string       `json:"lastSyncAt"`
	LastSyncError *string       `json:"lastSyncError"`
	Values        []RecordValue `json:"values"`
	CreatedAt     string        `json:"createdAt"`
	UpdatedAt     string        `json:"updatedAt"`
}

// RecordValue represents a value for a DNS record
type RecordValue struct {
	ID        string  `json:"id,omitempty"`
	Value     string  `json:"value"`
	Priority  int     `json:"priority"`
	Weight    int     `json:"weight,omitempty"`
	MonitorID *string `json:"monitorId,omitempty"`
}

// CreateRecordRequest is the request to create a record
type CreateRecordRequest struct {
	Name          string                     `json:"name"`
	RecordType    string                     `json:"recordType"`
	TTL           int64                      `json:"ttl,omitempty"`
	IsDynamic     bool                       `json:"isDynamic,omitempty"`
	RoutingPolicy string                     `json:"routingPolicy,omitempty"`
	Values        []CreateRecordValueRequest `json:"values"`
}

// CreateRecordValueRequest is the request to create a record value
type CreateRecordValueRequest struct {
	Value     string  `json:"value"`
	Priority  int     `json:"priority,omitempty"`
	Weight    int     `json:"weight,omitempty"`
	MonitorID *string `json:"monitorId,omitempty"`
}

// UpdateRecordRequest is the request to update a record
type UpdateRecordRequest struct {
	Name          string                     `json:"name,omitempty"`
	TTL           int64                      `json:"ttl,omitempty"`
	IsDynamic     *bool                      `json:"isDynamic,omitempty"`
	RoutingPolicy string                     `json:"routingPolicy,omitempty"`
	Values        []CreateRecordValueRequest `json:"values,omitempty"`
}

// =============================================================================
// Monitor Types
// =============================================================================

// Monitor represents a health monitor
type Monitor struct {
	ID                   string  `json:"id"`
	TenantID             string  `json:"tenantId"`
	Name                 string  `json:"name"`
	Description          *string `json:"description"`
	Protocol             string  `json:"protocol"`
	TargetHost           string  `json:"targetHost"`
	TargetPort           int     `json:"targetPort"`
	TargetPath           string  `json:"targetPath"`
	CheckIntervalSeconds int     `json:"checkIntervalSeconds"`
	FailureThreshold     int     `json:"failureThreshold"`
	TimeoutSeconds       int     `json:"timeoutSeconds"`
	CustomHostHeader     *string `json:"customHostHeader"`
	CustomSNI            *string `json:"customSni"`
	ForceIPAddress       *string `json:"forceIpAddress"`
	IsActive             bool    `json:"isActive"`
	CurrentStatus        string  `json:"currentStatus"`
	LastCheckAt          *string `json:"lastCheckAt"`
	CreatedAt            string  `json:"createdAt"`
	UpdatedAt            string  `json:"updatedAt"`
}

// CreateMonitorRequest is the request to create a monitor
type CreateMonitorRequest struct {
	Name                 string `json:"name"`
	Description          string `json:"description,omitempty"`
	Protocol             string `json:"protocol"`
	TargetHost           string `json:"targetHost"`
	TargetPort           int    `json:"targetPort,omitempty"`
	TargetPath           string `json:"targetPath,omitempty"`
	CheckIntervalSeconds int    `json:"checkIntervalSeconds,omitempty"`
	FailureThreshold     int    `json:"failureThreshold,omitempty"`
	TimeoutSeconds       int    `json:"timeoutSeconds,omitempty"`
	CustomHostHeader     string `json:"customHostHeader,omitempty"`
	CustomSNI            string `json:"customSni,omitempty"`
	ForceIPAddress       string `json:"forceIpAddress,omitempty"`
}

// UpdateMonitorRequest is the request to update a monitor
type UpdateMonitorRequest struct {
	Name                 string `json:"name,omitempty"`
	Description          string `json:"description,omitempty"`
	TargetHost           string `json:"targetHost,omitempty"`
	TargetPort           int    `json:"targetPort,omitempty"`
	TargetPath           string `json:"targetPath,omitempty"`
	CheckIntervalSeconds int    `json:"checkIntervalSeconds,omitempty"`
	FailureThreshold     int    `json:"failureThreshold,omitempty"`
	TimeoutSeconds       int    `json:"timeoutSeconds,omitempty"`
	CustomHostHeader     string `json:"customHostHeader,omitempty"`
	CustomSNI            string `json:"customSni,omitempty"`
	ForceIPAddress       string `json:"forceIpAddress,omitempty"`
}

// MonitorStatus contains the current status of a monitor
type MonitorStatus struct {
	Status       string  `json:"status"`
	StatusReason *string `json:"statusReason"`
	LastCheckAt  *string `json:"lastCheckAt"`
}

// =============================================================================
// Zone API Methods
// =============================================================================

// GetZones returns all zones
func (c *Client) GetZones(ctx context.Context) ([]Zone, error) {
	var zones []Zone
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/zones", nil, &zones)
	return zones, err
}

// GetZone returns a zone by ID
func (c *Client) GetZone(ctx context.Context, id string) (*Zone, error) {
	var zone Zone
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/zones/%s", id), nil, &zone)
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

// CreateZone creates a new zone
func (c *Client) CreateZone(ctx context.Context, req CreateZoneRequest) (*Zone, error) {
	var zone Zone
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/zones", req, &zone)
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

// UpdateZone updates a zone
func (c *Client) UpdateZone(ctx context.Context, id string, req UpdateZoneRequest) (*Zone, error) {
	var zone Zone
	err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/api/v1/zones/%s", id), req, &zone)
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

// DeleteZone deletes a zone
func (c *Client) DeleteZone(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/zones/%s", id), nil, nil)
}

// GetZoneNameservers returns the nameservers for a zone
func (c *Client) GetZoneNameservers(ctx context.Context, id string) (*ZoneNameservers, error) {
	var ns ZoneNameservers
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/zones/%s/nameservers", id), nil, &ns)
	if err != nil {
		return nil, err
	}
	return &ns, nil
}

// SyncZone triggers a sync for a zone
func (c *Client) SyncZone(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/v1/zones/%s/sync", id), nil, nil)
}

// =============================================================================
// Record API Methods
// =============================================================================

// GetRecords returns all records for a zone
func (c *Client) GetRecords(ctx context.Context, zoneID string) ([]Record, error) {
	var records []Record
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/zones/%s/records", zoneID), nil, &records)
	return records, err
}

// GetRecord returns a record by ID
func (c *Client) GetRecord(ctx context.Context, zoneID, recordID string) (*Record, error) {
	var record Record
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/zones/%s/records/%s", zoneID, recordID), nil, &record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// CreateRecord creates a new record
func (c *Client) CreateRecord(ctx context.Context, zoneID string, req CreateRecordRequest) (*Record, error) {
	var record Record
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/v1/zones/%s/records", zoneID), req, &record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// UpdateRecord updates a record
func (c *Client) UpdateRecord(ctx context.Context, zoneID, recordID string, req UpdateRecordRequest) (*Record, error) {
	var record Record
	err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/api/v1/zones/%s/records/%s", zoneID, recordID), req, &record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// DeleteRecord deletes a record
func (c *Client) DeleteRecord(ctx context.Context, zoneID, recordID string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/zones/%s/records/%s", zoneID, recordID), nil, nil)
}

// RetryRecordSync triggers a sync retry for a record
func (c *Client) RetryRecordSync(ctx context.Context, zoneID, recordID string) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/v1/zones/%s/records/%s/retry", zoneID, recordID), nil, nil)
}

// =============================================================================
// Monitor API Methods
// =============================================================================

// GetMonitors returns all monitors
func (c *Client) GetMonitors(ctx context.Context) ([]Monitor, error) {
	var monitors []Monitor
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/monitors", nil, &monitors)
	return monitors, err
}

// GetMonitor returns a monitor by ID
func (c *Client) GetMonitor(ctx context.Context, id string) (*Monitor, error) {
	var monitor Monitor
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/monitors/%s", id), nil, &monitor)
	if err != nil {
		return nil, err
	}
	return &monitor, nil
}

// CreateMonitor creates a new monitor
func (c *Client) CreateMonitor(ctx context.Context, req CreateMonitorRequest) (*Monitor, error) {
	var monitor Monitor
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/monitors", req, &monitor)
	if err != nil {
		return nil, err
	}
	return &monitor, nil
}

// UpdateMonitor updates a monitor
func (c *Client) UpdateMonitor(ctx context.Context, id string, req UpdateMonitorRequest) (*Monitor, error) {
	var monitor Monitor
	err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/api/v1/monitors/%s", id), req, &monitor)
	if err != nil {
		return nil, err
	}
	return &monitor, nil
}

// DeleteMonitor deletes a monitor
func (c *Client) DeleteMonitor(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/monitors/%s", id), nil, nil)
}

// GetMonitorStatus returns the current status of a monitor
func (c *Client) GetMonitorStatus(ctx context.Context, id string) (*MonitorStatus, error) {
	var status MonitorStatus
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/monitors/%s/status", id), nil, &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// ToggleMonitor activates or suspends a monitor
func (c *Client) ToggleMonitor(ctx context.Context, id string) (*Monitor, error) {
	var monitor Monitor
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/v1/monitors/%s/toggle", id), nil, &monitor)
	if err != nil {
		return nil, err
	}
	return &monitor, nil
}

// RetryMonitorSync triggers a sync retry for a monitor
func (c *Client) RetryMonitorSync(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/v1/monitors/%s/retry", id), nil, nil)
}
