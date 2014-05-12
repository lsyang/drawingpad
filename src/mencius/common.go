package mencius

import "sync"
import "time"

type Operation struct{
  OpName string
  ClientStroke Stroke
  OperationId int64
  ClientId int64
  SeqNum int
  Dep []int
  Status string
  Index, Lowlink int
}

type Stroke struct{
  Start_x int
  Start_y int
  End_x int
  End_y int
  Color string
  Size int
}


const (
  AcceptorStateMap="acceptorStateMap"
  PeersDoneValue="peersDoneValue"
  StatusMap="statusMap"
  Max="max"
  Min="min"
  Store=false
  Interval= 50 //take a snapshot every 50 instances are decided
  PingInterval = time.Millisecond * 200
  DeadPings = 10
)

type Lock struct{
   mu sync.Mutex
}

type PrepareArgs struct {
  SeqNo int
  ProposalNo int
}

type PrepareReply struct {
  HighestPrepareNo int
  HighestProposalNo int
  Value interface{}
  Ok bool
}

type AcceptRevokeArgs struct {
  SeqNo int
  ProposalNo int
  Value interface{}
}

type AcceptRevokeReply struct {
  Ok bool
}

type AcceptSuggestArgs struct {
  SeqNo int
  ProposalNo int
  Value interface{}
}

type AcceptSuggestReply struct {
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


type PingArgs struct {
  Me int     // "host:port"
}

type PingReply struct {
}

