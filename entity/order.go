package entity

import "slices"

type SortOrder int64

const (
	AscendingOrder SortOrder = iota
	DescendingOrder
	DefaultSortOrder = AscendingOrder
)

var sortOrderKeys = []string{
	"asc",
	"desc",
}

func (o SortOrder) String() string {
	return sortOrderKeys[o]
}

func SortOrderOrDefault(sortOrder string) SortOrder {
	o := slices.Index(sortOrderKeys, sortOrder)
	if o == -1 {
		return DefaultSortOrder
	}

	return SortOrder(o)
}
