package config

import "time"

// VietnamTZ is the timezone for Vietnam (UTC+7)
var VietnamTZ = time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)

// ParseDateVN parses a date string in YYYY-MM-DD format using Vietnam timezone
func ParseDateVN(dateStr string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", dateStr, VietnamTZ)
}

// NowVN returns current time in Vietnam timezone
func NowVN() time.Time {
	return time.Now().In(VietnamTZ)
}

// ToVN converts a time to Vietnam timezone
func ToVN(t time.Time) time.Time {
	return t.In(VietnamTZ)
}
