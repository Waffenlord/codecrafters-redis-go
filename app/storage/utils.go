package storage

func lcp(first string, second string) int {
	i := 0
	for i < len(first) && i < len(second) && first[i] == second[i] {
		i++
	}
	return i
}
