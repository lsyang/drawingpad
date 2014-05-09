package projectserver

import "net/rpc"
import "fmt"
import "time"
import "crypto/rand"
import "math/big"
import "sync"
// import "log"
// import "net/http"
// import "encoding/json"
// import "strconv"


type Clerk struct {
  mu sync.Mutex // one RPC at a time
  servers []string
  // You will have to modify this struct.
  me int64
  requestID int64
  max_operation_num int
  keys []int
  // putChan chan Stroke 
  // getChan chan GetUpdateReply
}


func MakeClerk(servers []string) *Clerk {
  ck := new(Clerk)
  ck.servers = servers
  // You'll have to add code here.
  ck.requestID=1
  ck.me=nrand()
  // ck.putChan=make(chan Stroke, 10) //at most 10 ops
  // ck.getChan=make(chan GetUpdateReply, 0)
  // go ck.Put()
  // go ck.GetUpdate()
  return ck
}

func nrand() int64 {
  max := big.NewInt(int64(1) << 62)
  bigx, _ := rand.Int(rand.Reader, max)
  x := bigx.Int64()
  return x
}
//
// call() sends an RPC to the rpcname handler on server srv
// with arguments args, waits for the reply, and leaves the
// reply in reply. the reply argument should be a pointer
// to a reply structure.
//
// the return value is true if the server responded, and false
// if call() was not able to contact the server. in particular,
// the reply's contents are only valid if call() returned true.
//
// you should assume that call() will time out and return an
// error after a while if it doesn't get a reply from the server.
//
// please use call() to send all RPCs, in client.go and server.go.
// please don't change this function.
//
func call(srv string, rpcname string,
          args interface{}, reply interface{}) bool {
  c, errx := rpc.Dial("unix", srv)
  if errx != nil {
    return false
  }
  defer c.Close()
    
  err := c.Call(rpcname, args, reply)
  if err == nil {
    return true
  }

  fmt.Println(err)
  return false
}

//
// Get update from the server
//
//func (ck *Clerk) GetUpdate() (bool bool map[int]int []Operation) {
func (ck *Clerk) GetUpdate() GetUpdateReply{
  ck.mu.Lock()
  defer ck.mu.Unlock()

  for{
    var reply GetUpdateReply
    args := &GetUpdateArgs{ck.max_operation_num,ck.me,ck.requestID}
    for _, srv := range ck.servers {
      ok := call(srv, "KVPaxos.GetUpdate", args, &reply)
      if ok {
        //ck.getChan <- reply
        // fmt.Println("%v, got an reply", ck.me)
        ck.requestID++
        operations:=reply.New_operations
        if reply.Has_operation{
          //fmt.Println("%v, has operation" , ck.me)
         ck.max_operation_num=operations[len(operations)-1].SeqNum
        }
        //break
        return reply
        //time.Sleep(80 * time.Millisecond) //goes faster and throws away empty ops
      }
    }
    time.Sleep(80 * time.Millisecond)
    //fmt.Println("here")
  }
  //return reply
}

func (ck *Clerk) Get(start_x int,start_y int) string {
  operationId := ck.requestID
  //increment the operationId to be the next one
  ck.requestID += 1
  args := &GetArgs{start_x,start_y, ck.me, operationId}
  for {
    //try sending request for all the servers
    for _, server := range ck.servers {
      reply := &GetReply{}
      ok := call(server, "KVPaxos.Get", args, reply)
      if ok == true && reply.Err == "" {
        return reply.Value 
      }
    }
    time.Sleep(time.Second)
  }
}

//
// Put operation by client
//
func (ck *Clerk) Put(op Stroke) {
  ck.mu.Lock()
  defer ck.mu.Unlock()
  
  for{
    var reply PutReply
   // op:= <-ck.putChan //if no op, will stuck right? 
    //fmt.Println("%v, there is an op", ck.me)
    args := &PutArgs{ck.max_operation_num,op,ck.me,ck.requestID}
    for _, srv := range ck.servers {
      
      ok := call(srv, "KVPaxos.Put", args, &reply)
      if ok {
        ck.requestID++
        return 
        //break
      }
    }
    time.Sleep(100 * time.Millisecond)
  }
}

// func (ck *Clerk) PutChan(op Stroke){
//   //fmt.Println("here!")
//   ck.putChan <- op
// }
// func (ck *Clerk) GetChan() GetUpdateReply{
//   //fmt.Println("here!")
//   reply:=<-ck.getChan //wait until it gets something?
//   //fmt.Println("callinhg on get")
//   return reply
// }


