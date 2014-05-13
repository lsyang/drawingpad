package mencius

//import "fmt"


type AcceptorState struct {
  N_p int
  N_a int
  V_a interface{}
}

//The prepare message, args.Value could only be a noop

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

//Accept the revoke message. The args.Value could only be noop

func (px *Paxos) AcceptRevoke(args *AcceptRevokeArgs, reply *AcceptRevokeReply) error {
  px.mu.Lock()
  defer px.mu.Unlock()
  SeqNo := args.SeqNo
  ProposalNo := args.ProposalNo
  Value := args.Value
  px.UpdateMax(SeqNo)

  if SeqNo >= px.min {
    //status, exist := px.statusMap[SeqNo] //added
    state, ok := px.acceptorStateMap[SeqNo]
    if (ok && ProposalNo < state.N_p){//  || (exist && status.Done)
        reply.Ok = false
    } else {
        px.acceptorStateMap[SeqNo] = AcceptorState{ProposalNo, ProposalNo, Value}
        reply.Ok = true
    }
  } 
  return nil
}


//Accept the suggest message from a leader of that instance
func (px *Paxos) AcceptSuggest(args *AcceptSuggestArgs, reply *AcceptSuggestReply) error {
  //fmt.Println("Replica #", px.me," receives suggest for instance",args.SeqNo, ", px.next_ins=",  px.next_ins);
  px.mu.Lock()
  defer px.mu.Unlock()
  SeqNo := args.SeqNo  
  ProposalNo := args.ProposalNo
  Value := args.Value
  px.UpdateMax(SeqNo)
   
  ///////////////////////////////////////////////
  ///Mencius Rule 2
  /// Updates next_ins and execute Skips
  if SeqNo > px.next_ins && !px.IsLeader(SeqNo){
      skip := px.next_ins
      //first, increment px.next_ins to be the first instance coordinated by me that is larger than SeqNo
      px.next_ins = SeqNo
      for px.next_ins % px.num_srv != px.me{
          px.next_ins += 1
      }
      // Execute Skip for EVERY that I coordinate in the range of [old_next_ins, new_next_ins)
      for skip < px.next_ins{
          if px.IsLeader(skip){
              seq := skip
              //fmt.Printf("Replica #%v decides to skip instance %v. Old_next_ins is %v, new_next_ins is %v \n", px.me, seq, old_next_ins, px.next_ins)
              px.Skip(seq) //This is the only place that we call Skip
              skip += px.num_srv
          }
      }

  }
  ////////////////////////////////////////////////

  if SeqNo >= px.min {
    // status, exist := px.statusMap[SeqNo] //added //also check StatusMap to see if the instance is already decided
    state, ok := px.acceptorStateMap[SeqNo]
    //only accept suggest message with poposal number higher than n_p 
    if (ok && ProposalNo < state.N_p)  {///|| (exist && status.Done)
        reply.Ok = false
    } else {
        px.acceptorStateMap[SeqNo] = AcceptorState{ProposalNo, ProposalNo, Value}
        reply.Ok = true
       // fmt.Printf("Replica # %v accepts suggest for instance %v, px.next_ins=%v, time \n", px.me, SeqNo, px.next_ins);
    }   
  } 
   // fmt.Printf("Replica # %v accepts suggest for instance %v, px.next_ins=%v \n", px.me, SeqNo, px.next_ins);
  return nil
}



func (px *Paxos) Decide(args *DecideArgs, reply *DecideReply) error {
  px.mu.Lock()
  defer px.mu.Unlock()
  px.UpdateMax(args.SeqNo)
  if args.SeqNo >= px.min {
    px.statusMap[args.SeqNo] = Status{args.Value, true}
  }

  px.peersDoneValue[args.Me] = args.MaxDoneSeq

  leader:= px.getLeader(args.SeqNo)
  px.peersCoordinatedDone[leader] = args.SeqNo

  //TODO: px.CleanMemory()
  //fmt.Printf("I am %d, seq instance %d, I decided the value is %v \n", px.me, args.SeqNo, args.Value)
  reply.Ok = true
  return nil;
}


func (px *Paxos) UpdateMax(seq int) {
  if seq > px.max {
    px.max = seq
  }
  return 
}


func (px *Paxos) getLeader(seq int) int {
   return seq % px.num_srv
}
