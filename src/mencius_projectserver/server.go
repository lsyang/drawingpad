package projectserver

import "net"
import "fmt"
import "net/rpc"
import "log"
import "sync"
import "os"
import "syscall"
import "encoding/gob"
import "math/rand"
import "time"

import "mencius"
import "epaxos"
const Debug=0
func DPrintf(format string, a ...interface{}) (n int, err error) {
  if Debug > 0 {log.Printf(format, a...) }
  return
}

//cache the state of a specific request, store the OperationId and result
type CachedRequestState struct {
  OperationId int64
  Result string
}

type KVPaxos struct { 
  logsMu sync.Mutex //lock when the opLogs is accessed.
  executionMu sync.Mutex //lock when the server actually execute put and get action
  stackMu sync.Mutex //
  
  mu sync.Mutex
  l net.Listener
  me int
  dead bool // for testing
  unreliable bool // for testing
  px *epaxos.Paxos

  stack []Node
  
  opLogs map[int]mencius.Operation
  boardWidth int
  testMapColor map[int]string //map a position (x*boardWidth+y) to color
  maxExecutedOpNum int//the largest Operation that has been executed
  checkDuplicate map[int64]CachedRequestState // mapping clientId of a client to her last executed RPC's CachedRequestState.
  
} 


//Return the  operations that the client misses
func (kv *KVPaxos) GetUpdate(args *GetUpdateArgs, reply *GetUpdateReply) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()	
	ck_newest_op_num := args.Client_newest_op_num
	if (ck_newest_op_num>=kv.maxExecutedOpNum){
		//the client already has the newest log
		reply.Has_operation=false
		return nil
	}
	len := kv.maxExecutedOpNum-ck_newest_op_num
	missed_ops := make([]mencius.Operation,len)
	for i := 0; i < len; i++ {
	  	index := kv.maxExecutedOpNum-(len-1-i)
		  missed_ops[i] = kv.opLogs[index]
	}
	reply.Has_operation=true
	reply.New_operations=missed_ops
	return nil
}


//Return nil if instance is committed
func (kv *KVPaxos) Put(args *PutArgs, reply *PutReply) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	op := mencius.Operation{OpName:"put", ClientStroke:args.ClientStroke,OperationId:args.RequestID, ClientId :args.ClientID,Index:0, Lowlink:0}
	seq := kv.CommitOp(op)
	go kv.ExecuteUntil(seq)
	reply.Err = ""
	//transfer the operations
	ck_newest_op_num := args.Client_newest_op_num
	if (ck_newest_op_num>=kv.maxExecutedOpNum){
		//the client already has the newest log
		reply.Has_operation=false
		return nil
	}
	len := kv.maxExecutedOpNum-ck_newest_op_num
	missed_ops := make([]mencius.Operation,len)
	for i := 0; i < len; i++ {
		index := kv.maxExecutedOpNum-(len-1-i)
		missed_ops[i] = kv.opLogs[index]
	}
	reply.Has_operation=true
	reply.New_operations=missed_ops
	return nil
}

//For testing only
//Need dependency list
func (kv *KVPaxos) Get(args *GetArgs, reply *GetReply) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	stroke :=mencius.Stroke{Start_x:args.Start_x,Start_y:args.Start_y}
	op := mencius.Operation{OpName:"get", ClientStroke:stroke, OperationId:args.RequestID, ClientId :args.ClientID,Index:0, Lowlink:0}
	seq := kv.CommitOp(op)
	val := kv.ExecuteUntil(seq)
	reply.Value=val
	reply.Err = ""
	return nil
}


//Insert a operation and wait for it to be committed
//Rule 4
func (kv *KVPaxos) CommitOp(op mencius.Operation) int {
  for !kv.dead {
	  instance := kv.px.Suggest(op)
    var agreed_op mencius.Operation
    var decided bool
    
    to := 10 * time.Millisecond
    for kv.dead == false {
      var v interface{}
      decided, v = kv.px.Status(instance)
      if decided {
        agreed_op = v.(mencius.Operation)
		    kv.MarkAsCommitted(instance, agreed_op)
        if op.OperationId == agreed_op.OperationId && op.ClientId == agreed_op.ClientId {
          return instance
        }
        break
      }

      if instance < kv.maxExecutedOpNum {
        break
      }
      time.Sleep(to)
      if to<10*time.Second {to*=2}
    }
  }
  return 0
}


