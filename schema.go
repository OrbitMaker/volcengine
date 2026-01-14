package volcengine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const defaultRegionID string = "cn-beijing"
const addressOfAPI string = "https://dns.volcengineapi.com"
const defaultVersion string = "2018-08-01"

// CredentialInfo implements param of the credential
type CredentialInfo struct {
	// The API Key ID Required by VolcEngine for accessing the API
	AccessKeyID string `json:"access_key_id"`
	// The API Key Secret Required by VolcEngine for accessing the API
	AccessKeySecret string `json:"access_key_secret"`
	// Optional for identifying the region of the VolcEngine Service, default is cn-beijing
	RegionID string `json:"region_id,omitempty"`
}

// volcClientSchema abstracts the volcengine DNS client
type volcClientSchema struct {
	mutex        sync.Mutex
	APIHost      string
	headerPairs  keyPairs
	requestBody  map[string]interface{}
	signString   string
	signPassword string
	credential   *CredentialInfo
	action       string // Store Action parameter
}

// keyPair implements K-V struct
type keyPair struct {
	Key   string
	Value string
}

func getClientSchema(cred *CredentialInfo) (*volcClientSchema, error) {
	if cred.AccessKeyID == "" || cred.AccessKeySecret == "" {
		return nil, errors.New("empty AccessKeyID or AccessKeySecret")
	}
	if len(cred.RegionID) == 0 {
		cred.RegionID = defaultRegionID
	}
	return defaultSchema(cred)
}

func (c *volcClientSchema) UpsertHeader(key string, value string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	var err error
	c.headerPairs, err = c.headerPairs.Upsert(key, value)
	return err
}

func (c *volcClientSchema) SetRequestBody(body map[string]interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.requestBody = body
	return nil
}

func (c *volcClientSchema) SetAction(action string) error {
	if len(action) == 0 {
		return errors.New("empty action to set")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.action = action
	return nil
}

// HttpRequest generates http.Request from schema
func (c *volcClientSchema) HttpRequest(ctx context.Context, method string) (*http.Request, error) {
	if method == "" {
		method = http.MethodGet
	}

	if err := c.signReq(method); err != nil {
		return nil, err
	}

	requestUrl := c.APIHost
	var bodyReader io.Reader

	// Action and Version are always in URL query string (per official example)
	queryParams := url.Values{}
	c.mutex.Lock()
	if c.action != "" {
		queryParams.Set("Action", c.action)
	}
	queryParams.Set("Version", defaultVersion)
	c.mutex.Unlock()

	if method == http.MethodGet {
		// GET request: all parameters in URL query string
		if len(c.requestBody) > 0 {
			c.mutex.Lock()
			for k, v := range c.requestBody {
				queryParams.Add(k, fmt.Sprintf("%v", v))
			}
			c.mutex.Unlock()
		}
		if len(queryParams) > 0 {
			requestUrl = fmt.Sprintf("%s?%s", requestUrl, queryParams.Encode())
		}
	} else if method == http.MethodPost {
		// POST request: Action and Version in query string, other parameters in JSON body (per official example)
		if len(queryParams) > 0 {
			requestUrl = fmt.Sprintf("%s?%s", requestUrl, queryParams.Encode())
		}
		// POST request body only contains business parameters, not Action and Version
		if len(c.requestBody) > 0 {
			bodyBytes, err := json.Marshal(c.requestBody)
			if err != nil {
				return nil, err
			}
			bodyReader = bytes.NewReader(bodyBytes)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, requestUrl, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}

	for _, v := range c.headerPairs {
		req.Header.Set(v.Key, v.Value)
	}

	return req, nil
}

type keyPairs []keyPair

func (p keyPairs) Upsert(key, value string) (keyPairs, error) {
	if key == "" || value == "" {
		return p, errors.New("key or value is empty")
	}
	srcEl := keyPair{Key: key, Value: value}
	for i, el := range p {
		if srcEl.Key == el.Key {
			p[i] = srcEl
			return p, nil
		}
	}
	p = append(p, srcEl)
	return p, nil
}

func (p keyPairs) SplitToString(pair, pairs string) string {
	result := ""
	if len(pair) == 0 {
		pair = ":"
	}
	if len(pairs) == 0 {
		pairs = ","
	}
	for _, el := range p {
		result += el.Key + pair + el.Value + pairs
	}
	return strings.TrimSuffix(result, pairs)
}

func (p keyPairs) Keys() []string {
	result := make([]string, p.Len())
	for index, el := range p {
		result[index] = el.Key
	}
	return result
}

func (p keyPairs) Len() int {
	return len(p)
}

func (p keyPairs) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p keyPairs) Less(i, j int) bool {
	return p[i].Key < p[j].Key
}
