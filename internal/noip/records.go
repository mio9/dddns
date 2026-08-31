package noip

import (
	"fmt"
	"strings"
)

func RelativeRecordName(zoneName, recordName string) (string, error) {
	zoneName = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(zoneName)), ".")
	recordName = strings.TrimSpace(recordName)

	if recordName == "" {
		return "", fmt.Errorf("record name is empty")
	}
	if recordName == "@" {
		return "@", nil
	}

	normalizedRecordName := strings.TrimSuffix(strings.ToLower(recordName), ".")
	if normalizedRecordName == zoneName {
		return "@", nil
	}

	zoneSuffix := "." + zoneName
	if strings.HasSuffix(normalizedRecordName, zoneSuffix) {
		relativeName := strings.TrimSuffix(normalizedRecordName, zoneSuffix)
		if relativeName == "" {
			return "@", nil
		}
		return relativeName, nil
	}

	if strings.Contains(normalizedRecordName, ".") {
		return "", fmt.Errorf("record name %q is not in zone %q", recordName, zoneName)
	}

	return normalizedRecordName, nil
}

func DisplayName(zoneName, relativeName string) string {
	if relativeName == "@" {
		return zoneName
	}
	return relativeName + "." + zoneName
}

func RRSetHasValue(rrset *RRSet, ip string) bool {
	for _, rdata := range rrset.Rdata {
		if rdata.Value == ip {
			return true
		}
	}
	return false
}

func ReplaceRdataValues(rrset *RRSet, ip string) []Rdata {
	if rrset == nil || len(rrset.Rdata) == 0 {
		return []Rdata{{Value: ip}}
	}

	replaced := make([]Rdata, len(rrset.Rdata))
	for index, rdata := range rrset.Rdata {
		replaced[index] = Rdata{
			Value: ip,
			Label: rdata.Label,
		}
	}
	return replaced
}

func CurrentValue(rrset *RRSet) string {
	if rrset == nil || len(rrset.Rdata) == 0 {
		return ""
	}
	return rrset.Rdata[0].Value
}
