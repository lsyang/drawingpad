package EPaxos

import "net/rpc"
import "fmt"
import "time"
import "crypto/rand"
import "math/big"
import "sync"

type Clerk struct {
  mu sync.Mutex // one RPC at a time
  servers []string
  // You will have to modify this struct.
  me int64
  requestID int64
  max_operation_num int
}


func MakeClerk(servers []string) *Clerk {
  ck := new(Clerk)
  ck.servers = servers
  // You'll have to add code here.
  ck.requestID=1
  ck.me=nrand()
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
  var reply GetUpdateReply

  for _, srv := range ck.servers {
    args := &GetUpdateArgs{ck.max_operation_num,ck.me,ck.requestID}
    
    ok := call(srv, "EPaxos.GetUpdate", args, &reply)
    if ok {
      ck.requestID++
      return reply
    }
    time.Sleep(100 * time.Millisecond)
  }
  return reply
}

//
// Put operation by client
//
func (ck *Clerk) Put(key int, value string) PutReply {
  ck.mu.Lock()
  defer ck.mu.Unlock()
  var new_op Operation
  new_op.Key=key
  new_op.Value=value
  var reply PutReply

  for _, srv := range ck.servers {
    args := &PutArgs{ck.max_operation_num,new_op,ck.me,ck.requestID}
    
    ok := call(srv, "EPaxos.Put", args, &reply)
    if ok {
      ck.requestID++
      //check if current op is in reply
      return reply
    }
    time.Sleep(100 * time.Millisecond)
  }
  return reply
}


