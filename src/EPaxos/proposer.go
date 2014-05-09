package epaxos

import "sort"
import "math"

func (px *Paxos) DriveProposing(seq int, v interface{}, op_key int) {
	maxProposalNo := 0
	length := len(px.peers)
	f:= int(math.Floor(float64(length/2))) //tolerated_failier_count F
	seqNum:= 1

	//get seq_lamba and deps_lambda && modify keytoins for tracking dependency
	deps,exist := px.keytoins[op_key]
	if exist {
		sort.Ints(deps)
		seqNum= px.maxSeqNum(deps)+1 
		newSlice := make([]int, len(deps)+1, cap(deps)+1)
		copy(newSlice, deps)
		newSlice[len(deps)]=seq     
		px.keytoins[op_key]=newSlice
	}else{
	    px.keytoins[op_key]=[]int{seq}

	}
	//add the operation to statusMap, i.e. cmds_L[L][i_L]
	status:= Status{value:nil, done:false, Key:op_key, deps:deps, seqNum:seqNum}
	px.statusMap[seq] =status

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
		//Send pre-accept messages to all other replicas in F
		////////////////////////////
		n := px.GetNextNumber(maxProposalNo)
		prepare_ok_count := 0
		n_a := 0
		var v_a interface{}
        equal :=true
		//fmt.Printf("I am %d, seq instance %d, I am now sending prepare request, prepare number %d\n", px.me, seq, n)
        //Todo: update cmds_L[L][i] to accepted			 
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
					//MERGE ATRRIBUTES AND TEST IF THEY ARE SAME
					seqNum,deps,equal = px.mergeAttributes(seqNum, deps, preparereply.SeqNum, preparereply.Deps)
					if preparereply.HighestProposalNo > n_a {
						n_a = preparereply.HighestProposalNo
						v_a = preparereply.Value
					}
				}
			}
		}
     
		if prepare_ok_count >= f+1 {	   
			//get v_prime
			v_prime := v
			if v_a!=nil{ 
			 v_prime = v_a
			}
			//Todo: check if received preaccept_ok from {F}\L, F=length-1 , as well as with the same deps and seqNum
			//fast_path := (prepare_ok_count >= 1+f+int(math.Floor(float64((f+1)/2)))) && equal
			fast_path := (prepare_ok_count >= length) && equal
			fast_path =false
			accept_ok_count := 0		 
			if !fast_path{
				////////////////////////////
				// slow path / optional accept phase
				////////////////////////////
				//fmt.Printf("I am %d, seq instance %d, I am now sending accept request, prepare number %d, value %v\n", px.me, seq, n, v_prime)
				for _, peer := range px.peers {
					acceptargs := &AcceptArgs{ SeqNo:seq,  ProposalNo:n,Value:v_prime, Deps: deps, SeqNum: seqNum}
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
			}
			 
            ////////////////////////////////////////////
            // Commit phase			
			// fast path or accept_ok(n) from majority
			/////////////////////////////////////////
			if fast_path || accept_ok_count >= f+1{
				px.mu.Lock()
				MaxDoneSeq := px.peersDoneValue[px.me]
				px.mu.Unlock()
				//fmt.Printf("%d is decided, I am %d\n", seq, px.me)
				//fmt.Printf("I am %d, seq instance %d, my proposed number is %d, the value is %v \n", px.me, seq, n, v_prime)
				for _, peer := range px.peers {
					decideargs := &DecideArgs{SeqNo:seq, Value:v_prime, Me:px.me,  MaxDoneSeq:MaxDoneSeq,Deps:deps, SeqNum:seqNum}
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

//Return max{ {seqNums in deps} union {0}}
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




