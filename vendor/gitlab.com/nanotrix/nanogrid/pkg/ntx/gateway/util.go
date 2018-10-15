package gateway

import (
	"sort"
)

func (s Client_State) Terminal() bool {
	return s == Client_FAILED || s == Client_COMPLETED
}

func SortClientsByCompletedDesc(cl []*Client) {
	sort.Slice(cl, func(i, j int) bool {
		if cl[i].Completed == cl[j].Completed {
			return cl[i].Seqn > cl[j].Seqn
		} else {
			return cl[i].Completed > cl[j].Completed
		}
	})
}

func SortClientsByCreatedDesc(cl []*Client) {
	sort.Slice(cl, func(i, j int) bool {
		if cl[i].Created == cl[j].Created {
			return cl[i].Seqn > cl[j].Seqn
		} else {
			return cl[i].Created > cl[j].Created
		}
	})
}
