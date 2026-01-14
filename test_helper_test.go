package volcengine

// Test credentials
// Note: Replace with real AccessKeyID and AccessKeySecret for actual testing
const (
	TestAccessKeyID     = "<Input your AccessKeyID here>"
	TestAccessKeySecret = "<Input your AccessKeySecret here>"
)

// Test Provider instance
// Shared by all test files
// Can specify the domain to test by setting ZID or Zone
var p0 = Provider{
	CredentialInfo: CredentialInfo{
		AccessKeyID:     TestAccessKeyID,
		AccessKeySecret: TestAccessKeySecret,
	},
	// ZID: 304092,  // Can directly set domain ID
	// Zone: "example.com",  // Or set domain name, will automatically look up ZID via ListZones API
}
