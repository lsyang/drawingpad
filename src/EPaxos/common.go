package epaxos


type PrepareArgs struct {
  SeqNo int
  ProposalNo int
  Key int
  SeqNum int
  Deps []int
}

type PrepareReply struct {
  HighestPrepareNo int
  HighestProposalNo int
  Value interface{}
  Ok bool
  SeqNum int
  Deps []int
}

type AcceptArgs struct {
  SeqNo int
  ProposalNo int
  Value interface{}
  SeqNum int
  Deps []int
}

type AcceptReply struct {
  Ok bool
}

type DecideArgs struct {
  SeqNo int
  Value interface{}
  Me int
  MaxDoneSeq int
  SeqNum int
  Deps []int
}

type DecideReply struct {
  Ok bool
}


