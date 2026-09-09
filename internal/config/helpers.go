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
