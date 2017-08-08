package internal

type State uint8

const (
	StateFollower = iota
	StateCandidate
	StateLeader
)
