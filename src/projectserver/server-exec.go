package projectserver

import "time"
//import "fmt"

//A seperate execution thread, one for each server. 
//Keep to see if there is any new instances committed and execute the new instances.
func (kv *KVPaxos) ExecutionThread() {
    num :=kv.maxExecutedOpNum
	for (!kv.dead){
        _,exist:=kv.opLogs[num+1]
		if (!exist){
			time.Sleep(100*time.Millisecond)
		}else{  
			kv.ExecuteUntil(num+1)
			num +=1
		}
   }
}
  
//Execute all the instances until the current instance with instance number ins_num
//Return the result of the last operation.
func (kv *KVPaxos) ExecuteUntil(ins_num int) string {
	kv.executionMu.Lock()
	defer kv.executionMu.Unlock()
	if ins_num <= kv.maxExecutedOpNum { 
		//the operation with instance number ins_num has already been executed before
		return kv.ExecuteOp(ins_num) //Reexecution does not create any problem
	} else { 
		//the operation with instance number ins_num has not been executed before
		val := ""
		for i := kv.maxExecutedOpNum + 1; i <=ins_num; i++ { 
			kv.logsMu.Lock()
			op, alreadyCommitted := kv.opLogs[i] 
			kv.logsMu.Unlock() 
			if !alreadyCommitted {	 
                //wait for it to be decided			
				decided := false
				var v interface{}
				decided, v = kv.px.Status(i)
				if decided {
					op = v.(Operation)
					kv.MarkAsCommitted(i,op)
				} else { //if the operation us neither logged or decided, then we know that there is a hole or the server is far behind. So we add a nop here.
					 op = kv.InsertNop(i)
				}
			}
			//Important: Execute this operation!
			val = kv.findSCC(i)
			//TODO:Flush to disk persistant storage
			//kv.cleanMemory(done) //delete instances up to done		
			kv.maxExecutedOpNum = ins_num
		}	
		return val
	}
}


//Don't need to check everything
//Just execute the operation, mark the execution and return the result.
func (kv *KVPaxos) ExecuteOp(ins_num int) string{
	op :=kv.opLogs[ins_num]
	switch op.OpName {
		case "nop":
			kv.MarkAsExecuted(ins_num)
			return ""
		case "get":
		    x:=op.ClientStroke.Start_x
			y:=op.ClientStroke.Start_y
		    key := x*kv.boardWidth+y
            color:= kv.testMapColor[key]
            request_state, seen_client := kv.checkDuplicate[op.ClientId]
		    if seen_client {
		      if op.OperationId > request_state.OperationId { 
		        //the operation is new, so we should update the client's last RPC information
		        kv.checkDuplicate[op.ClientId] = CachedRequestState{op.OperationId, color}
		      }		      
		    }			       
		  //  op.ClientStroke=Stroke{x,y,0,0,color}
			kv.opLogs[ins_num]=op	
			kv.MarkAsExecuted(ins_num)
	        return color 
		case "put":						
			OperationId := op.OperationId
			//check whether the action is duplicated
			duplicated := false
			result := ""
			request_state, ok := kv.checkDuplicate[op.ClientId]
			if ok {
				if OperationId <= request_state.OperationId { 
					//there is a duplicate, so we should not execute the operation again.
					duplicated = true
					//fmt.Println("set duplicated to true ",OperationId, "<=", request_state.OperationId)
					result = request_state.Result
			    }
			}
			if !duplicated {
			    //update color in testmap
				value := op.ClientStroke.Color
				x:=op.ClientStroke.Start_x
				y:=op.ClientStroke.Start_y
				key := x*kv.boardWidth+y	
				kv.testMapColor[key] = value
				kv.checkDuplicate[op.ClientId] = CachedRequestState{OperationId, value} 
				result=value
				kv.MarkAsExecuted(ins_num)
			}		
			return result
	}
	return ""
}


//Mark the operation with instance number as MarkAsExecuted in the log
func (kv *KVPaxos) MarkAsExecuted(ins_num int){
   kv.logsMu.Lock()
   op:=kv.opLogs[ins_num]
   op.Status="EXECUTED"
   kv.opLogs[ins_num]=op
   kv.logsMu.Unlock()
}


//return true if the operation with seq number has already been MarkAsExecuted
func (kv *KVPaxos) isDone(ins_num int) bool{
	kv.logsMu.Lock()
	defer kv.logsMu.Unlock()
	if(ins_num <= kv.maxExecutedOpNum){
		return true
	}
	return false
}