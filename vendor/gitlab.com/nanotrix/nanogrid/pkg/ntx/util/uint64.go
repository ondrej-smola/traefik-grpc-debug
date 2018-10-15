package util

func CloneUint64Map(a map[uint64]uint64) map[uint64]uint64 {
	if a == nil {
		return nil
	}

	cpy := make(map[uint64]uint64, len(a))

	for i, s := range a {
		cpy[i] = s
	}

	return cpy
}
