package projectserver

import "net"
import "fmt"
import "net/rpc"
import "log"
import "epaxos"
import "sync"
import "os"
import "syscall"
import "encoding/gob"
import "math/rand"
import "time"


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
 
  board [1000*1000]string //the current board
  board_snapshot [1000*1000]string
  snapshot_interval int
  
  opLogs map[int]Operation
  //maxCommittedOpNum int //the largest Operation that has been committed
  maxExecutedOpNum int//the largest Operation that has been executed
  checkDuplicate map[int64]CachedRequestState // mapping clientId of a client to her last executed RPC's CachedRequestState.
} 


//Return the boarad_snapshot(optional) and operations that the client misses
func (kv *KVPaxos) GetUpdate(args *GetUpdateArgs, reply *GetUpdateReply) error {
	// kv.mu.Lock()
	// defer kv.mu.Unlock()	
	ck_newest_op_num := args.Client_newest_op_num
	if (ck_newest_op_num>=kv.maxExecutedOpNum){
		//the client already has the newest log
		reply.Has_map=false
		reply.Has_operation=false
		return nil
	}
	//Check if client is one epoch behind of the server
	epoch_diff :=int(ck_newest_op_num/kv.snapshot_interval) - int(kv.maxExecutedOpNum/kv.snapshot_interval)
	if epoch_diff>0{
		reply.Has_map=true
		reply.Board=kv.board_snapshot
		len := kv.maxExecutedOpNum%kv.snapshot_interval
		missed_ops := make([]Operation,len)
		for i := 0; i < len; i++ {  //A simple check: maxExecutedOpNum=2, i=0,1,2, len=3; index=0,1,2
			index := kv.maxExecutedOpNum-(len-1-i)
			missed_ops[i] = kv.opLogs[index]
		}
		reply.Has_operation=true
		reply.New_operations=missed_ops
	}else{
		reply.Has_map=false
		len := kv.maxExecutedOpNum-ck_newest_op_num
		missed_ops := make([]Operation,len)
		for i := 0; i < len; i++ {
			index := kv.maxExecutedOpNum-(len-1-i)
			missed_ops[i] = kv.opLogs[index]
		}
		reply.Has_operation=true
		reply.New_operations=missed_ops
	}
	return nil
}


//Return nil if instance is committed
func (kv *KVPaxos) Put(args *PutArgs, reply *PutReply) error {
	// kv.mu.Lock()
	// defer kv.mu.Unlock()
  op := Operation{OpName:"put", Key:args.Key, Value: args.Value,  OperationId:args.RequestID, ClientId :args.ClientID,Index:0, Lowlink:0}
  seq := kv.CommitOp(op)
  go kv.ExecuteUntil(seq)
  reply.Err = ""
  return nil
}

//For testing only
//Need dependency list
func (kv *KVPaxos) Get(args *GetArgs, reply *GetReply) error {
 kv.mu.Lock()
 defer kv.mu.Unlock()
	
 op := Operation{OpName:"get", Key:args.Key, Value:"", OperationId:args.RequestID, ClientId :args.ClientID,	Index:0, Lowlink:0}
 seq := kv.CommitOp(op)
 
   val := kv.ExecuteUntil(seq)
  /*
   *  opInLogs:=kv.opLogs[seq]
  for (opInLogs.Status!="EXECUTED"){
     opInLogs=kv.opLogs[seq]
     fmt.Println()
  }
  */
   reply.Value=val
   reply.Err = ""
   return nil
}


//Insert an operation, and return the instance number that the operation is at
func (kv *KVPaxos) CommitOp(op Operation) int {
  for kv.dead == false {
    instance := kv.px.Max() + 1 //try to make an agreement using sequence number instance
    kv.px.Start(instance, op, op.Key)
    var agreed_op Operation
    var decided bool
    //waiting for the result
    to := 10 * time.Millisecond
    for kv.dead == false {
      var v interface{}
      decided, v = kv.px.Status(instance)
      if decided {
        agreed_op = v.(Operation)
		kv.MarkAsCommitted(instance, agreed_op)
        if op.OperationId == agreed_op.OperationId && op.ClientId == agreed_op.ClientId {
          return instance
        }
        break
      }
      //should break if this instance number is already filled up
      if instance < kv.maxExecutedOpNum {
        break
      }
      time.Sleep(to)
      if to<10*time.Second {to*=2}
    }
  }
  return 0
}


//Insert an Nop, return the actual operation agreed at that sequence number
func (kv *KVPaxos) InsertNop(instance int) Operation{
  op := Operation{OpName:"nop", Key:-1, Value:"",  OperationId:0, ClientId :0,Index:0, Lowlink:0}
  kv.px.Start(instance, op, op.Key)
  var agreed_op Operation
  var decided bool
  //waiting for the result
  to := 10 * time.Millisecond
  for kv.dead == false {
    var v interface{}
    decided, v = kv.px.Status(instance)
    if decided {
      agreed_op = v.(Operation)
      kv.MarkAsCommitted(instance, agreed_op)
      return agreed_op
    }
    time.Sleep(to)
    if to<10*time.Second {to *= 2}
  }
  return Operation{}
}


////update the opLogs of the server % mark the operation with instance number as MarkAsExecuted
func (kv *KVPaxos) MarkAsCommitted(ins_num int, op Operation){
   kv.logsMu.Lock()
   op.SeqNum=ins_num
   op.Status = "COMMITTED"
   kv.opLogs[ins_num] = op
   kv.logsMu.Unlock()
  // fmt.Printf("Server %v committed instance with ins_num %v \n", kv.me, ins_num)
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
  gob.Register(Operation{})

  kv := new(KVPaxos)
  kv.me = me
  

  // TODO: Your initialization code here.
  //TODO: go kv.ExecutionThread()
  for i:=0;i<1000*1000;i++{
	kv.board[i] = "0"
	kv.board_snapshot[i] = "0"
  }
  kv.snapshot_interval=1000
  kv.opLogs = make(map[int]Operation)
  kv.maxExecutedOpNum = -1
  kv.checkDuplicate = make(map[int64]CachedRequestState)
  
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
 
  // go kv.ExecutionThread()
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

