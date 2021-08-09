package utils

// Unique returns the list of unique strings
func Unique(list []string) (unique []string) {
	keys := make(map[string]struct{})

	for _, item := range list {
		if _, ok := keys[item]; ok == false {
			unique = append(unique, item)
			keys[item] = struct{}{}
		}
	}

	return
}
