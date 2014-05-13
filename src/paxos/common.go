package paxos

import "sync"
const (
  AcceptorStateMap="acceptorStateMap"
  PeersDoneValue="peersDoneValue"
  StatusMap="statusMap"
  Max="max"
  Min="min"
  Store=true
  Interval=50
)
type Lock struct{
   mu sync.Mutex
}


type PrepareArgs struct {
  SeqNo int
  ProposalNo int
  Key int
}

type PrepareReply struct {
  HighestPrepareNo int
  HighestProposalNo int
  Value interface{}
  Ok bool
}

type AcceptArgs struct {
  SeqNo int
  ProposalNo int
  Value interface{}
}

type AcceptReply struct {
  Ok bool
}

type DecideArgs struct {
  SeqNo int
  Value interface{}
  Me int
  MaxDoneSeq int
}

type DecideReply struct {
  Ok bool
}


