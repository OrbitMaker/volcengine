package volcengine

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

type ttl_t = uint32

type DomainRecord struct {
	Type     string
	Name     string
	Value    string
	TTL      ttl_t
	Priority ttl_t
	ID       string
}

func (r DomainRecord) RR() libdns.RR {
	result := libdns.RR{
		Type: r.Type,
		Name: r.Name,
		Data: r.Value,
		TTL:  time.Duration(r.TTL) * time.Second,
	}
	if r.Priority > 0 {
		result.Data = fmt.Sprintf("%d %v", r.Priority, r.Value)
	}
	return result
}

type volcDomainRecord struct {
	Host       string `json:"Host,omitempty"`
	Line       string `json:"Line,omitempty"`
	DomainType string `json:"Type,omitempty"`
	DomainName string `json:"DomainName,omitempty"`
	Value      string `json:"Value,omitempty"`
	RecordID   string `json:"RecordID,omitempty"`
	TTL        ttl_t  `json:"TTL,omitempty"`
	Weight     int    `json:"Weight,omitempty"`
	Priority   ttl_t  `json:"Priority,omitempty"`
	ZID        int64  `json:"ZID,omitempty"`
}

func (r volcDomainRecord) DomainRecord() DomainRecord {
	return DomainRecord{
		Type:     r.DomainType,
		Name:     r.Host,
		Value:    r.Value,
		TTL:      r.TTL,
		Priority: r.Priority,
		ID:       r.RecordID,
	}
}

type volcRecordListResult struct {
	ResponseMetadata volcResponseMetadata `json:"ResponseMetadata,omitempty"`
	Result           volcRecordsResult    `json:"Result,omitempty"`
}

type volcEmptyResult struct {
	ResponseMetadata volcResponseMetadata `json:"ResponseMetadata,omitempty"`
}

type volcAPIResponse struct {
	ResponseMetadata struct {
		RequestID string `json:"RequestId,omitempty"`
		Action    string `json:"Action,omitempty"`
		Version   string `json:"Version,omitempty"`
		Service   string `json:"Service,omitempty"`
		Region    string `json:"Region,omitempty"`
		Error     struct {
			Code    string `json:"Code,omitempty"`
			Message string `json:"Message,omitempty"`
		} `json:"Error,omitempty"`
	} `json:"ResponseMetadata,omitempty"`
}

type volcResponseMetadata struct {
	RequestID string `json:"RequestId,omitempty"`
	Action    string `json:"Action,omitempty"`
	Version   string `json:"Version,omitempty"`
	Service   string `json:"Service,omitempty"`
	Region    string `json:"Region,omitempty"`
}

type volcRecordResult struct {
	RecordID    string   `json:"RecordID,omitempty"`
	FQDN        string   `json:"FQDN,omitempty"`
	PQDN        string   `json:"PQDN,omitempty"`
	Host        string   `json:"Host,omitempty"`
	Type        string   `json:"Type,omitempty"`
	TTL         ttl_t    `json:"TTL,omitempty"`
	Line        string   `json:"Line,omitempty"`
	Value       string   `json:"Value,omitempty"`
	Weight      int      `json:"Weight,omitempty"`
	RecordSetID string   `json:"RecordSetID,omitempty"`
	Enable      bool     `json:"Enable,omitempty"`
	Tags        []string `json:"Tags,omitempty"`
	Remark      string   `json:"Remark,omitempty"`
	CreatedAt   string   `json:"CreatedAt,omitempty"`
	UpdatedAt   string   `json:"UpdatedAt,omitempty"`
	Operators   []string `json:"Operators,omitempty"`
	Priority    ttl_t    `json:"Priority,omitempty"`
}

type volcRecordResultWrapper struct {
	ResponseMetadata volcResponseMetadata `json:"ResponseMetadata,omitempty"`
	Result           volcRecordResult     `json:"Result,omitempty"`
}

type volcRecordsResult struct {
	PageNumber int                `json:"PageNumber,omitempty"`
	PageSize   int                `json:"PageSize,omitempty"`
	Records    []volcRecordResult `json:"Records,omitempty"`
	TotalCount int                `json:"TotalCount,omitempty"`
}

// ListZones API response structure
type volcZoneTag struct {
	Key   string `json:"Key,omitempty"`
	Value string `json:"Value,omitempty"`
}

