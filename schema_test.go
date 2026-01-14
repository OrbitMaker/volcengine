package volcengine

import (
	"context"
	"testing"
)

func Test_RequestUrl(t *testing.T) {
	err := p0.getClient()
	if err != nil {
		t.Log("err:", err)
		return
	}
	// Test request building using ListRecords API
	// ZID should be obtained from query results, using example value for testing here
	zid := int64(304092) // Example ZID
	p0.client.SetAction("ListRecords")
	p0.client.SetRequestBody("ZID", zid)
	p0.client.SetRequestBody("PageSize", 10)
	r, err := p0.client.schema.HttpRequest(context.TODO(), "GET")
	if err != nil {
		t.Log("err:", err)
		return
	}
	t.Log("url:", r.URL.String(), "err:", err)
}
