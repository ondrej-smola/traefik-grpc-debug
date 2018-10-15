package util

func CloneStringSlice(a []string) []string {
	if len(a) == 0 {
		return nil
	}

	cpy := make([]string, len(a))

	for i, s := range a {
		cpy[i] = s
	}

	return cpy
}

func CloneStringMap(a map[string]string) map[string]string {
	if a == nil {
		return nil
	}

	cpy := make(map[string]string, len(a))

	for i, s := range a {
		cpy[i] = s
	}

	return cpy
}

func DeduplicateStringSlice(a []string) []string {
	lookup := make(map[string]bool)
	var res []string

	for _, s := range a {
		if !lookup[s] {
			lookup[s] = true
			res = append(res, s)
		}
	}

	return res
}
