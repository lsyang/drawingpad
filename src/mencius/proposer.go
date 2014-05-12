package mencius

//import "fmt"

//A leader could either propose a value or propose Skip
//A non-leader could only propose no-op


//For a non-leader
//Could only propose no-op for that instance
//TODO: make sure that the instance is not suggested or skipped
func (px *Paxos) DriveRevoking(seq int, v interface{}) {
	//fmt.Printf("Replica %v starts to drive revoking for instance %v", px.me, seq)
	maxProposalNo := 0
	length := len(px.peers)
	for !px.dead {
		px.mu.Lock()
		decided := px.statusMap[seq].Done
		current_min := px.min
		px.mu.Unlock()
		if decided || px.dead == true || seq < current_min{
			//fmt.Printf("Replica %v finished Revoke for instance %v... \n", px.me,seq)
		    break
		}

		////////////////////////////
		//phase 1: prepare phase
		//Send pre-accept messages to all other replicas
		////////////////////////////
		px.mu.Lock()
		//fmt.Printf("replica # %v starts to drive revoking for instance %v. px.next_ins= %v", px.me, seq,px.next_ins)
		n := px.GetNextNumber(maxProposalNo)
		prepare_ok_count := 0
		n_a := 0
		var v_a interface{}
		px.mu.Unlock()

		for _, peer := range px.peers {
			prepareargs := &PrepareArgs{SeqNo:seq, ProposalNo:n}
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

     	////////////////////////////
		// accept phase
		////////////////////////////
		if prepare_ok_count > length / 2 {	   
			v_prime := v
		    if v_a != nil {
		        v_prime = v_a
		    }
		    accept_ok_count := 0
		    //fmt.Printf("I am %d, seq instance %d, I am now sending accept request, prepare number %d, value %v\n", px.me, seq, n, v_prime)
		    for _, peer := range px.peers {
		        acceptargs := &AcceptRevokeArgs{seq, n, v_prime}
		        acceptreply := &AcceptRevokeReply{}
		        ok := true
		        if px.peers[px.me] == peer {
		          px.AcceptRevoke(acceptargs, acceptreply)
		        } else {
		          ok = call(peer, "Paxos.AcceptRevoke", acceptargs, acceptreply)
		        }          
		        if ok {
		          if acceptreply.Ok == true {
		            accept_ok_count += 1
		          }
		        }
		    }


			if accept_ok_count > length / 2 { 
            ////////////////////////////////////////////
            // Commit phase			
			/////////////////////////////////////////
				px.mu.Lock()
				MaxDoneSeq :=  px.peersDoneValue[px.me]
				px.mu.Unlock()
				for _, peer := range px.peers {
					decideargs := &DecideArgs{SeqNo:seq, Value:v_prime, Me:px.me, MaxDoneSeq:MaxDoneSeq}
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


//A leader suggests some value for the slot. Don't need to do the prepare phase
func  (px *Paxos) DriveSuggesting(seq int, v interface{}) {
	length := len(px.peers)
	n := 0 // r=0
	for !px.dead{
        px.mu.Lock()
		decided := px.statusMap[seq].Done
		current_min := px.min
		px.mu.Unlock()
		if decided || px.dead || seq < current_min{

		  break
		}
		v_prime := v
	    accept_ok_count := 0
	    //fmt.Printf("I am %d, seq instance %d, I am now suggesting accept request, prepare number %d, value %v\n", px.me, seq, n, v_prime)
	    for index, peer := range px.peers {
	        acceptargs := &AcceptSuggestArgs{seq, n, v_prime}
	        acceptreply := &AcceptSuggestReply{}
	        ok := true
	        if index == px.me {
	          px.AcceptSuggest(acceptargs, acceptreply)
	        } else {
	        	ok = call(peer, "Paxos.AcceptSuggest", acceptargs, acceptreply)
	        }          
	        if ok {
	          if acceptreply.Ok == true {
	            accept_ok_count += 1
	          }
	        }
	    }

        if px.statusMap[seq].Done{
		    break	
		}


		if accept_ok_count > length / 2 { 
        ////////////////////////////////////////////
        // Commit phase			
		/////////////////////////////////////////			
			px.mu.Lock()
			MaxDoneSeq :=  px.peersDoneValue[px.me]
			px.mu.Unlock()
			//fmt.Printf("Suggest accept_ok_count> len/2, sending decide now ... \n")
			for index, peer := range px.peers {
				decideargs := &DecideArgs{SeqNo:seq, Value:v_prime, Me:px.me, MaxDoneSeq:MaxDoneSeq}
				decidereply := &DecideReply{}
				if index==px.me {
					px.Decide(decideargs, decidereply)
				} else {
					call(peer, "Paxos.Decide", decideargs, decidereply)
				}
			}
			
		}       
	}
	//fmt.Printf("finished Suggest for instance %v... \n", seq)	
}

//A leader sends out Skip. v must be a Skip operation
//TODO: how to make sure that when we sends out SKIP message, we would not send out suggest message anymore
func  (px *Paxos) DriveSkipping(seq int, v interface{}) {
    //fmt.Printf("Drive Skipping called by replica # %v for instance %v ... \n", px.me, seq)
    px.mu.Lock()
	decided := px.statusMap[seq].Done
	current_min := px.min
	px.mu.Unlock()
	if decided || px.dead || seq < current_min{
		return
	}

	px.mu.Lock()
	MaxDoneSeq :=  px.peersDoneValue[px.me]
	px.mu.Unlock()

	for _, peer := range px.peers {
		decideargs := &DecideArgs{SeqNo:seq, Value:v, Me:px.me, MaxDoneSeq:MaxDoneSeq}
		decidereply := &DecideReply{}
		if px.peers[px.me] == peer {
			px.Decide(decideargs, decidereply)
		} else {
			call(peer, "Paxos.Decide", decideargs, decidereply)
		}
	}
	//fmt.Printf("finished Skip for instance %v... \n", seq)
}


func (px *Paxos) GetNextNumber(maxsofar int) int {
	length := len(px.peers)
	return maxsofar + length - (maxsofar % length) + px.me
}