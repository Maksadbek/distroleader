package agent

import (
	"fmt"
	"log"

	"github.com/maksadbek/distroleader/internal"
	"github.com/pkg/errors"
)

var (
	errVoteDenied   = errors.New("vote denied")
	errAppendFailed = errors.New("append failed")
)

type Agent struct {
	// persistent states
	currentTerm uint64
	votedFor    uint64
	logs        []Log

	// volatile states for all servers
	//
	commitIndex  uint64
	lastApplied  uint64
	currentState internal.State

	// peers is a slice of peer addresses
	peers []string

	// volatile states for all leaders
	nextIndex  []uint64
	matchIndex []int64

	logger *log.Logger
}

// request vote
//
type RequestVoteRequest struct {
	Term         uint64
	CandidateID  uint64
	LastLogIndex uint64
	LastLogTerm  uint64
}

type RequestVoteResponse struct {
	Term        uint64
	VoteGranted bool
}

func (a *Agent) RequestVote(request RequestVoteRequest, response *RequestVoteResponse) error {
	a.logger.Printf("agent: RequestVote: request: %v", request)
	err := fmt.Errorf("vote denied, candidate: %d, candidate term: %d, current term: %d", request.CandidateID, request.Term, a.currentTerm)

	// if term < currentTerm, reply false
	if request.Term < a.currentTerm {
		return err
	}

	// if votedFor is null or candidateID, and candidate's log is at least as up-to-date as receiver's log, grant vote
	if (a.votedFor == 0 || request.CandidateID == 0) && (request.LastLogIndex >= a.lastApplied && request.LastLogTerm >= a.currentTerm) {
		response.VoteGranted = true
		response.Term = a.currentTerm
		return nil
	}

	return err
}

// append entries
//
type AppendEntriesRequest struct {
	Term         uint64
	LeaderID     uint64
	PrevLogIndex uint64
	PrevLogTerm  uint64
	LeaderCommit uint64
	Entries      []Log
}

type AppendEntriesResponse struct {
	Term    uint64
	Success bool
}

func (a *Agent) AppendEntries(request AppendEntriesRequest, response *AppendEntriesResponse) error {
	a.logger.Printf("agent: AppendEntries: request: %v", request)

	// reply false if term < currentTerm
	if a.currentTerm < request.Term {
		return errors.Wrapf(errAppendFailed,
			"invalid leader term: leader term: %d, current term: %d",
			request.Term, a.currentTerm,
		)
	}

	// reply false if log doesn't contain an entry at prevLogIndex whose term mastches perLogTerm
	var contains bool

	for _, l := range a.logs {
		if l.Index == request.PrevLogIndex && l.Term == request.PrevLogTerm {
			contains = true
		}
	}

	if !contains {
		return errors.Wrapf(errAppendFailed, "invalid leader PrevLogIndex and PrevLogTerm")
	}

	// if an existing entry conflicts with a new one(same index but different terms)
	// delete the existing entry and all that follow it.
	for i := range a.logs {
		if (a.logs[i].Index == request.Entries[i].Index) && a.logs[i].Term != request.Entries[i].Term {
			a.logs = a.logs[:i]
		}

		a.logs = append(a.logs, request.Entries...)
	}

	if request.LeaderCommit > a.commitIndex {
		if request.LeaderCommit < a.logs[len(a.logs)-1].Index {
			a.commitIndex = request.LeaderCommit
		} else {
			a.commitIndex = a.logs[len(a.logs)-1].Index
		}
	}
	return nil
}
