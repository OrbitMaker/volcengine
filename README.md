VolcEngine DNS for [`libdns`](https://github.com/libdns/libdns)
=======================
[![Go Reference](https://pkg.go.dev/badge/github.com/libdns/volcengine.svg)](https://pkg.go.dev/github.com/libdns/volcengine)

This package implements the [libdns interfaces](https://github.com/libdns/libdns), allowing you to manage DNS records with the [VolcEngine DNS API](https://api.volcengine.com/api-docs/view/overview?serviceCode=dns&version=2018-08-01).

The official VolcEngine Go SDK is available [here](https://github.com/volcengine/volcengine-go-sdk).

## Authenticating

To authenticate you need to supply your `AccessKeyID` and `AccessKeySecret` to the Provider. You can obtain these credentials from the [VolcEngine Console](https://console.volcengine.com/).

The credentials can be provided in two ways:
1. **Environment variables**: Set `ACCESS_KEY_ID` and `ACCESS_KEY_SECRET` environment variables
2. **Direct configuration**: Set `AccessKeyID` and `AccessKeySecret` in the Provider struct

## Domain Identification

VolcEngine DNS API uses `ZID` (Zone ID) to identify domains. This provider supports three ways to specify a domain, with the following priority order:

1. **ZID (highest priority)**: Set `Provider.ZID` directly with the domain ID (e.g., `304092`)
2. **Zone field**: Set `Provider.Zone` with the domain name (e.g., `example.com`), and the provider will automatically look up the corresponding ZID using the `ListZones` API
3. **zone parameter**: Pass the domain name as the `zone` parameter in method calls (e.g., `GetRecords(ctx, "example.com")`)

**Priority**: `Provider.ZID` > `Provider.Zone` > `zone` parameter

## Example

Here's a minimal example of how to get all your DNS records using this `libdns` provider:

```go
package main

import (
	"context"
	"fmt"
	"github.com/libdns/volcengine"
)

func main() {
	provider := volcengine.Provider{
		CredentialInfo: volcengine.CredentialInfo{
			AccessKeyID:     "<AccessKeyID from your VolcEngine console>",
			AccessKeySecret: "<AccessKeySecret from your VolcEngine console>",
		},
		// Option 1: Set ZID directly if you know it (highest priority)
		// ZID: 304092,
		// Option 2: Set Zone domain name, will auto-lookup ZID via ListZones API
		// Zone: "example.com",
	}

	// GetRecords: if Provider.ZID or Provider.Zone is set, zone parameter can be empty
	// Otherwise, pass domain name as zone parameter
	records, err := provider.GetRecords(context.TODO(), "example.com")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, record := range records {
		tmp := record.RR()
		fmt.Printf("%s %v %s %s\n", tmp.Name, tmp.TTL.Seconds(), tmp.Type, tmp.Data)
	}
}
```

For complete demo check [_demo/demo.go](_demo/demo.go)

## ACME DNS-01 Support

This provider **fully supports ACME DNS-01 challenges** for Let's Encrypt and other ACME certificate providers.

### What is ACME DNS-01 Challenge?

ACME DNS-01 challenge is a method used by Let's Encrypt and other ACME providers to verify domain ownership before issuing SSL/TLS certificates. The process works as follows:

1. **Challenge Request**: When you request a certificate, the ACME server provides a challenge token
2. **DNS Record Creation**: Your ACME client (e.g., Caddy, cert-manager, Traefik) creates a TXT record:
   - **Name**: `_acme-challenge.<your-domain>`
   - **Value**: The challenge token provided by ACME server
3. **Verification**: The ACME server queries this TXT record to verify you control the domain
4. **Cleanup**: After verification, the challenge record is deleted

### How libdns Enables ACME DNS-01 Support

The `libdns` library provides a standardized interface for DNS record management. ACME clients use `libdns` providers to:
- **Add** challenge TXT records (`AppendRecords`)
- **Delete** challenge TXT records (`DeleteRecords`)

Since this provider implements all required `libdns` interfaces, it automatically works with any ACME client that supports `libdns`.

### Supported ACME Clients

This provider can be used with any ACME client that supports `libdns`, including:

- **[Caddy](https://caddyserver.com/)** - Automatic HTTPS with DNS-01 challenge
- **[cert-manager](https://cert-manager.io/)** - Kubernetes certificate management
- **[Traefik](https://traefik.io/)** - Cloud-native reverse proxy
- **[acme.sh](https://github.com/acmesh-official/acme.sh)** - Shell script for ACME
- **[lego](https://github.com/go-acme/lego)** - Let's Encrypt client and ACME library

### Example: Using with Caddy

Configure Caddy to use this provider for DNS-01 challenges:

**Option 1: Using ZID (recommended for better performance)**

```json
{
  "apps": {
    "tls": {
      "challenges": {
        "dns": {
          "provider": {
            "name": "volcengine",
            "access_key_id": "your_access_key_id",
            "access_key_secret": "your_access_key_secret",
            "zid": 304092
          }
        }
      }
    }
  }
}
```

**Option 2: Using Zone (domain name)**

```json
{
  "apps": {
    "tls": {
      "challenges": {
        "dns": {
          "provider": {
            "name": "volcengine",
            "access_key_id": "your_access_key_id",
            "access_key_secret": "your_access_key_secret",
            "zone": "example.com"
          }
        }
      }
    }
  }
}
```

**Note**: You can use either `zid` or `zone` (not both). If both are provided, `zid` takes priority. The `zone` option will automatically look up the corresponding ZID using the `ListZones` API.

### Implementation Details

This provider implements the following `libdns` interfaces required for ACME DNS-01 support:

- ✅ **RecordGetter** (`GetRecords`) - Query existing records
- ✅ **RecordAppender** (`AppendRecords`) - Create challenge TXT records
- ✅ **RecordDeleter** (`DeleteRecords`) - Remove challenge TXT records after verification
- ✅ **RecordSetter** (`SetRecords`) - Update records if needed

The provider automatically:
- Extracts the root domain from challenge record names (e.g., `_acme-challenge.example.com` → `example.com`)
- Handles both ZID and domain name identification
- Supports all DNS record types including TXT records required for ACME challenges
