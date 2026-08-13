// go.mod (inline)
// module github.com/lavrox/SDK-IPWho-Go
// go 1.21

// Package ipwho provides a Go client for the IPWho IP Intelligence API.
// API docs: https://api.ipwho.org
package ipwho

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ────────────────────────────────────────────────────────────────────────────
// Version
// ────────────────────────────────────────────────────────────────────────────

const Version = "1.0.0"

// ────────────────────────────────────────────────────────────────────────────
// Domain models — mirrors OpenAPI components/schemas exactly.
// ────────────────────────────────────────────────────────────────────────────

// GeoLocation represents the geoLocation object.
type GeoLocation struct {
	Continent      *string  `json:"continent,omitempty"`
	ContinentCode  *string  `json:"continentCode,omitempty"`
	Country        *string  `json:"country,omitempty"`
	CountryCode    *string  `json:"countryCode,omitempty"`
	Capital        *string  `json:"capital,omitempty"`
	Region         *string  `json:"region,omitempty"`
	RegionCode     *string  `json:"regionCode,omitempty"`
	City           *string  `json:"city,omitempty"`
	PostalCode     *string  `json:"postal_Code,omitempty"`
	DialCode       *string  `json:"dial_code,omitempty"`
	IsInEu         *bool    `json:"is_in_eu,omitempty"`
	Latitude       *float64 `json:"latitude,omitempty"`
	Longitude      *float64 `json:"longitude,omitempty"`
	AccuracyRadius *float64 `json:"accuracy_radius,omitempty"`
}

// Timezone represents the timezone object.
type Timezone struct {
	TimeZone    *string  `json:"time_zone,omitempty"`
	Abbr        *string  `json:"abbr,omitempty"`
	Offset      *float64 `json:"offset,omitempty"`
	IsDst       *bool    `json:"is_dst,omitempty"`
	Utc         *string  `json:"utc,omitempty"`
	CurrentTime *string  `json:"current_time,omitempty"`
}

// Flag represents the flag object.
type Flag struct {
	FlagIcon    *string `json:"flag_Icon,omitempty"`
	FlagUnicode *string `json:"flag_unicode,omitempty"`
}

// Currency represents the currency object.
type Currency struct {
	Code        *string `json:"code,omitempty"`
	Symbol      *string `json:"symbol,omitempty"`
	Name        *string `json:"name,omitempty"`
	NamePlural  *string `json:"name_plural,omitempty"`
	HexUnicode  *string `json:"hex_unicode,omitempty"`
}

// Connection represents the connection object.
type Connection struct {
	AsnNumber      *float64 `json:"asn_number,omitempty"`
	AsnOrg         *string  `json:"asn_org,omitempty"`
	Isp            *string  `json:"isp,omitempty"`
	Org            *string  `json:"org,omitempty"`
	Domain         *string  `json:"domain,omitempty"`
	ConnectionType *string  `json:"connection_type,omitempty"`
}

// Security represents the security object.
type Security struct {
	IsVpn    *bool   `json:"isVpn,omitempty"`
	IsTor    *bool   `json:"isTor,omitempty"`
	IsThreat *string `json:"isThreat,omitempty"` // "low" | "medium" | "high"
}

// Browser represents the userAgent > browser sub-object.
type Browser struct {
	Name    *string `json:"name,omitempty"`
	Version *string `json:"version,omitempty"`
}

// Engine represents the userAgent > engine sub-object.
type Engine struct {
	Name    *string `json:"name,omitempty"`
	Version *string `json:"version,omitempty"`
}

// OS represents the userAgent > os sub-object.
type OS struct {
	Name    *string `json:"name,omitempty"`
	Version *string `json:"version,omitempty"`
}

// Device represents the userAgent > device sub-object.
type Device struct {
	Type   *string `json:"type,omitempty"`
	Vendor *string `json:"vendor,omitempty"`
	Model  *string `json:"model,omitempty"`
}

// CPU represents the userAgent > cpu sub-object.
type CPU struct {
	Architecture *string `json:"architecture,omitempty"`
}

// UserAgent represents the userAgent object.
type UserAgent struct {
	Browser *Browser `json:"browser,omitempty"`
	Engine  *Engine  `json:"engine,omitempty"`
	OS      *OS      `json:"os,omitempty"`
	Device  *Device  `json:"device,omitempty"`
	CPU     *CPU     `json:"cpu,omitempty"`
}

