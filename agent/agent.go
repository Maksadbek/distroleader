package agent

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"sync/atomic"
	"time"

	"github.com/maksadbek/distroleader/internal"
	"github.com/maksadbek/distroleader/internal/kv"
	"github.com/pkg/errors"
)

const (
	maxTimeout = 300
	minTimeout = 150
)

var (
	errVoteDenied   = errors.New("vote denied")
	errAppendFailed = errors.New("append failed")
)

type Agent struct {
	// persistent states
	currentTerm uint64
	votedFor    uint64
	logs        []internal.Log
	kv          kv.Store

	// volatile states for all servers
	//
	commitIndex  uint64
	lastApplied  uint64
	currentState internal.State
	leaderAddr   string
	id           int

	// peers is a slice of peer addresses
	peers []string

	// volatile states for all leaders
	nextIndex  []uint64
	matchIndex []int64

	logger *log.Logger

	internal.Config
}

func (a *Agent) Run() error {
	// start with follower state
	a.currentState = internal.StateFollower
	a.id = rand.Intn(1000)

	err := rpc.Register(a)
	if err != nil {
		return err
	}

	rpc.HandleHTTP()

	// get any random free port
	ln, err := net.Listen("tcp", "")
	if err != nil {
		return err
	}

	// start server
	go func() {
		a.logger.Fatal(http.Serve(ln, nil))
	}()

	a.logger.Printf("Started listening on address: %s", ln.Addr())

	rand.Seed(time.Now().Unix())

	after := make(<-chan time.Time)

	for {
		select {
		case <-after:
			// timeout
			// convert to candidate state, and start election
			a.logger.Printf("timeout, starting election, ID: #%s", ln.Addr())
			a.currentState = internal.StateCandidate

			err := a.StartElection()
			if err != nil {
				// failed election, turn to follower
				a.currentState = internal.StateFollower
				// fallthrough
			}

			// no error, no problem. Won election, convert to leader state
			a.currentState = internal.StateLeader
			a.logger.Printf("won election! converted to a leader, ID: #%s", ln.Addr())
		default:
			// reset timer
			after = time.After(time.Millisecond * time.Duration(rand.Int31n(maxTimeout)))
		}
	}

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
	Entries      []internal.Log
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

// TODO: implement
func (a *Agent) StartElection() error {
	return nil
}

// TODO: implement
// JoinCluster joins to the existing cluster
// receives leader address
//  	leader address must be a address with port number, e.g: 127.0.0.1:59324
// returns error as a status of join
//   	error is nil if joined successfully
func (a *Agent) JoinCluster(addr string) error {
	return nil
}

// SendHeartbeats sends log entry request with empty entry slice to all peers
// if any of them do not respond, start election
func (a *Agent) sendHeartbeats() error {
	for _, peerAddr := range a.peers {
		go func() {
			err := a.sendHeartbeat(peerAddr)
			if err != nil {
				a.logger.Printf("peer #%s not responding", peerAddr)
			}
		}()
	}

	return nil
}

// sendHeartbeat creates a RPC connection between peer
// receives peer address
func (a *Agent) sendHeartbeat(addr string) error {
	timeout := time.After(time.Duration(rand.Int31n(300)) * time.Millisecond)
	ch := make(chan struct{})

	go func() {
		// send empty log entries RPC
	}()

	select {
	case <-timeout:
		return errors.New("timeout")
	case <-ch:
		return nil
	}

	return nil
}

type AddLogRequest struct {
	Op    internal.LogOp
	Key   string
	Value string
}

type AddLogResponse struct {
	Success bool
	Value   string
}

func (a *Agent) AddLog(request AddLogRequest, response *AddLogResponse) error {
	reply := AddLogResponse{}

	// if get operation, get and send value back
	if request.Op == internal.LogOpGet {
		value, err := a.kv.Get(request.Key)
		if err != nil {
			return err
		}

		reply.Success = true
		reply.Value = value

		return nil
	}

	switch a.currentState {
	case internal.StateFollower:
		// redirect request to leader
		client, err := rpc.DialHTTP("tcp", a.leaderAddr)
		if err != nil {
			return err
		}

		err = client.Call("Agent.AddLog", request, &reply)
		if err != nil {
			return err
		}
	case internal.StateLeader:
		// add log and send AppendEntries request
		var index uint64
		if len(a.logs) > 0 {
			index = a.logs[len(a.logs)-1].Index + 1
		} else {
			index = 1
		}

		a.logs = append(a.logs, internal.Log{
			Op:    request.Op,
			Index: index,
			Key:   request.Key,
			Value: request.Value,
			Term:  a.currentTerm,
		})

		var successCount int32
		var wg sync.WaitGroup

		wg.Add(len(a.peers))

		for _, p := range a.peers {
			go func(peerAddr string) {
				defer wg.Done()

				client, err := rpc.DialHTTP("tcp", peerAddr)
				if err != nil {
					a.logger.Println(err)
				}

				newLog := AppendEntriesRequest{
					Term:         a.currentTerm,
					LeaderID:     uint64(a.id),
					PrevLogIndex: 0,
					PrevLogTerm:  0,
					LeaderCommit: a.commitIndex,
					Entries: []internal.Log{
						internal.Log{
							Op:    request.Op,
							Index: index,
							Key:   request.Key,
							Value: request.Value,
							Term:  a.currentTerm,
						},
					},
				}

				reply := AppendEntriesResponse{}
				err = client.Call("Agent.AppendEntries", newLog, &reply)
				if err != nil {
					a.logger.Println(err)
				}

				atomic.AddInt32(&successCount, 1)
			}(p)
		}

		wg.Wait()

		// if majority of servers accepted entry, then commit and flush into KV store
		if len(a.peers)/2+1 <= int(successCount) {
			switch request.Op {
			case internal.LogOpPut:
				err := a.kv.Put(request.Key, request.Value)
				if err != nil {
					return err
				}
			case internal.LogUpDelete:
				err := a.kv.Delete(request.Key)
				if err != nil {
					return err
				}
			default:
				return errors.New("invalid log operation")
			}
		} else {

		}
	case internal.StateCandidate:
	default:
	}

	return nil
}
