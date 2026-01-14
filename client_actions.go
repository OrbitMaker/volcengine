package volcengine

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

func (c *volcClient) addDomainRecord(ctx context.Context, rc volcDomainRecord) (recID string, err error) {
	if c.schema == nil {
		return "", errors.New("schema was not initialized properly")
	}
	c.Lock()
	defer c.Unlock()

	// Set default values
	if rc.TTL <= 0 {
		rc.TTL = 600
	}
	line := rc.Line
	if line == "" {
		line = "default" // Default line
	}

	// Set required parameters
	c.SetAction("CreateRecord")
	c.SetRequestBody("ZID", rc.ZID)
	c.SetRequestBody("Host", rc.Host)
	c.SetRequestBody("Type", rc.DomainType)
	c.SetRequestBody("Value", rc.Value)

	// Set optional parameters
	c.SetRequestBody("Line", line)
	c.SetRequestBody("TTL", rc.TTL)
	if rc.Weight > 0 {
		c.SetRequestBody("Weight", rc.Weight)
	}

	rs := volcRecordResultWrapper{}
	err = c.doAPIRequest(ctx, &rs)
	if err != nil {
		return "", err
	}
	// RecordID is already a string type, return directly
	return rs.Result.RecordID, nil
}

func (c *volcClient) delDomainRecord(ctx context.Context, rc volcDomainRecord) (recID string, err error) {
	if c.schema == nil {
		return "", errors.New("schema was not initialized properly")
	}
	c.Lock()
	defer c.Unlock()
	// RecordID should be string type (per official example)
	c.SetAction("DeleteRecord")
	c.SetRequestBody("RecordID", rc.RecordID)
	rs := volcEmptyResult{}
	err = c.doAPIRequest(ctx, &rs)
	if err != nil {
		return "", err
	}
	return rc.RecordID, nil
}

func (c *volcClient) setDomainRecord(ctx context.Context, rc volcDomainRecord) (recID string, err error) {
	if c.schema == nil {
		return "", errors.New("schema was not initialized properly")
	}
	c.Lock()
	defer c.Unlock()

	// Set default values
	if rc.TTL <= 0 {
		rc.TTL = 600
	}
	line := rc.Line
	if line == "" {
		line = "default" // Default line
	}

	// Set required parameters
	c.SetAction("UpdateRecord")
	c.SetRequestBody("RecordID", rc.RecordID) // RecordID should be string type (per official example)
	c.SetRequestBody("Host", rc.Host)
	c.SetRequestBody("Type", rc.DomainType)
	c.SetRequestBody("Value", rc.Value)

	// Set optional parameters
	c.SetRequestBody("Line", line)
	c.SetRequestBody("TTL", rc.TTL)
	if rc.Weight > 0 {
		c.SetRequestBody("Weight", rc.Weight)
	}

	rs := volcRecordResultWrapper{}
	err = c.doAPIRequest(ctx, &rs)
	if err != nil {
		return "", err
	}
	return rs.Result.RecordID, nil
}

func (c *volcClient) getDomainRecord(ctx context.Context, recID string) (volcDomainRecord, error) {
	if c.schema == nil {
		return volcDomainRecord{}, errors.New("schema was not initialized properly")
	}
	c.Lock()
	defer c.Unlock()
	// RecordID should be string type (per official example)
	c.SetAction("QueryRecord")
	c.SetRequestBody("RecordID", recID)
	rs := volcRecordResultWrapper{}
	err := c.doAPIRequest(ctx, &rs, http.MethodGet)
	if err != nil {
		return volcDomainRecord{}, err
	}
	return rs.Result.ToDomainRecord(), nil
}

func (c *volcClient) queryDomainRecords(ctx context.Context, zoneID int64) ([]volcDomainRecord, error) {
	if c.schema == nil {
		return nil, errors.New("schema was not initialized properly")
	}
	c.Lock()
	defer c.Unlock()
	c.SetAction("ListRecords")
	c.SetRequestBody("ZID", zoneID)
	c.SetRequestBody("PageSize", 500)
	rs := volcRecordListResult{}
	err := c.doAPIRequest(ctx, &rs, http.MethodGet)
	if err != nil {
		return []volcDomainRecord{}, err
	}
	var records []volcDomainRecord
	for _, rec := range rs.Result.Records {
		record := rec.ToDomainRecord()
		record.ZID = zoneID
		records = append(records, record)
	}
	return records, nil
}

func (c *volcClient) queryDomainRecord(ctx context.Context, host, zoneName string, recType string, recVal ...string) (volcDomainRecord, error) {
	if c.schema == nil {
		return volcDomainRecord{}, errors.New("schema was not initialized properly")
	}
	c.Lock()
	defer c.Unlock()
	pqdn := host
	if host != "@" {
		pqdn = host + "." + strings.Trim(zoneName, ".")
	} else {
		pqdn = strings.Trim(zoneName, ".")
	}
	c.SetAction("QueryRecord")
	c.SetRequestBody("PQDN", pqdn)
	c.SetRequestBody("Type", recType)
	c.SetRequestBody("Line", "default") // Required parameter, default line
	// Value parameter: use if provided, otherwise set to empty string (per documentation, all parameters required when RecordID is absent)
	value := ""
	if len(recVal) > 0 {
		value = recVal[0]
	}
	c.SetRequestBody("Value", value)
	rs := volcRecordResultWrapper{}
	err := c.doAPIRequest(ctx, &rs, http.MethodGet)
	if err != nil {
		return volcDomainRecord{}, err
	}
	return rs.Result.ToDomainRecord(), nil
}

// listZones gets the list of zones (domains)
func (c *volcClient) listZones(ctx context.Context, pageNumber, pageSize int) ([]volcZoneInfo, error) {
	if c.schema == nil {
		return nil, errors.New("schema was not initialized properly")
	}
	c.Lock()
	defer c.Unlock()
	c.SetAction("ListZones")
	if pageNumber > 0 {
		c.SetRequestBody("PageNumber", pageNumber)
	}
	if pageSize > 0 {
		c.SetRequestBody("PageSize", pageSize)
	}
	rs := volcZonesListResult{}
	err := c.doAPIRequest(ctx, &rs, http.MethodGet)
	if err != nil {
		return nil, err
	}
	return rs.Result.Zones, nil
}