//Insert an Nop for instance int
//return the actual operation agreed at that specific sequence number
func (kv *KVPaxos) InsertNop(instance int) mencius.Operation{
  op := mencius.Operation{OpName:"nop", OperationId:-1, ClientId :-1,Index:0, Lowlink:0}
  //original: kv.px.Start(instance, op)
  kv.px.InsertNoOp(instance, op)
  var agreed_op mencius.Operation
  var decided bool
  //waiting for the instance to be decided by paxos
  to := 10 * time.Millisecond
  for !kv.dead {
    var v interface{}
    decided, v = kv.px.Status(instance)
    if decided {
      agreed_op = v.(mencius.Operation)
      kv.MarkAsCommitted(instance, agreed_op)
      return agreed_op
    }
    time.Sleep(to)
    if to<10*time.Second {to *= 2}
  }
  return mencius.Operation{}
}

// update the opLogs of the server 
// and mark the operation with instance number as MarkAsExecuted
func (kv *KVPaxos) MarkAsCommitted(ins_num int, op mencius.Operation){
   kv.logsMu.Lock()
   op.SeqNum=ins_num
   op.Status = "COMMITTED"
   kv.opLogs[ins_num] = op
   kv.logsMu.Unlock()

}


// tell the server to shut itself down.
// please do not change this function.
func (kv *KVPaxos) kill() {
  DPrintf("Kill(%d): die\n", kv.me)
  kv.dead = true
  kv.l.Close()
  kv.px.Kill()
}

//
// servers[] contains the ports of the set of
// servers that will cooperate via Paxos to
// form the fault-tolerant key/value service.
// me is the index of the current server in servers[].
// 
func StartServer(servers []string, me int) *KVPaxos {
  // call gob.Register on structures you want
  // Go's RPC library to marshall/unmarshall.
  gob.Register(mencius.Operation{})

  kv := new(KVPaxos)
  kv.me = me
 
  kv.boardWidth=1000
  kv.testMapColor=make(map[int]string) //map a position (x*boardWidth+y) to color

  if (Store){
  
  kv.opLogs,_ = ReadOpLogs(kv.me)
  kv.maxExecutedOpNum,_ = ReadMaxExecutedOpNum(kv.me)
  kv.checkDuplicate,_ = ReadCachedRequestState(kv.me)
  
  }else{
  kv.checkDuplicate = make(map[int64]CachedRequestState)
  kv.opLogs=make(map[int]mencius.Operation)
  kv.maxExecutedOpNum=-1
  }
 
  kv.stack=	 make([]Node, 0, 100)

  rpcs := rpc.NewServer()
  rpcs.Register(kv)

  kv.px = epaxos.Make(servers, me, rpcs)

  os.Remove(servers[me])
  l, e := net.Listen("unix", servers[me]);
  if e != nil {
    log.Fatal("listen error: ", e);
  }
  kv.l = l
 
  go kv.ExecutionThread()
  // please do not change any of the following code,
  // or do anything to subvert it.
  
  go func() {
    for kv.dead == false {
      conn, err := kv.l.Accept()
      if err == nil && kv.dead == false {
        if kv.unreliable && (rand.Int63() % 1000) < 100 {
          // discard the request.
          conn.Close()
        } else if kv.unreliable && (rand.Int63() % 1000) < 200 {
          // process the request but force discard of reply.
          c1 := conn.(*net.UnixConn)
          f, _ := c1.File()
          err := syscall.Shutdown(int(f.Fd()), syscall.SHUT_WR)
          if err != nil {
            fmt.Printf("shutdown: %v\n", err)
          }
          go rpcs.ServeConn(conn)
        } else {
          go rpcs.ServeConn(conn)
        }
      } else if err == nil {
        conn.Close()
      }
      if err != nil && kv.dead == false {
        fmt.Printf("KVPaxos(%v) accept: %v\n", me, err.Error())
        kv.kill()
      }
    }
  }()

  return kv
}

