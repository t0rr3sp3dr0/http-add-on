package main

// getHostCount gets proper count for given host regardless whether
// host is in counts or only in routerTable
func getHostCount(
	host string,
	counts map[string]int,
) (int, bool) {
	if count, exists := counts[host]; exists {
		return count, true
	}

	return 0, false
}
