package projectserver
//store both paxos log and key value storage on disk
// the key value storage is only a snapshot of the system at previous check stage.
// the current state of the system is the snapshot + operations from paxos log

import (
        "os"
        "encoding/gob"
        "strconv"       
        "log"
        "mencius"
   
)

func WriteToDisk(me int,  anything interface{} ){
    if (Store){
    lock := Lock{}
	lock.mu.Lock()
	defer lock.mu.Unlock() 
	var name string
	switch anything.(type) {

                    case map[int]mencius.Operation:
                           name=OperationLogs
                     case map[int64]CachedRequestState:
                           name=CachedRequest
                      case int:
                           name=MaxExecuted
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


func ReadOpLogs(me int) (map[int]mencius.Operation,bool) { 
    lock := Lock{}
	lock.mu.Lock()
	defer lock.mu.Unlock()  
      f, err := os.Open(OperationLogs+strconv.Itoa(me))
        //file does not exist
        if err != nil {
                os.Create(OperationLogs+strconv.Itoa(me))
                newState:=make(map[int]mencius.Operation)
                WriteToDisk(me,newState)
                return newState,false
        }
        defer f.Close()
        var opLogs map[int]mencius.Operation
        enc := gob.NewDecoder(f)
        if err := enc.Decode(&opLogs); err != nil {
               log.Fatal(err)
                panic("cant decode")
        }
        return opLogs, true        
}


 
func ReadCachedRequestState(me int) (map[int64]CachedRequestState,bool) {  
    lock := Lock{}
	lock.mu.Lock()
	defer lock.mu.Unlock() 
      f, err := os.Open(CachedRequest+strconv.Itoa(me))
        //file does not exist
        if err != nil {
                os.Create(CachedRequest+strconv.Itoa(me))
                newState:=make(map[int64]CachedRequestState)
                WriteToDisk(me,newState)
                return newState,false
        }
        defer f.Close()
        var cachedClientRequest map[int64]CachedRequestState
        enc := gob.NewDecoder(f)
        if err := enc.Decode(&cachedClientRequest); err != nil {
               log.Fatal(err)
                panic("cant decode")
        }
        return cachedClientRequest, true        
}




func ReadMaxExecutedOpNum(me int) (int,bool) {  
    lock := Lock{}
	lock.mu.Lock()
	defer lock.mu.Unlock() 
      f, err := os.Open(MaxExecuted+strconv.Itoa(me))
        if err != nil {
                 os.Create(MaxExecuted+strconv.Itoa(me))
                 WriteToDisk(me,-1)
                return -1,false
        }
        defer f.Close()
        var ins_num  int
        enc := gob.NewDecoder(f)
        if err := enc.Decode(&ins_num); err != nil {
                panic("cant decode")
        }
        return ins_num, true      
}




