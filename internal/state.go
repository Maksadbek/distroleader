package internal

type State uint8

const (
	StateFollower State = iota
	StateCandidate
	StateLeader
)
