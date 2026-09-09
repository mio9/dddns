package config

func findProviderByType(providers []Provider, providerType string) Provider {
	index := findProviderIndexByType(providers, providerType)
	if index < 0 {
		return Provider{}
	}
	return providers[index]
}

func findProviderIndexByType(providers []Provider, providerType string) int {
	for index, provider := range providers {
		if provider.Type == providerType {
			return index
		}
	}
	return -1
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func mergeRecords(existing, incoming []Record) []Record {
	merged := append([]Record{}, existing...)
	for _, record := range incoming {
		index := findRecordIndex(merged, record)
		if index >= 0 {
			merged[index] = record
			continue
		}
		merged = append(merged, record)
	}
	return merged
}

func findRecordIndex(records []Record, target Record) int {
	for index, record := range records {
		if recordsMatch(record, target) {
			return index
		}
	}
	return -1
}

func recordsMatch(left, right Record) bool {
	if left.ID != "" && right.ID != "" && left.ID == right.ID {
		return true
	}
	if left.Name == "" || right.Name == "" {
		return false
	}
	leftType := left.Type
	if leftType == "" {
		leftType = "A"
	}
	rightType := right.Type
	if rightType == "" {
		rightType = "A"
	}
	return left.Name == right.Name && leftType == rightType
}
