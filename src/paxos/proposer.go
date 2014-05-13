package paxos


func (px *Paxos) DriveProposing(seq int, v interface{}) {
	maxProposalNo := 0
	length := len(px.peers)
	for {
		px.mu.Lock()
		decided := px.statusMap[seq].Done
		current_min := px.min
		px.mu.Unlock()
		if decided || px.dead == true || seq < current_min{
		  break
		}

		////////////////////////////
		//phase 1: prepare phase
		//Send prepare messages to all other replicas in F
		////////////////////////////
		n := px.GetNextNumber(maxProposalNo)
		prepare_ok_count := 0
		n_a := 0
		var v_a interface{}
		 
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
     
		if prepare_ok_count > length/2 {	   
			//get v_prime
			v_prime := v
			if v_a!=nil{ 
			 v_prime = v_a
			}
			accept_ok_count := 0		 
		
		////////////////////////////
		// slow path / optional accept phase
		////////////////////////////
		for _, peer := range px.peers {
			acceptargs := &AcceptArgs{ SeqNo:seq,  ProposalNo:n,Value:v_prime, }
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
	
			 
            ////////////////////////////////////////////
            // Commit phase			

	/////////////////////////////////////////
	if accept_ok_count > length/2{
		px.mu.Lock()
		MaxDoneSeq := px.peersDoneValue[px.me]
		px.mu.Unlock()
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





