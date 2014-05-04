package epaxos


type PrepareArgs struct {
  SeqNo int
  ProposalNo int
  Key int
  Deps []int
  SeqNum int
}

type PrepareReply struct {
  HighestPrepareNo int
  HighestProposalNo int
  Value interface{}
  Ok bool
  Deps []int
  SeqNum int
}

type AcceptArgs struct {
  SeqNo int
  ProposalNo int
  Value interface{}
  Deps []int
  SeqNum int
}

type AcceptReply struct {
  Ok bool
}

type DecideArgs struct {
  SeqNo int
  Value interface{}
  Me int
  MaxDoneSeq int
  Deps []int
  SeqNum int
}

type DecideReply struct {
  Ok bool
}


