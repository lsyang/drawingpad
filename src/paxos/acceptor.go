package paxos


type AcceptorState struct {
  N_p int
  N_a int
  V_a interface{}
}

func (px *Paxos) Prepare(args *PrepareArgs, reply *PrepareReply) error {
  px.mu.Lock()
  defer px.mu.Unlock()
  SeqNo := args.SeqNo
  ProposalNo := args.ProposalNo
  state, ok := px.acceptorStateMap[SeqNo]
  px.UpdateMax(SeqNo)
  
  if SeqNo >= px.min {
    if ok{
      if ProposalNo > state.N_p {
        px.acceptorStateMap[SeqNo] = AcceptorState{ProposalNo, state.N_a, state.V_a}       
        reply.HighestPrepareNo = state.N_p
        reply.HighestProposalNo = state.N_a
        reply.Value = state.V_a
        reply.Ok = true    
      } else {
        reply.HighestPrepareNo = state.N_p
        reply.HighestProposalNo = state.N_a
        reply.Value = state.V_a
        reply.Ok = false
      }
    } else {
      px.acceptorStateMap[SeqNo] = AcceptorState{ProposalNo, 0, nil}   
 
      reply.HighestPrepareNo = 0
      reply.HighestProposalNo = 0
      reply.Value = nil
      reply.Ok = true
    }
  }
  return nil
}


func (px *Paxos) Accept(args *AcceptArgs, reply *AcceptReply) error {
  px.mu.Lock()
  defer px.mu.Unlock()

  SeqNo := args.SeqNo
  ProposalNo := args.ProposalNo
  Value := args.Value
  px.UpdateMax(SeqNo)
  if SeqNo >= px.min {
    state, ok := px.acceptorStateMap[SeqNo]
    if ok && ProposalNo < state.N_p {
        reply.Ok = false
    } else {
        px.acceptorStateMap[SeqNo] = AcceptorState{ProposalNo, ProposalNo, Value}
    if (SeqNo%Interval==0){
        WriteToDisk(px.me,px.acceptorStateMap)
          WriteMin(px.me,px.min)
            WriteMax(px.me,px.max)
             WriteToDisk(px.me,px.peersDoneValue)
             WriteToDisk(px.me, px.statusMap)
        }
        reply.Ok = true
    }
  } 
  return nil
}



func (px *Paxos) UpdateMax(seq int) {
  if seq > px.max {
    px.max = seq
  }
  
  return 
}