type volcZoneInfo struct {
	CreatedAt         string        `json:"CreatedAt,omitempty"`
	UpdatedAt         string        `json:"UpdatedAt,omitempty"`
	ZID               int64         `json:"ZID,omitempty"`
	ZoneName          string        `json:"ZoneName,omitempty"`
	InstanceID        string        `json:"InstanceID,omitempty"`
	Remark            string        `json:"Remark,omitempty"`
	ConfigurationCode string        `json:"ConfigurationCode,omitempty"`
	RecordCount       int           `json:"RecordCount,omitempty"`
	ExpiredTime       int64         `json:"ExpiredTime,omitempty"`
	LastOperator      string        `json:"LastOperator,omitempty"`
	CacheStage        int           `json:"CacheStage,omitempty"`
	TradeCode         string        `json:"TradeCode,omitempty"`
	DnsSecurity       string        `json:"DnsSecurity,omitempty"`
	IsSubDomain       bool          `json:"IsSubDomain,omitempty"`
	ProjectName       string        `json:"ProjectName,omitempty"`
	Tags              []volcZoneTag `json:"Tags,omitempty"`
	AutoRenew         string        `json:"AutoRenew,omitempty"`
}

type volcZonesResult struct {
	Total int            `json:"Total,omitempty"`
	Zones []volcZoneInfo `json:"Zones,omitempty"`
}

type volcZonesListResult struct {
	ResponseMetadata volcResponseMetadata `json:"ResponseMetadata,omitempty"`
	Result           volcZonesResult      `json:"Result,omitempty"`
}

func (r *volcRecordResult) ToDomainRecord() volcDomainRecord {
	return volcDomainRecord{
		RecordID:   r.RecordID,
		DomainType: r.Type,
		Host:       r.Host,
		Value:      r.Value,
		TTL:        r.TTL,
		Line:       r.Line,
		Weight:     r.Weight,
		Priority:   r.Priority,
	}
}

// extractRootDomain extracts root domain from full domain name (usually the last two labels)
// Example: "www.baidu.com" -> "baidu.com", "sub1.sub2.example.com" -> "example.com"
func extractRootDomain(fullDomain string) string {
	fullDomain = strings.Trim(fullDomain, ".")
	parts := strings.Split(fullDomain, ".")
	if len(parts) < 2 {
		return fullDomain
	}
	// Return the last two labels as root domain
	return strings.Join(parts[len(parts)-2:], ".")
}

func volcRecord(r libdns.Record, zone string, zid int64) volcDomainRecord {
	result := volcDomainRecord{
		ZID:  zid,
		Line: "default",
	}
	if r == nil {
		return result
	}
	tmpRR := r.RR()

	// If zone is empty, try to automatically extract root domain from record name
	actualZone := zone
	if len(actualZone) == 0 && len(tmpRR.Name) > 0 {
		// If record name is a full domain name (contains dots), extract root domain
		if strings.Contains(tmpRR.Name, ".") {
			actualZone = extractRootDomain(tmpRR.Name)
		}
	}

	if len(actualZone) > 0 {
		result.Host = libdns.RelativeName(tmpRR.Name, actualZone)
		if result.Host == "" {
			result.Host = "@"
		}
		result.DomainName = strings.Trim(actualZone, ".")
	} else {
		// If still cannot determine zone, use original name directly
		result.Host = tmpRR.Name
	}

	result.DomainType = tmpRR.Type
	result.Value = tmpRR.Data
	result.TTL = ttl_t(tmpRR.TTL.Seconds())
	if result.TTL <= 0 {
		result.TTL = 600
	}
	if svcb, svcbok := r.(libdns.ServiceBinding); svcbok {
		result.Priority = ttl_t(svcb.Priority)
		result.Value = fmt.Sprintf("%s %s", svcb.Target, svcb.Params)
	}
	if rec, ok := r.(DomainRecord); ok {
		result.RecordID = rec.ID
	}
	return result
}

type opErrors struct {
	Op            string
	length        uint64
	errMsgBuilder strings.Builder
}

func OpErrors(op string) *opErrors {
	return &opErrors{
		Op:            op,
		errMsgBuilder: strings.Builder{},
		length:        0,
	}
}

func OpError(op string, err error) error {
	errs := OpErrors(op)
	errs.JoinError(err)
	return errs.Error()
}

func (e *opErrors) JoinError(err error) *opErrors {
	if err != nil {
		msg := "caused with: "
		msg += err.Error()
		e.errMsgBuilder.WriteString(msg + ",")
		e.length += 1
	}
	return e
}

func (e *opErrors) JoinRecord(record libdns.Record, err error) *opErrors {
	if err != nil {
		msg := "caused at record named "
		msg += "'" + record.RR().Name + "'"
		msg += ": "
		msg += err.Error()
		e.errMsgBuilder.WriteString(msg + ",")
		e.length += 1
	}
	return e
}

func (e *opErrors) errorMsg() string {
	msg := "there is something error "
	if len(e.Op) > 0 {
		msg += "when '" + e.Op + "'"
	}
	msg += ": "
	return strings.TrimSuffix(msg+e.errMsgBuilder.String(), ",")
}

func (e *opErrors) Error() error {
	if e.length > 0 {
		return errors.New(e.errorMsg())
	}
	return nil
}
