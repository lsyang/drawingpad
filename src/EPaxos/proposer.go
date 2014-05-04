package epaxos

import "sort"

func (px *Paxos) DriveProposing(seq int, v interface{}, op_key int) {
  maxProposalNo := 0
  length := len(px.peers)
  seqNum:= 1
  var exist bool
  deps :=make([]int,10)
  deps,exist = px.keytoins[op_key]
  if exist {
	  sort.Ints(deps)
	 seqNum= px.maxSeqNum(deps)+1
  }
  px.statusMap[seq] = Status{value:nil, done:false, Key:op_key, deps:deps, seqNum:seqNum}

  for {
    //fmt.Printf("I am %d, seq instance %d, I am now starting a job :)\n", px.me, seq)
    px.mu.Lock()
    decided := px.statusMap[seq].done
    current_min := px.min
    px.mu.Unlock()
    if decided || px.dead == true || seq < current_min{
      break
    }

    ////////////////////////////
    //phase 1: prepare phase
    ////////////////////////////
    n := px.GetNextNumber(maxProposalNo)
    prepare_ok_count := 0
    n_a := 0
    var v_a interface{}

    //fmt.Printf("I am %d, seq instance %d, I am now sending prepare request, prepare number %d\n", px.me, seq, n)
    //var AllDeps
    
    for _, peer := range px.peers {
      prepareargs := &PrepareArgs{SeqNo:seq, ProposalNo:n, Deps: deps, SeqNum: seqNum}
      preparereply := &PrepareReply{}
      ok := true 
      if px.peers[px.me] == peer {
        px.Prepare(prepareargs, preparereply)
      } else {
        ok = call(peer, "Paxos.Prepare", prepareargs, preparereply)
      }
      if ok {
        if preparereply.HighestPrepareNo > maxProposalNo {
            maxProposalNo = preparereply.HighestPrepareNo
        } 
        if preparereply.Ok == true {
            prepare_ok_count += 1
            
            if preparereply.HighestProposalNo > n_a {
	            n_a = preparereply.HighestProposalNo
	            v_a = preparereply.Value
            }
        }
      }
    }


    if prepare_ok_count > length / 2 {
       //get v_prime
       v_prime := v
       if v_a!=nil{ //original: if v_a!=nil
         v_prime = v_a
       }
       //check if received preaccept_ok from {F}\L, F=length-1 , as well as with the same deps and seqNum 
       
	     ////////////////////////////
	    //phase 2: accept phase
	    ////////////////////////////
        //slow path

	      accept_ok_count := 0
	      //fmt.Printf("I am %d, seq instance %d, I am now sending accept request, prepare number %d, value %v\n", px.me, seq, n, v_prime)
	      for _, peer := range px.peers {
	        acceptargs := &AcceptArgs{ SeqNo:seq,  ProposalNo:n,Value:v_prime}
	        acceptreply := &AcceptReply{}
	        ok := true
	        if px.peers[px.me] == peer {
	          px.Accept(acceptargs, acceptreply)
	        } else {
	          ok = call(peer, "Paxos.Accept", acceptargs, acceptreply)
	        }          
	        if ok {
	          if acceptreply.Ok == true {
	            accept_ok_count += 1
	          }
	        }
	      }
	      
	      //accept_ok(n) from majority. Commit!
	      if accept_ok_count > length / 2 {
	        px.mu.Lock()
	        MaxDoneSeq := px.peersDoneValue[px.me]
	        px.mu.Unlock()
	        //fmt.Printf("%d is decided, I am %d\n", seq, px.me)
	        //fmt.Printf("I am %d, seq instance %d, my proposed number is %d, the value is %v \n", px.me, seq, n, v_prime)
	        for _, peer := range px.peers {
	          decideargs := &DecideArgs{SeqNo:seq, Value:v_prime, Me:px.me,  MaxDoneSeq:MaxDoneSeq}
	          decidereply := &DecideReply{}
	          if px.peers[px.me] == peer {
	            px.Decide(decideargs, decidereply)
	          } else {
	            call(peer, "Paxos.Decide", decideargs, decidereply)
	          }
	        }
	      }
	}
  }
}



func (px *Paxos) GetNextNumber(maxsofar int) int {
  length := len(px.peers)
  return maxsofar + length - (maxsofar % length) + px.me
}


func (px *Paxos) maxSeqNum(deps []int) int {
   max :=0
  for _,ins_num:=range(deps){
     seq := px.statusMap[ins_num].seqNum
     if seq>max{
     	max=seq
     }
  }
  return max
}




