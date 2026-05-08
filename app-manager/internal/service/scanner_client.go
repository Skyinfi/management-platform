package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Skyinfi/management-platform/app-manager/internal/model"
)

type ScannerClient struct {
	addr   string
	client *http.Client
}

func NewScannerClient(addr string) *ScannerClient {
	return &ScannerClient{
		addr: strings.TrimRight(addr, "/"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type scannerAPIResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type scannerApp struct {
	PID       int       `json:"pid"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CmdLine   string    `json:"cmdLine"`
	WorkDir   string    `json:"workDir"`
	ExePath   string    `json:"exePath"`
	Ports     []int     `json:"ports"`
	User      string    `json:"user"`
	StartTime time.Time `json:"startTime"`
	Managed   bool      `json:"managed"`
	Status    string    `json:"status"`
	InDocker  bool      `json:"inDocker"`
}

type scanResult struct {
	Count int           `json:"count"`
	Apps  []scannerApp  `json:"apps"`
}

type appListResult struct {
	Items []scannerApp `json:"items"`
}

type watchResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (c *ScannerClient) RunScan(ctx context.Context) ([]model.DiscoveredProcess, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.addr+"/api/scanner/run", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	body, err := c.doJSON(req)
	if err != nil {
		return nil, err
	}

	var result scanResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode scan result: %w", err)
	}

	return convertApps(result.Apps), nil
}

func (c *ScannerClient) GetApps(ctx context.Context) ([]model.DiscoveredProcess, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.addr+"/api/scanner/apps", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	body, err := c.doJSON(req)
	if err != nil {
		return nil, err
	}

	var result appListResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode app list: %w", err)
	}

	return convertApps(result.Items), nil
}

func (c *ScannerClient) WatchApp(ctx context.Context, pid int) (bool, string, error) {
	url := fmt.Sprintf("%s/api/scanner/apps/%d/watch", c.addr, pid)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return false, "", fmt.Errorf("create request: %w", err)
	}

	body, err := c.doJSON(req)
	if err != nil {
		return false, "", err
	}

	var result watchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return false, "", fmt.Errorf("decode watch result: %w", err)
	}

	return result.Success, result.Message, nil
}

func (c *ScannerClient) doJSON(req *http.Request) (json.RawMessage, error) {
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("scanner request: %w", err)
	}
	defer resp.Body.Close()

	var apiResp scannerAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if resp.StatusCode != http.StatusOK || apiResp.Code != 0 {
		return nil, fmt.Errorf("scanner error: status=%d code=%d message=%s",
			resp.StatusCode, apiResp.Code, apiResp.Message)
	}

	return apiResp.Data, nil
}

func convertApps(apps []scannerApp) []model.DiscoveredProcess {
	result := make([]model.DiscoveredProcess, 0, len(apps))
	for _, app := range apps {
		endpoint := formatEndpoint(app.Ports)
		result = append(result, model.DiscoveredProcess{
			ID:        discoveryID(app.PID, endpoint),
			Name:      app.Name,
			PID:       app.PID,
			User:      app.User,
			Endpoint:  endpoint,
			Protocol:  "tcp",
			Command:   app.CmdLine,
			ExePath:   app.ExePath,
			Cwd:       app.WorkDir,
			Managed:   app.Managed,
			InDocker:  app.InDocker,
			Adoptable: !app.Managed && !app.InDocker,
		})
	}
	return result
}

func formatEndpoint(ports []int) string {
	if len(ports) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(ports))
	for _, p := range ports {
		parts = append(parts, fmt.Sprintf(":%d", p))
	}
	return strings.Join(parts, ", ")
}

func discoveryID(pid int, endpoint string) string {
	sum := sha1.Sum([]byte(fmt.Sprintf("%d|%s", pid, endpoint)))
	return hex.EncodeToString(sum[:8])
}
