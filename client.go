package volcengine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type volcClient struct {
	schema *volcClientSchema
	mutex  sync.Mutex
}

func getClient(cred *CredentialInfo) (*volcClient, error) {
	schema, err := getClientSchema(cred)
	if err != nil {
		return nil, err
	}
	return &volcClient{
		schema: schema,
	}, nil
}

func (c *volcClient) SetAction(action string) error {
	if c.schema == nil {
		return errors.New("schema was not initialized properly")
	}
	return c.schema.SetAction(action)
}

func (c *volcClient) SetRequestBody(key string, value interface{}) error {
	if c.schema == nil {
		return errors.New("schema was not initialized properly")
	}
	return c.schema.UpsertRequestBody(key, value)
}

func (c *volcClientSchema) UpsertRequestBody(key string, value interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.requestBody == nil {
		c.requestBody = make(map[string]interface{})
	}
	c.requestBody[key] = value
	return nil
}

func (c *volcClient) Lock() {
	c.mutex.Lock()
}

func (c *volcClient) Unlock() {
	c.mutex.Unlock()
}

func (c *volcClient) doAPIRequest(ctx context.Context, result interface{}, methods ...string) error {
	method := http.MethodPost
	if len(methods) > 0 {
		method = methods[0]
	}
	req, err := c.schema.HttpRequest(ctx, method)
	if err != nil {
		return err
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	var buf []byte
	buf, err = io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	var apiResp volcAPIResponse
	err = json.Unmarshal(buf, &apiResp)
	if err != nil {
		return err
	}

	if rsp.StatusCode != 200 {
		return fmt.Errorf("API error: HTTP %d: %s", rsp.StatusCode, apiResp.ResponseMetadata.Error.Message)
	}

	if apiResp.ResponseMetadata.Error.Code != "" {
		return fmt.Errorf("API error: %s - %s", apiResp.ResponseMetadata.Error.Code, apiResp.ResponseMetadata.Error.Message)
	}

	return json.Unmarshal(buf, result)
}
