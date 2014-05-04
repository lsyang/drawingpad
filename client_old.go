package projectserver

import "net/rpc"
import "fmt"
import "time"
import "crypto/rand"
import "math/big"

type Clerk struct {
  servers []string
  operationId int64
  clientId int64
}


func MakeClerk(servers []string) *Clerk {
  ck := new(Clerk)
  ck.servers = servers
  ck.clientId = nrand()
  ck.operationId = 0
  return ck
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

// generate a random 64-bit number
func nrand() int64 {
  max := big.NewInt(int64(1) << 62)
  bigx, _ := rand.Int(rand.Reader, max)
  x := bigx.Int64()
  return x
}

//
// fetch the current value for a key.
// returns "" if the key does not exist.
// keeps trying forever in the face of all other errors.
//
func (ck *Clerk) Get(key int) string {
  operationId := ck.operationId
  //increment the operationId to be the next one
  ck.operationId += 1
  args := &GetArgs{key, ck.clientId, operationId}
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
// set the value for a key.
// keeps trying until it succeeds.
//
func (ck *Clerk) PutExt(key int, value string, dohash bool) string {
  operationId := ck.operationId
  //increment the operationId to be the next one
  ck.operationId += 1
  args := &PutArgs{0,key, value, ck.clientId, operationId}
  for {
    //try sending request for all the servers
    for _, server := range ck.servers {
      reply := &PutReply{}
      ok := call(server, "KVPaxos.Put", args, reply)
      if ok == true && reply.Err == "" {
        return ""
      }
    }
    time.Sleep(time.Second)
  }
}

func (ck *Clerk) Put(key int, value string) {
  ck.PutExt(key, value, false)
}
func (ck *Clerk) PutHash(key int, value string) string {
  v := ck.PutExt(key, value, true)
  return v
}