// GeoData is the ``data`` payload inside a successful IpGeoResponse.
type GeoData struct {
	IP          string       `json:"ip"`
	GeoLocation *GeoLocation `json:"geoLocation,omitempty"`
	Timezone    *Timezone    `json:"timezone,omitempty"`
	Flag        *Flag        `json:"flag,omitempty"`
	Currency    *Currency    `json:"currency,omitempty"`
	Connection  *Connection  `json:"connection,omitempty"`
	Security    *Security    `json:"security,omitempty"`
	UserAgent   *UserAgent   `json:"userAgent,omitempty"`
}

// IpGeoResponse is the top-level API response wrapper.
type IpGeoResponse struct {
	Success bool      `json:"success"`
	Data    *GeoData  `json:"data,omitempty"`
	Message *string   `json:"message,omitempty"`
}

// BulkResponseData wraps the responseArray for bulk lookups.
type BulkResponseData struct {
	ResponseArray []IpGeoResponse `json:"responseArray,omitempty"`
}

// BulkResponse is the top-level wrapper for /bulk/ endpoint responses.
type BulkResponse struct {
	Success bool              `json:"success"`
	Data    *BulkResponseData `json:"data,omitempty"`
}

// ErrorResponse represents an error payload returned on non-2xx responses.
type ErrorResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
}

// ────────────────────────────────────────────────────────────────────────────
// Errors
// ────────────────────────────────────────────────────────────────────────────

// IpWhoError is the base error type returned by the client.
type IpWhoError struct {
	StatusCode int
	Message    string
}

func (e *IpWhoError) Error() string {
	return fmt.Sprintf("ipwho: HTTP %d: %s", e.StatusCode, e.Message)
}

// NotFoundError is returned when the API returns 404.
type NotFoundError struct{ IpWhoError }

// RateLimitError is returned when the API returns 429.
type RateLimitError struct{ IpWhoError }

// ────────────────────────────────────────────────────────────────────────────
// Client
// ────────────────────────────────────────────────────────────────────────────

const defaultBaseURL = "https://api.ipwho.org"
const defaultTimeout = 30 * time.Second

// ClientOption allows functional configuration of the Client.
type ClientOption func(*Client)

// WithBaseURL overrides the default API base URL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(baseURL, "/")
	}
}

// WithHTTPClient overrides the default HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// LookupOptions holds optional parameters for Lookup and Me calls.
type LookupOptions struct {
	Format string // "json" (default), "xml", "csv"
	Fields string // Comma-separated list of fields to return
}

// Client is the IPWho API client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new IPWho API client.
// apiKey is required. Additional options can be provided via ClientOption
// functions such as WithBaseURL or WithHTTPClient.
func NewClient(apiKey string, opts ...ClientOption) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("ipwho: apiKey is required")
	}
	c := &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// Lookup returns geolocation data for a specific IPv4 or IPv6 address.
func (c *Client) Lookup(ip string, opts *LookupOptions) (*IpGeoResponse, error) {
	return c.doGet(fmt.Sprintf("/ip/%s", ip), opts)
}

// Me returns geolocation data for the caller's own IP address.
func (c *Client) Me(opts *LookupOptions) (*IpGeoResponse, error) {
	return c.doGet("/me", opts)
}

// Bulk performs geolocation lookups for multiple IP addresses in a single request.
// The returned slice contains one IpGeoResponse per IP address in the same order.
func (c *Client) Bulk(ips []string, opts *LookupOptions) ([]IpGeoResponse, error) {
	if len(ips) == 0 {
		return nil, errors.New("ipwho: ips must not be empty")
	}
	bulkParam := strings.Join(ips, ",")
	// For bulk, we use LookupOptions for query params (apiKey is added automatically)
	params := c.buildParams(opts)
	// Remove any "format" param that isn't json — bulk only returns JSON
	delete(params, "format")

	urlStr := fmt.Sprintf("%s/bulk/%s", c.baseURL, bulkParam)
	body, statusCode, err := c.doRequest(urlStr, params)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, mapHTTPError(statusCode, body)
	}

	var bulkResp BulkResponse
	if err := json.Unmarshal(body, &bulkResp); err != nil {
		return nil, &IpWhoError{
			StatusCode: statusCode,
			Message:    fmt.Sprintf("failed to parse bulk response: %v", err),
		}
	}
	if !bulkResp.Success {
		return nil, &IpWhoError{
			StatusCode: statusCode,
			Message:    "API returned success=false for bulk request",
		}
	}
	if bulkResp.Data == nil {
		return nil, &IpWhoError{
			StatusCode: statusCode,
			Message:    "bulk response data is nil",
		}
	}
	return bulkResp.Data.ResponseArray, nil
}

// ── internal helpers ────────────────────────────────────────────────────────

