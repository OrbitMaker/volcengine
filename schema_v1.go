package volcengine

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func defaultSchema(cred *CredentialInfo) (*volcClientSchema, error) {
	if cred == nil {
		return &volcClientSchema{}, errors.New("volcengine: credentials missing")
	}

	now := time.Now().UTC()
	dateStr := now.Format("20060102")
	timeStr := now.Format("150405")

	result := &volcClientSchema{
		APIHost: addressOfAPI,
		headerPairs: []keyPair{
			{Key: "X-Date", Value: dateStr + "T" + timeStr + "Z"},
		},
		requestBody:  make(map[string]interface{}),
		signPassword: cred.AccessKeySecret,
		credential:   cred,
	}

	return result, nil
}

// hmacSHA256 calculates HMAC-SHA256, returns byte array (exactly per official example)
func hmacSHA256(key []byte, content string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(content))
	return mac.Sum(nil)
}

// hashSHA256 calculates SHA256 hash, returns byte array (exactly per official example)
func hashSHA256(data []byte) []byte {
	hash := sha256.New()
	if _, err := hash.Write(data); err != nil {
		// Ignore error, per official example
	}
	return hash.Sum(nil)
}

// getSignedKey gets signing key (exactly per official example)
func getSignedKey(secretKey, date, region, service string) []byte {
	kDate := hmacSHA256([]byte(secretKey), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "request")
	return kSigning
}

func (c *volcClientSchema) signReq(method string) error {
	if c.signPassword == "" {
		return errors.New("volcengine: AccessKeySecret is missing")
	}

	mUrl, _ := url.Parse(c.APIHost)
	path := mUrl.Path
	if path == "" {
		path = "/"
	}

	// 1. Build request parameters
	// Action and Version are always in URL query string (per official example)
	var queries url.Values
	var bodyBytes []byte
	c.mutex.Lock()
	queries = url.Values{}
	if c.action != "" {
		queries.Set("Action", c.action)
	}
	queries.Set("Version", defaultVersion)
	c.mutex.Unlock()

	if method == http.MethodGet {
		// GET request: all parameters in query string
		c.mutex.Lock()
		if len(c.requestBody) > 0 {
			for k, v := range c.requestBody {
				queries.Set(k, fmt.Sprintf("%v", v))
			}
		}
		c.mutex.Unlock()
	} else if method == http.MethodPost {
		// POST request: Action and Version in query string, body is JSON format (per official example)
		c.mutex.Lock()
		if len(c.requestBody) > 0 {
			bodyBytes, _ = json.Marshal(c.requestBody)
		}
		c.mutex.Unlock()
	}

	// 2. Build signing material (exactly per official example)
	now := time.Now()
	date := now.UTC().Format("20060102T150405Z")
	authDate := date[:8]

	// Set X-Date header
	c.UpsertHeader("X-Date", date)

	// Set host header
	c.UpsertHeader("host", mUrl.Host)

	// Calculate payload (SHA256 of body)
	payload := hex.EncodeToString(hashSHA256(bodyBytes))
	c.UpsertHeader("X-Content-Sha256", payload)

	// Set Content-Type
	if method == http.MethodPost {
		c.UpsertHeader("Content-Type", "application/json; charset=UTF-8")
	}

	// Build query string (replace + with %20)
	queryString := strings.Replace(queries.Encode(), "+", "%20", -1)

	// Build signedHeaders (exactly per official example)
	signedHeaders := []string{"host", "x-date", "x-content-sha256", "content-type"}
	var headerList []string
	for _, header := range signedHeaders {
		if header == "host" {
			headerList = append(headerList, header+":"+mUrl.Host)
		} else {
			// Get header value from headerPairs
			var headerValue string
			for _, el := range c.headerPairs {
				if strings.ToLower(el.Key) == header {
					headerValue = strings.TrimSpace(el.Value)
					break
				}
			}
			headerList = append(headerList, header+":"+headerValue)
		}
	}
	headerString := strings.Join(headerList, "\n")

	// Build canonicalString (exactly per official example)
	canonicalString := strings.Join([]string{
		method,
		path,
		queryString,
		headerString + "\n",
		strings.Join(signedHeaders, ";"),
		payload,
	}, "\n")

	// Calculate hashedCanonicalString
	hashedCanonicalString := hex.EncodeToString(hashSHA256([]byte(canonicalString)))

	// Build credentialScope and signString (exactly per official example)
	credentialScope := authDate + "/" + c.credential.RegionID + "/dns/request"
	signString := strings.Join([]string{
		"HMAC-SHA256",
		date,
		credentialScope,
		hashedCanonicalString,
	}, "\n")

	// 3. Build authorization request header (exactly per official example)
	signedKey := getSignedKey(c.signPassword, authDate, c.credential.RegionID, "dns")
	signature := hex.EncodeToString(hmacSHA256(signedKey, signString))

	authorization := "HMAC-SHA256" +
		" Credential=" + c.credential.AccessKeyID + "/" + credentialScope +
		", SignedHeaders=" + strings.Join(signedHeaders, ";") +
		", Signature=" + signature

	c.UpsertHeader("Authorization", authorization)
	return nil
}
