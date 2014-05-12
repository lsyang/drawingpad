package epaxos
//store both paxos log and key value storage on disk
// the key value storage is only a snapshot of the system at previous check stage.
// the current state of the system is the snapshot + operations from paxos log

import (
        "os"
        "encoding/gob"
        "strconv"       
        "log"
)


//save the input argument to disk as a file
func WriteToDisk(me int,  anything interface{} ){
 if (Store){
    lock := Lock{}
	lock.mu.Lock()
	defer lock.mu.Unlock() 
	var name string
	switch anything.(type) {
                    case map[int]Status:
                            name=StatusMap
                    case map[int]AcceptorState:
                            name=AcceptorStateMap
                    case []int:
                            name=PeersDoneValue
    }
 	f, err := os.Create(name+strconv.Itoa(me))
        if err != nil {
                 log.Fatal(err)
        }
        defer f.Close()
        enc := gob.NewEncoder(f)
        if err := enc.Encode(anything); err != nil {
                panic("cant encode")
        }
     }
}

//Read statusMap from disk. If the file does not exist, 
//create the file and store the initialized map
func ReadStatusMap(me int) (map[int]Status,bool) {  
    lock := Lock{}
	lock.mu.Lock()
	defer lock.mu.Unlock() 
        f, err := os.Open(StatusMap+strconv.Itoa(me))
        //file does not exist
        if err != nil {
                os.Create(StatusMap+strconv.Itoa(me))
                newState:=make(map[int]Status)
                WriteToDisk(me,newState)
                return newState,false
        }
        defer f.Close()
        var statusMap map[int]Status
        enc := gob.NewDecoder(f)
        if err := enc.Decode(&statusMap); err != nil {
                panic("cant decode")
        }
        return statusMap, true        
}


func ReadAcceptorStateMap(me int) (map[int]AcceptorState,bool) {  
        lock := Lock{}
	lock.mu.Lock()
	defer lock.mu.Unlock() 
        f, err := os.Open(AcceptorStateMap+strconv.Itoa(me))
        //file does not exist
        if err != nil {
                os.Create(AcceptorStateMap+strconv.Itoa(me))
                newState:=make(map[int]AcceptorState)
                WriteToDisk(me,newState)
                return newState,false
        }
        defer f.Close()
        var acceptorStateMap map[int]AcceptorState
        enc := gob.NewDecoder(f)
        if err := enc.Decode(&acceptorStateMap); err != nil {
                panic("cant decode")
        }
        return acceptorStateMap, true        
}

func ReadPeersDoneValue(me int, length int) ([]int,bool) {  
        lock := Lock{}
	lock.mu.Lock()
	defer lock.mu.Unlock() 
        f, err := os.Open(PeersDoneValue+strconv.Itoa(me))
        if err != nil {
                os.Create(PeersDoneValue+strconv.Itoa(me))
                newState:=make([]int,length)
                for i := 0; i < length; i++ {
    				newState[i] = -1
  				}
                WriteToDisk(me,newState)
                return newState,false
        }
        defer f.Close()
        var val []int
        enc := gob.NewDecoder(f)
        if err := enc.Decode(&val); err != nil {
                panic("cant decode")
        }
        return val, true        
}

func WriteMin( me int,  val int){
    lock := Lock{}
	lock.mu.Lock()
	defer lock.mu.Unlock() 
	writeMaxMin(me,val,false)
}

func WriteMax( me int,  val int){
    lock := Lock{}
	lock.mu.Lock()
	defer lock.mu.Unlock() 
	writeMaxMin(me, val,true)
}
func writeMaxMin( me int,  val int, max bool){
  	var name string
  	if (max){
     	name=Max
  	}else{
  		name=Min
  	}
 	f, err := os.Create(name+strconv.Itoa(me))
        if err != nil {
              //     log.Fatal(err)
        }
        defer f.Close()
        enc := gob.NewEncoder(f)
        if err := enc.Encode(val); err != nil {
                panic("cant encode")
        }
}
func ReadMin( me int)(int, bool){
    lock := Lock{}
	lock.mu.Lock()
	defer lock.mu.Unlock() 
	return readMaxMin(me,false)
}

func ReadMax( me int) (int, bool){
    lock := Lock{}
	lock.mu.Lock()
	defer lock.mu.Unlock() 
	return readMaxMin(me,true)
}

func readMaxMin(me int,isMax bool) (int,bool) {  
  	var name string
  	if (isMax){
     	name=Max
  	}else{
  		name=Min
  	}
      f, err := os.Open(name+strconv.Itoa(me))
        if err != nil {
                 os.Create(name+strconv.Itoa(me))
                 if (isMax){
                   WriteMax(me,-1)
                    return -1,false
                 }else{
                  WriteMin(me,0)
                    return 0,false
                 }
             
        }
        defer f.Close()
        var val  int
        enc := gob.NewDecoder(f)
        if err := enc.Decode(&val); err != nil {
                panic("cant decode")
        }
        return val, true      
}
