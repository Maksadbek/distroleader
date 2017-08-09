package internal

type LogOp uint8

const (
	_ = iota
	LogOpPut
	LogOpGet
	LogUpDelete
)

type Log struct {
	Op    LogOp
	Index uint64
	Key   string
	Value interface{}

	// term when entry was received by leader, first index is 1
	Term uint64
}
