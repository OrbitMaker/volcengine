package volcengine

import (
	"errors"
	"testing"

	"github.com/libdns/libdns"
)

func Test_volcDomainRecordWithZone(t *testing.T) {
	type testCase struct {
		memo   string
		record libdns.Record
		zone   string
		zid    int64
		result volcDomainRecord
	}

	cases := []testCase{
		{
			memo: "record.Name without zone",
			record: libdns.RR{
				Name: "sub",
			},
			zone: "mydomain.com.",
			zid:  12345,
			result: volcDomainRecord{
				Host:       "sub",
				DomainName: "mydomain.com",
				ZID:        12345,
			},
		},
		{
			// libdns passes relative names; ACME DNS-01 records have two or more labels
			memo: "multi-label relative record.Name with zone",
			record: libdns.RR{
				Name: "_acme-challenge.sub",
			},
			zone: "mydomain.com.",
			zid:  12345,
			result: volcDomainRecord{
				Host:       "_acme-challenge.sub",
				DomainName: "mydomain.com",
				ZID:        12345,
			},
		},
		{
			memo: "record.Name with zone",
			record: libdns.RR{
				Name: "sub.mydomain.com",
			},
			zone: "mydomain.com.",
			zid:  12345,
			result: volcDomainRecord{
				Host:       "sub",
				DomainName: "mydomain.com",
				ZID:        12345,
			},
		},
	}

	for _, c := range cases {
		rec := volcRecord(c.record, c.zone, c.zid)
		if rec.Host != c.result.Host || rec.DomainName != c.result.DomainName || rec.ZID != c.result.ZID {
			t.Log("excepted:", c.result, "got:", rec)
			t.Fail()
		}
		t.Log("case ", c.memo, "was pass.")
	}
}

func Test_volcDomainRecord(t *testing.T) {
	type testCase struct {
		memo   string
		record libdns.Record
		zone   string
		zid    int64
		result volcDomainRecord
	}

	cases := []testCase{
		{
			memo: "normal record",
			record: libdns.RR{
				Name: "sub",
				Type: "A",
				Data: "1.1.1.1",
			},
			zone: "example.com",
			zid:  12345,
			result: volcDomainRecord{
				Host:       "sub",
				DomainType: "A",
				Value:      "1.1.1.1",
				DomainName: "example.com",
				ZID:        12345,
			},
		},
		{
			memo: "HTTPS record",
			record: libdns.ServiceBinding{
				Name:     "sub",
				Scheme:   "https",
				Target:   "target.com",
				Priority: 100,
				Params: map[string][]string{
					"alpn": {"333"},
				},
			},
			zone: "example.com",
			zid:  12345,
			result: volcDomainRecord{
				Host:       "sub",
				DomainType: "HTTPS",
				Priority:   100,
				DomainName: "example.com",
				ZID:        12345,
			},
		},
	}

	for _, c := range cases {
		rec := volcRecord(c.record, c.zone, c.zid)
		if rec.Host != c.result.Host || rec.DomainType != c.result.DomainType || rec.ZID != c.result.ZID {
			t.Log("excepted:", c.result, "got:", rec)
			t.Fail()
			return
		}
		t.Log("case ", c.memo, "was pass.")
	}
}

func TestOpError(t *testing.T) {
	type testCase struct {
		memo   string
		record error
		result error
	}

	cases := []testCase{
		{
			memo:   "error with op",
			record: OpError("op", errors.New("something wrong")),
			result: errors.New("there is something error when 'op': caused with: something wrong"),
		},
		{
			memo:   "error with none error",
			record: OpError("op", nil),
			result: nil,
		},
	}

	for _, c := range cases {
		if (c.record != nil && c.result != nil) && c.record.Error() != c.result.Error() {
			t.Log("excepted:", c.result.Error(), "got:", c.record.Error())
			t.Fail()
			return
		}
		t.Log("case ", c.memo, "was pass.")
	}
}

func TestOpErrors(t *testing.T) {
	type testCase struct {
		memo   string
		record error
		result error
	}

	cases := []testCase{
		{
			memo:   "error with op",
			record: OpErrors("op").JoinError(errors.New("something wrong")).Error(),
			result: errors.New("there is something error when 'op': caused with: something wrong"),
		},
		{
			memo:   "error with op and record",
			record: OpErrors("op").JoinRecord(libdns.RR{Name: "rec"}, errors.New("something wrong")).Error(),
			result: errors.New("there is something error when 'op': caused at record named 'rec': something wrong"),
		},
		{
			memo:   "error with none error",
			record: OpErrors("op").Error(),
			result: nil,
		},
	}

	for _, c := range cases {
		if (c.record != nil && c.result != nil) && c.record.Error() != c.result.Error() {
			t.Log("excepted:", c.result.Error(), "got:", c.record.Error())
			t.Fail()
			return
		}
		t.Log("case ", c.memo, "was pass.")
	}
}
