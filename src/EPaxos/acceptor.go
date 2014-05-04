package epaxos

import "sort"

type AcceptorState struct {
  n_p int
  n_a int
  v_a interface{}
}

func (px *Paxos) Prepare(args *PrepareArgs, reply *PrepareReply) error {
  px.mu.Lock()
  defer px.mu.Unlock()

  SeqNo := args.SeqNo
  ProposalNo := args.ProposalNo
  Deps :=args.Deps
  SeqNum  :=args.SeqNum  
  op_key :=args.Key
  
  max_seq_num:=1
  deps:=make([]int,10)
  interf,exist := px.keytoins[op_key]
  if (exist){
	  sort.Ints(interf)
	  self_max:= px.maxSeqNum(interf)+1 
	  max_seq_num, deps, _= px.mergeAttributes(self_max, interf, SeqNum, Deps)
  }
  
  state, ok := px.acceptorStateMap[SeqNo]
  px.UpdateMax(SeqNo)
  
  if SeqNo >= px.min {
    if ok{
      if ProposalNo > state.n_p {
        px.acceptorStateMap[SeqNo] = AcceptorState{ProposalNo, state.n_a, state.v_a}
        
        reply.HighestPrepareNo = state.n_p
        reply.HighestProposalNo = state.n_a
        reply.Value = state.v_a
        reply.Ok = true
        reply.Deps=deps
        reply.SeqNum=max_seq_num     
      } else {
        reply.HighestPrepareNo = state.n_p
        reply.HighestProposalNo = state.n_a
        reply.Value = state.v_a
        reply.Ok = false
      }
    } else {
      px.acceptorStateMap[SeqNo] = AcceptorState{ProposalNo, 0, nil}     
      
      reply.HighestPrepareNo = 0
      reply.HighestProposalNo = 0
      reply.Value = nil
      reply.Ok = true
      reply.Deps=deps
      reply.SeqNum=max_seq_num
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
    if ok && ProposalNo < state.n_p {
      reply.Ok = false
    } else {
      px.acceptorStateMap[SeqNo] = AcceptorState{ProposalNo, ProposalNo, Value}
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


//Take the max of seq1 and seq2
//Take the union of two lists deps1 and deps2
//Check if two maxs and two lists are strictly equal
func (px *Paxos) mergeAttributes(seq1 int, deps1 []int, seq2 int, deps2 []int) (int, []int, bool) {
	equal := true
	if seq1 != seq2 {
		equal = false
		if seq2 > seq1 {
			seq1 = seq2
		}
	}
	length := len(deps2)
	
	loop_list := deps2
	keep_list := deps1
	    
	if len(deps1)>len(deps2){
	    equal =false
	}else{
	     if len(deps1)<len(deps2) {equal=false}
    	 length = len(deps1)
	     loop_list = deps1
	     keep_list = deps2
    }
	for i := 0; i < length; i++ {
		if deps1[i] != deps2[i] {
			equal = false
		}
		valueExist:=false
		a:=loop_list[i]
		for _, b := range keep_list {
        	if b == a { valueExist= true}
   		}   		
		if !valueExist{
			 equal=false
	 newSlice := make([]int, len(keep_list)+1, cap(keep_list)+1)
     copy(newSlice, keep_list)
     newSlice[len(keep_list)]=loop_list[i] 
     
		   keep_list=newSlice
		   
		}
	}
	sort.Ints( keep_list)
	
	return seq1,  keep_list, equal
}

/*
func equal(deps1 []int, deps2 []int) bool {
	for i := 0; i < len(deps1); i++ {
		if deps1[i] != deps2[i] {
			return false
		}
	}
	return true
}
*/
