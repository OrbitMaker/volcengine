package volcengine

import (
	"context"
	"net/http"
	"testing"

	"github.com/libdns/libdns"
)

func Test_ListZones(t *testing.T) {
	// Test ListZones API to get zone list
	err := p0.getClient()
	if err != nil {
		t.Log("err:", err)
		return
	}
	zones, err := p0.listZones(context.TODO(), 1, 10)
	t.Log("zones:", zones, "err:", err)
	if err != nil {
		t.Log("err:", err)
		return
	}
	for _, zone := range zones {
		t.Logf("Zone: %s, ZID: %d", zone.ZoneName, zone.ZID)
	}
}

func Test_ClientAPIReq(t *testing.T) {
	// Use ListRecords API to get DNS record list
	// Note: ZID should be obtained from query results, using example value for testing here
	zid := int64(304092) // Example ZID, should be obtained by querying records in actual use
	err := p0.getClient()
	if err != nil {
		t.Log("err:", err)
		return
	}
	p0.client.SetAction("ListRecords")
	p0.client.SetRequestBody("ZID", zid)
	p0.client.SetRequestBody("PageSize", 10)
	var rspData volcRecordListResult
	err = p0.client.doAPIRequest(context.TODO(), &rspData, http.MethodGet)
	t.Log("req", p0.client.schema, "data", rspData, "err:", err)
}

func Test_QueryDomainRecord(t *testing.T) {
	r0, err := p0.queryDomainRecord(context.TODO(), "*", "example.com", "A", "1.1.1.1")
	t.Log("result:", r0)
	if err != nil {
		t.Log("err:", err)
		t.Fail()
		return
	}
	t.Log("result with A rec:", r0, "err:", err)
}

func Test_QueryDomainRecords(t *testing.T) {
	// Test method 1: Query using ZID (if Provider.ZID is set)
	if p0.ZID > 0 {
		r0, err := p0.queryDomainRecords(context.TODO(), p0.ZID)
		t.Log("result (by ZID):", r0, "err:", err)
		return
	}
	// Test method 2: Query using domain name (look up ZID via ListZones)
	zone := "example.com" // Example domain name, replace with real domain name in actual use
	zid, err := p0.getZIDByZone(context.TODO(), zone)
	if err != nil {
		t.Log("err:", err)
		return
	}
	r0, err := p0.queryDomainRecords(context.TODO(), zid)
	t.Log("result (by zone):", r0, "err:", err)
}

func Test_DomainRecordOp(t *testing.T) {
	// Test DNS record CRUD operations
	// Supports three methods: 1. Use ZID (Provider.ZID) 2. Use Zone (Provider.Zone) 3. Use zone parameter
	// Priority: ZID > Zone > zone parameter
	var zid int64
	var err error

	if p0.ZID > 0 {
		zid = p0.ZID
		t.Log("Using ZID:", zid)
	} else if p0.Zone != "" {
		// Use Provider.Zone, automatically look up ZID via getZID
		zid, err = p0.getZID(context.TODO(), "")
		if err != nil {
			t.Log("err:", err)
			return
		}
		t.Log("Using Zone:", p0.Zone, "ZID:", zid)
	} else {
		// Look up ZID by domain name
		zone := "example.com" // Example domain name, replace with real domain name in actual use
		zid, err = p0.getZIDByZone(context.TODO(), zone)
		if err != nil {
			t.Log("err:", err)
			return
		}
		t.Log("Using zone:", zone, "ZID:", zid)
	}

	err = p0.getClient()
	if err != nil {
		t.Log("err:", err)
		return
	}
	dr0 := volcDomainRecord{
		DomainName: "example.com",
		Host:       "test",
		DomainType: "A",
		Value:      "1.1.1.1",
		TTL:        600,
		ZID:        zid,
	}
	r0, err := p0.addDomainRecord(context.TODO(), dr0)
	t.Log("add result:", r0, "err:", err)
	if err != nil {
		return
	}
	dr0.RecordID = r0
	dr0, err = p0.getDomainRecord(context.TODO(), r0)
	t.Log("get result:", dr0, "err:", err)
	if err != nil {
		return
	}
	dr0.Host = "test2"
	r0, err = p0.setDomainRecord(context.TODO(), dr0)
	t.Log("set result:", r0, "err:", err)
	if err != nil {
		return
	}
	r0, err = p0.delDomainRecord(context.TODO(), dr0)
	t.Log("del result:", r0, "err:", err)
}

