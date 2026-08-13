# IPWho ([ipwho.org](https://www.ipwho.org)) Go SDK

[![Go version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://go.dev/) [![license](https://img.shields.io/badge/license-MIT-green.svg)](https://github.com/lavrox/SDK-IPWho-Go/blob/main/LICENSE)

Official Go client for the [IPWho](https://www.ipwho.org) IP Intelligence API. One call returns the **full** payload: geolocation, timezone, flag, currency, connection (ASN/ISP), security, and user-agent when present.

There is no extra registry: `go get` uses the GitHub module.

- Website: [https://www.ipwho.org](https://www.ipwho.org)
- API docs: [https://www.ipwho.org/docs](https://www.ipwho.org/docs)
- Get an API key: [https://www.ipwho.org](https://www.ipwho.org)
- Live API: [https://api.ipwho.org](https://api.ipwho.org)
- Source: [https://github.com/lavrox/SDK-IPWho-Go](https://github.com/lavrox/SDK-IPWho-Go)

## Installation

```bash
go get github.com/lavrox/SDK-IPWho-Go
```

Go 1.21+. Stdlib only.

## Quick Start

```go
client, err := ipwho.NewClient(os.Getenv("IPWHO_API_KEY"))
resp, err := client.Lookup("8.8.8.8", nil)              // GET /ip/{ip}
me, err := client.Me(nil)                               // GET /me
bulk, err := client.Bulk([]string{"8.8.8.8", "1.1.1.1"}, nil) // []IpGeoResponse
```

Successful JSON:

```
IpGeoResponse
├── Success bool
├── Message *string
└── Data *GeoData
    ├── IP
    ├── GeoLocation *GeoLocation
    ├── Timezone *Timezone
    ├── Flag *Flag
    ├── Currency *Currency
    ├── Connection *Connection
    ├── Security *Security
    └── UserAgent *UserAgent
```

Optional JSON fields are **pointers**. Check before dereference.

## Reading the full response (8.8.8.8)

Live [IPWho](https://www.ipwho.org) values: United States, ASN 15169, America/Chicago, dial code +1.

```go
resp, err := client.Lookup("8.8.8.8", nil)
if err != nil {
	log.Fatal(err)
}
data := resp.Data
fmt.Println(data.IP) // 8.8.8.8

geo := data.GeoLocation
fmt.Println(*geo.Country, *geo.CountryCode) // United States, US
fmt.Println(*geo.Continent, *geo.ContinentCode)
if geo.DialCode != nil {
	fmt.Println(*geo.DialCode) // +1
}
if geo.IsInEu != nil {
	fmt.Println(*geo.IsInEu)
}
fmt.Println(geo.Latitude, geo.Longitude, geo.AccuracyRadius)

tz := data.Timezone
fmt.Println(*tz.TimeZone) // America/Chicago
fmt.Println(tz.Abbr, tz.Offset, tz.IsDst, tz.Utc, tz.CurrentTime)

fmt.Println(*data.Flag.FlagIcon)    // 🇺🇸
fmt.Println(*data.Flag.FlagUnicode) // U+1F1FA U+1F1F8

fmt.Println(*data.Currency.Code, *data.Currency.NamePlural) // USD, US dollars

conn := data.Connection
fmt.Println(*conn.AsnNumber) // 15169
fmt.Println(*conn.AsnOrg)    // Google LLC
fmt.Println(conn.Isp, conn.Org, conn.Domain, conn.ConnectionType)

fmt.Println(data.Security.IsVpn, data.Security.IsTor, data.Security.IsThreat)

if data.UserAgent != nil && data.UserAgent.Browser != nil {
	fmt.Println(*data.UserAgent.Browser.Name)
}

me, err := client.Me(nil)
fmt.Println(me.Data.IP)

bulk, err := client.Bulk([]string{"8.8.8.8", "1.1.1.1"}, nil)
for _, item := range bulk {
	fmt.Println(item.Data.IP, *item.Data.GeoLocation.Country)
}
```

`LookupOptions`: `Format` (`json`/`xml`/`csv`), `Fields` (comma-separated filter). Pass `nil` for defaults. Bulk is JSON only and returns `[]IpGeoResponse` in request order.

## API Reference

### `NewClient(apiKey string, opts ...ClientOption) (*Client, error)`

`WithBaseURL`, `WithHTTPClient`. Empty key returns an error. Query param: `apiKey`.

### `Lookup(ip string, opts *LookupOptions) (*IpGeoResponse, error)`

### `Me(opts *LookupOptions) (*IpGeoResponse, error)`

### `Bulk(ips []string, opts *LookupOptions) ([]IpGeoResponse, error)`

### Errors

`*IpWhoError` with `StatusCode` and `Message`.

## Type Definitions

```go
type GeoLocation struct {
	Continent, ContinentCode, Country, CountryCode, Capital *string
	Region, RegionCode, City, PostalCode, DialCode          *string
	IsInEu                                                  *bool
	Latitude, Longitude, AccuracyRadius                     *float64
}

type Timezone struct {
	TimeZone, Abbr, Utc, CurrentTime *string
	Offset                           *float64
	IsDst                            *bool
}

type Flag struct{ FlagIcon, FlagUnicode *string }

type Currency struct {
	Code, Symbol, Name, NamePlural, HexUnicode *string
}

type Connection struct {
	AsnNumber                                *float64
	AsnOrg, Isp, Org, Domain, ConnectionType *string
}

type Security struct {
	IsVpn, IsTor *bool
	IsThreat     *string // low | medium | high
}
```

JSON tags match the live wire (`postal_Code`, `flag_Icon`, `isVpn`, `asn_number`, …).

## Troubleshooting

- Empty key: `ipwho: apiKey is required`. Key: [https://www.ipwho.org](https://www.ipwho.org).
- HTTP 403: SDK sends `ipwho-go-sdk/1.0.0`.
- Nil pointers: city and user-agent are often unset on anycast IPs.

## Testing

```bash
IPWHO_API_KEY=your_key go run ./cmd/check
```

The live check lives in `cmd/check/test_ipwho.go` (it cannot sit next to `ipwho.go` at the repo root — that would mix `package main` and `package ipwho` in one directory).

## Changelog

### v1.0.0

- `Lookup`, `Me`, `Bulk` matching [https://api.ipwho.org](https://api.ipwho.org)

## License

MIT License — see [LICENSE](LICENSE).

## Support

- Documentation: [https://www.ipwho.org/docs](https://www.ipwho.org/docs)
- Contact: [https://www.ipwho.org/contact](https://www.ipwho.org/contact)
- GitHub Issues: [https://github.com/lavrox/SDK-IPWho-Go/issues](https://github.com/lavrox/SDK-IPWho-Go/issues)
- Website: [https://www.ipwho.org](https://www.ipwho.org)

---

Product by [https://lavrox.com](https://lavrox.com)
