package volcengine

import (
	"context"
	"fmt"
	"strings"

	"github.com/libdns/libdns"
)

// Provider implements the libdns interfaces for VolcEngine DNS.
type Provider struct {
	client *volcClient
	CredentialInfo
	// ZID is the domain ID used to directly query DNS records (highest priority)
	// VolcEngine API uses ZID instead of zone to identify domains
	ZID int64 `json:"zid,omitempty"`
	// Zone is the domain name. When ZID is not set, it will look up the corresponding ZID via ListZones API
	// Priority: ZID > Zone > zone parameter
	Zone string `json:"zone,omitempty"`
}

// AppendRecords adds records. It returns the records that were added.
// ZID resolution priority: Provider.ZID > Provider.Zone > zone parameter
func (p *Provider) AppendRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	var rls []libdns.Record
	var errs = OpErrors("AppendRecords")
	if err := p.getClient(); err != nil {
		return nil, OpError("AppendRecords", err)
	}

	zid, err := p.getZID(ctx, zone)
	if err != nil {
		return nil, OpError("AppendRecords", err)
	}

	for _, rec := range recs {
		// zone parameter is not used, volcRecord will automatically extract root domain from record name
		vr := volcRecord(rec, "", zid)
		rid, err := p.addDomainRecord(ctx, vr)
		if err != nil {
			errs.JoinRecord(rec, err)
			continue
		}
		vr.RecordID = rid
		rls = append(rls, vr.DomainRecord())
	}
	return rls, errs.Error()
}

// DeleteRecords deletes the records. If a record does not have an ID,
// it will be looked up. It returns the records that were deleted.
// ZID resolution priority: Provider.ZID > Provider.Zone > zone parameter
func (p *Provider) DeleteRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	var rls []libdns.Record
	var errs = OpErrors("DeleteRecords")
	if err := p.getClient(); err != nil {
		return nil, OpError("DeleteRecords", err)
	}

	zid, err := p.getZID(ctx, zone)
	if err != nil {
		return nil, OpError("DeleteRecords", err)
	}

	for _, rec := range recs {
		recRR := rec.RR()
		// Use zone parameter to help volcRecord extract root domain correctly
		// When zone is provided, use it; otherwise volcRecord will extract from record name
		vr := volcRecord(rec, zone, zid)
		if vr.RecordID == "" {
			// Get root domain for queryDomainRecord
			// Priority: use zone parameter > extract from record name > use DomainName from vr
			rootDomain := strings.Trim(zone, ".")
			if rootDomain == "" {
				// If zone is empty, try to extract from record name
				if strings.Contains(recRR.Name, ".") {
					rootDomain = extractRootDomain(recRR.Name)
				} else {
					// If record name is relative and zone is empty, use DomainName from vr
					rootDomain = vr.DomainName
				}
			}

			// queryDomainRecord will build PQDN internally: host + "." + rootDomain
			// Pass rootDomain as zoneName parameter
			r0, err := p.queryDomainRecord(ctx, vr.Host, rootDomain, vr.DomainType, vr.Value)
			if err != nil {
				errs.JoinRecord(rec, err)
				continue
			}
			vr = r0
		}
		_, err := p.delDomainRecord(ctx, vr)
		if err != nil {
			errs.JoinRecord(rec, err)
			continue
		}
		rls = append(rls, vr.DomainRecord())
	}
	return rls, errs.Error()
}

// GetRecords lists all the records.
// ZID resolution priority: Provider.ZID > Provider.Zone > zone parameter
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	var rls []libdns.Record
	if err := p.getClient(); err != nil {
		return nil, OpError("GetRecords", err)
	}

	zid, err := p.getZID(ctx, zone)
	if err != nil {
		return nil, OpError("GetRecords", err)
	}

	recs, err := p.queryDomainRecords(ctx, zid)
	if err != nil {
		return nil, OpError("GetRecords", err)
	}
	for _, rec := range recs {
		rls = append(rls, rec.DomainRecord())
	}
	return rls, nil
}

