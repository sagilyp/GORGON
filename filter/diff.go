package filter

func DiffIPs(prev, curr []string) (remove, add []string) {
	prevMap := make(map[string]struct{}, len(prev))
	for _, ip := range prev {
		prevMap[ip] = struct{}{}
	}
	currMap := make(map[string]struct{}, len(curr))
	for _, ip := range curr {
		currMap[ip] = struct{}{}
	}
	for ip := range prevMap {
		if _, ok := currMap[ip]; !ok {
			remove = append(remove, ip)
		}
	}
	for ip := range currMap {
		if _, ok := prevMap[ip]; !ok {
			add = append(add, ip)
		}
	}
	return
}
