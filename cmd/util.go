package cmd

func valueOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