// doGet performs a GET request and unmarshals the JSON response into an IpGeoResponse.
func (c *Client) doGet(path string, opts *LookupOptions) (*IpGeoResponse, error) {
	params := c.buildParams(opts)
	urlStr := fmt.Sprintf("%s%s", c.baseURL, path)
	body, statusCode, err := c.doRequest(urlStr, params)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, mapHTTPError(statusCode, body)
	}

	// Non-JSON formats: return a minimal response with raw text
	if opts != nil && opts.Format != "" && opts.Format != "json" {
		return &IpGeoResponse{
			Success: true,
			Data:    &GeoData{IP: string(body)},
		}, nil
	}

	var resp IpGeoResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, &IpWhoError{
			StatusCode: statusCode,
			Message:    fmt.Sprintf("failed to parse response: %v", err),
		}
	}
	if !resp.Success {
		return nil, &IpWhoError{
			StatusCode: statusCode,
			Message:    "API returned success=false",
		}
	}
	return &resp, nil
}

// buildParams constructs the query parameters from LookupOptions.
func (c *Client) buildParams(opts *LookupOptions) url.Values {
	params := url.Values{}
	params.Set("apiKey", c.apiKey)
	if opts != nil {
		if opts.Format != "" && opts.Format != "json" {
			params.Set("format", opts.Format)
		}
		if opts.Fields != "" {
			params.Set("get", opts.Fields)
		}
	}
	return params
}

// doRequest executes the HTTP GET and reads the full response body.
func (c *Client) doRequest(urlStr string, params url.Values) ([]byte, int, error) {
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, 0, &IpWhoError{StatusCode: 0, Message: fmt.Sprintf("failed to create request: %v", err)}
	}
	req.URL.RawQuery = params.Encode()
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("ipwho-go-sdk/%s", Version))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, &IpWhoError{StatusCode: 0, Message: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, &IpWhoError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("failed to read response body: %v", err),
		}
	}
	return body, resp.StatusCode, nil
}

// mapHTTPError converts an HTTP status code to the appropriate error type.
func mapHTTPError(statusCode int, body []byte) error {
	msg := fmt.Sprintf("HTTP %d", statusCode)
	if len(body) > 0 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
			msg = errResp.Message
		} else {
			msg = strings.TrimSpace(string(body))
		}
	}
	switch statusCode {
	case http.StatusNotFound:
		return &NotFoundError{IpWhoError{StatusCode: statusCode, Message: msg}}
	case http.StatusTooManyRequests:
		return &RateLimitError{IpWhoError{StatusCode: statusCode, Message: msg}}
	default:
		return &IpWhoError{StatusCode: statusCode, Message: msg}
	}
}

// Ensure convenience constructors satisfy the error interface.
var _ error = (*IpWhoError)(nil)
var _ error = (*NotFoundError)(nil)
var _ error = (*RateLimitError)(nil)

// ────────────────────────────────────────────────────────────────────────────
// Example usage
// ────────────────────────────────────────────────────────────────────────────

/*
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lavrox/SDK-IPWho-Go"
)

func main() {
	apiKey := os.Getenv("IPWHO_API_KEY")
	if apiKey == "" || strings.HasPrefix(apiKey, "sk.xx") {
		log.Fatal("Set IPWHO_API_KEY to run the example.")
	}
	client, err := ipwho.NewClient(apiKey)
	if err != nil {
		log.Fatal(err)
	}

	// Single lookup
	resp, err := client.Lookup("8.8.8.8", nil)
	if err != nil {
		log.Fatal(err)
	}
	if resp != nil && resp.Data != nil && resp.Data.GeoLocation != nil {
		gl := resp.Data.GeoLocation
		fmt.Printf("IP: %s  |  %s, %s (%v, %v)\n",
			resp.Data.IP, strOr(gl.City), strOr(gl.Country),
			floatOr(gl.Latitude), floatOr(gl.Longitude))
	}
	if resp != nil && resp.Data != nil && resp.Data.Currency != nil {
		cu := resp.Data.Currency
		fmt.Printf("Currency: %s (%s)\n", strOr(cu.Code), strOr(cu.Symbol))
	}

	// Self lookup
	me, err := client.Me(nil)
	if err != nil {
		log.Fatal(err)
	}
	if me != nil && me.Data != nil {
		fmt.Printf("\nMy IP: %s\n", me.Data.IP)
	}

	// Bulk lookup
	bulk, err := client.Bulk([]string{"8.8.8.8", "1.1.1.1"}, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nBulk results: %d IPs\n", len(bulk))
	for _, b := range bulk {
		fmt.Printf("  - %s\n", b.Data.IP)
	}
}

func strOr(s *string) string {
	if s == nil { return "" }
	return *s
}
func floatOr(f *float64) float64 {
	if f == nil { return 0 }
	return *f
}
*/