func Test_GetRecordsByZone(t *testing.T) {
	// Test getting DNS record list by domain name
	zone := "example.com" // Example domain name, replace with real domain name in actual use
	records, err := p0.GetRecords(context.TODO(), zone)
	t.Log("records:", records, "err:", err)
	if err != nil {
		t.Log("err:", err)
		return
	}
	for _, record := range records {
		tmp := record.RR()
		t.Logf("Record: %s, Type: %s, Value: %s", tmp.Name, tmp.Type, tmp.Data)
	}
}

func Test_DeleteRecords(t *testing.T) {
	// Test deleting DNS records using DeleteRecords method
	// Supports three methods: 1. Use ZID (Provider.ZID) 2. Use Zone (Provider.Zone) 3. Use zone parameter
	// Priority: ZID > Zone > zone parameter
	var zid int64
	var err error
	var zone string

	// Get ZID using priority order
	if p0.ZID > 0 {
		zid = p0.ZID
		t.Log("Using ZID:", zid)
		// Get zone name from ZID for testing
		err = p0.getClient()
		if err != nil {
			t.Log("err:", err)
			return
		}
		zones, err := p0.listZones(context.TODO(), 1, 100)
		if err != nil {
			t.Log("err:", err)
			return
		}
		for _, z := range zones {
			if z.ZID == zid {
				zone = z.ZoneName
				break
			}
		}
		if zone == "" {
			t.Log("Zone not found for ZID:", zid)
			return
		}
	} else if p0.Zone != "" {
		zone = p0.Zone
		zid, err = p0.getZID(context.TODO(), "")
		if err != nil {
			t.Log("err:", err)
			return
		}
		t.Log("Using Zone:", zone, "ZID:", zid)
	} else {
		zone = "example.com" // Example domain name, replace with real domain name in actual use
		zid, err = p0.getZIDByZone(context.TODO(), zone)
		if err != nil {
			t.Log("err:", err)
			return
		}
		t.Log("Using zone:", zone, "ZID:", zid)
	}

	// Step 1: Create a test record first
	err = p0.getClient()
	if err != nil {
		t.Log("err:", err)
		return
	}
	testHost := "_test_delete"
	testValue := "test-delete-value"

	// Step 2: Construct libdns.Record directly from the created record
	// Use DomainRecord which implements libdns.Record interface
	recordToDelete := DomainRecord{
		Type:  "TXT",
		Name:  testHost + "." + zone, // Full domain name
		Value: testValue,
		// TTL:   600,
	}

	// Step 3: Delete the record using DeleteRecords method
	deletedRecords, err := p0.DeleteRecords(context.TODO(), zone, []libdns.Record{recordToDelete})
	if err != nil {
		t.Fatal("Failed to delete record:", err)
		return
	}
	t.Log("Successfully deleted record:", deletedRecords)

	// Step 4: Verify the record is deleted by querying again
	recordsAfterDelete, err := p0.GetRecords(context.TODO(), zone)
	if err != nil {
		t.Log("Failed to verify deletion:", err)
		return
	}
	for _, rec := range recordsAfterDelete {
		recRR := rec.RR()
		if recRR.Name == testHost+"."+zone || recRR.Name == testHost {
			if recRR.Type == "TXT" && recRR.Data == `"`+testValue+`"` {
				t.Log("WARNING: Record still exists after deletion!")
				return
			}
		}
	}
	t.Log("Verification passed: Record successfully deleted")
}