// SetRecords sets the records, either by updating existing records
// or creating new ones. It returns the updated records.
// ZID resolution priority: Provider.ZID > Provider.Zone > zone parameter
func (p *Provider) SetRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	var rls []libdns.Record
	var errs = OpErrors("SetRecords")
	if err := p.getClient(); err != nil {
		return nil, OpError("SetRecords", err)
	}

	zid, err := p.getZID(ctx, zone)
	if err != nil {
		return nil, OpError("SetRecords", err)
	}

	for _, rec := range recs {
		recRR := rec.RR()
		// Use zone parameter to help volcRecord extract root domain correctly
		// When zone is provided, use it; otherwise volcRecord will extract from record name
		vr := volcRecord(rec, zone, zid)
		if vr.RecordID == "" {
			// Get root domain for queryDomainRecord
			// Priority: use zone parameter > extract from record name > use DomainName from vr
			rootDomain := strings.Trim(zone, ".")
			if rootDomain == "" {
				// If zone is empty, try to extract from record name
				if strings.Contains(recRR.Name, ".") {
					rootDomain = extractRootDomain(recRR.Name)
				} else {
					// If record name is relative and zone is empty, use DomainName from vr
					rootDomain = vr.DomainName
				}
			}
			// queryDomainRecord will build PQDN internally: host + "." + rootDomain
			// Pass rootDomain as zoneName parameter
			r0, err := p.queryDomainRecord(ctx, vr.Host, rootDomain, vr.DomainType, vr.Value)
			if err != nil {
				vr.RecordID, err = p.addDomainRecord(ctx, vr)
				if err != nil {
					errs.JoinRecord(rec, err)
					continue
				}
			} else {
				vr.RecordID = r0.RecordID
				vr.ZID = r0.ZID
			}
		}
		_, err := p.setDomainRecord(ctx, vr)
		if err != nil {
			errs.JoinRecord(rec, err)
			continue
		}
		rls = append(rls, vr.DomainRecord())
	}
	return rls, errs.Error()
}

func (p *Provider) getClient() error {
	if p.client == nil {
		var err error
		p.client, err = getClient(&p.CredentialInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Provider) addDomainRecord(ctx context.Context, rc volcDomainRecord) (recID string, err error) {
	if err := p.getClient(); err != nil {
		return "", err
	}
	return p.client.addDomainRecord(ctx, rc)
}

func (p *Provider) delDomainRecord(ctx context.Context, rc volcDomainRecord) (recID string, err error) {
	if err := p.getClient(); err != nil {
		return "", err
	}
	return p.client.delDomainRecord(ctx, rc)
}

func (p *Provider) setDomainRecord(ctx context.Context, rc volcDomainRecord) (recID string, err error) {
	if err := p.getClient(); err != nil {
		return "", err
	}
	return p.client.setDomainRecord(ctx, rc)
}

func (p *Provider) getDomainRecord(ctx context.Context, recID string) (volcDomainRecord, error) {
	if err := p.getClient(); err != nil {
		return volcDomainRecord{}, err
	}
	return p.client.getDomainRecord(ctx, recID)
}

func (p *Provider) queryDomainRecords(ctx context.Context, zid int64) ([]volcDomainRecord, error) {
	if err := p.getClient(); err != nil {
		return nil, err
	}
	return p.client.queryDomainRecords(ctx, zid)
}

func (p *Provider) queryDomainRecord(ctx context.Context, host, zone string, recType string, recVal string) (volcDomainRecord, error) {
	if err := p.getClient(); err != nil {
		return volcDomainRecord{}, err
	}
	return p.client.queryDomainRecord(ctx, host, zone, recType, recVal)
}

// listZones gets the list of zones (domains)
func (p *Provider) listZones(ctx context.Context, pageNumber, pageSize int) ([]volcZoneInfo, error) {
	if err := p.getClient(); err != nil {
		return nil, err
	}
	return p.client.listZones(ctx, pageNumber, pageSize)
}

// getZIDByZone looks up the corresponding ZID by domain name
func (p *Provider) getZIDByZone(ctx context.Context, zone string) (int64, error) {
	if zone == "" {
		return 0, fmt.Errorf("zone is required")
	}
	// Normalize domain name (remove trailing dots)
	zone = strings.Trim(strings.ToLower(zone), ".")

	// Get all zones (paginated, up to 500)
	zones, err := p.listZones(ctx, 1, 500)
	if err != nil {
		return 0, fmt.Errorf("failed to list zones: %w", err)
	}

	// Find matching zone
	for _, z := range zones {
		if strings.Trim(strings.ToLower(z.ZoneName), ".") == zone {
			return z.ZID, nil
		}
	}

	return 0, fmt.Errorf("zone %s not found", zone)
}

// getZID gets ZID with priority: Provider.ZID > Provider.Zone > zone parameter
func (p *Provider) getZID(ctx context.Context, zone string) (int64, error) {
	// Priority 1: Use Provider.ZID (if set)
	if p.ZID > 0 {
		return p.ZID, nil
	}

	// Priority 2: Use Provider.Zone (if set)
	if p.Zone != "" {
		return p.getZIDByZone(ctx, p.Zone)
	}

	// Priority 3: Use zone parameter (if provided)
	if zone != "" {
		return p.getZIDByZone(ctx, zone)
	}

	return 0, fmt.Errorf("ZID or zone is required (set Provider.ZID, Provider.Zone, or pass zone parameter)")
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
