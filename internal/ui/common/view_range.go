package common

type ViewRange struct {
	*ViewNode
	Start         int
	FirstRowIndex int
	LastRowIndex  int
}
