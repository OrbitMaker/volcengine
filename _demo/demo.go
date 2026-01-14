package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/libdns/libdns"
	"github.com/libdns/volcenginedns"
)

func main() {
	accessKeyID := strings.TrimSpace(os.Getenv("ACCESS_KEY_ID"))
	accessKeySec := strings.TrimSpace(os.Getenv("ACCESS_KEY_SECRET"))

	if accessKeyID == "" || accessKeySec == "" {
		fmt.Printf("ERROR: %s\n", "ACCESS_KEY_ID or ACCESS_KEY_SECRET missing")
		return
	}

	if len(os.Args) < 2 {
		fmt.Printf("ERROR: %s\n", "First arg ZID (domain ID) or zone (domain name) missing")
		fmt.Printf("Usage: %s <ZID|zone>\n", os.Args[0])
		fmt.Printf("  ZID: number, e.g. 304092\n")
		fmt.Printf("  zone: domain name, e.g. example.com\n")
		return
	}

	arg := strings.TrimSpace(os.Args[1])
	var zid int64
	var zone string

	// Try to parse as ZID (number)
	if parsedZID, parseErr := strconv.ParseInt(arg, 10, 64); parseErr == nil && parsedZID > 0 {
		zid = parsedZID
		fmt.Printf("Using ZID: %d\n", zid)
	} else {
		// Treat as domain name
		zone = arg
		fmt.Printf("Using zone: %s\n", zone)
	}

	provider := volcenginedns.Provider{
		CredentialInfo: volcenginedns.CredentialInfo{
			AccessKeyID:     accessKeyID,
			AccessKeySecret: accessKeySec,
		},
	}
	// Set ZID or Zone (priority: ZID > Zone > zone parameter)
	if zid > 0 {
		provider.ZID = zid
	} else if zone != "" {
		provider.Zone = zone
	}

	// Get all DNS records (if Provider.ZID or Provider.Zone is set, zone parameter can be empty)
	records, err := provider.GetRecords(context.TODO(), zone)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		return
	}

	testID := ""
	testName := "_libdns_test"
	for _, record := range records {
		tmp := record.RR()
		fmt.Printf("%s: %s, %s\n", tmp.Name, tmp.Data, tmp.Type)
		if testName == tmp.Name {
			r, _ := record.(volcenginedns.DomainRecord)
			testID = r.ID
		}
	}

	// Create test record
	if testID == "" {
		fmt.Println("Creating new entry for", testName)
		records, err = provider.AppendRecords(context.TODO(), zone, []libdns.Record{volcenginedns.DomainRecord{
			Type:  "TXT",
			Name:  testName,
			Value: fmt.Sprintf("This+is a test entry created by libdns %s", time.Now()),
			TTL:   600,
		}})
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			return
		}
		if len(records) == 1 {
			tmp, _ := records[0].(volcenginedns.DomainRecord)
			testID = tmp.ID
		}
	}

	// Update test record
	fmt.Println("Press any Key to update the test entry")
	fmt.Scanln()
	if testID != "" {
		fmt.Println("Replacing entry for", testName)
		records, err = provider.SetRecords(context.TODO(), zone, []libdns.Record{volcenginedns.DomainRecord{
			Type:  "TXT",
			Name:  testName,
			Value: fmt.Sprintf("Replacement test entry upgraded by libdns %s", time.Now()),
			TTL:   605,
			ID:    testID,
		}})
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			return
		}
	}

	// Delete test record
	fmt.Println("Press any Key to delete the test entry")
	fmt.Scanln()
	fmt.Println("Deleting the entry for", testName)
	_, err = provider.DeleteRecords(context.TODO(), zone, records)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		return
	}
}
