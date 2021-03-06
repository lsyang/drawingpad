package paxos

//
// Paxos library, to be included in an application.
// Multiple applications will run, each including
// a Paxos peer.
//
// Manages a sequence of agreed-on values.
// The set of peers is fixed.
// Copes with network failures (partition, msg loss, &c).
// Does not store anything persistently, so cannot handle crash+restart.
//
// The application interface:
//
// px = paxos.Make(peers []string, me string)
// px.Start(seq int, v interface{}) -- start agreement on new instance
// px.Status(seq int) (decided bool, v interface{}) -- get info about an instance
// px.Done(seq int) -- ok to forget all instances <= seq
// px.Max() int -- highest instance seq known, or -1
// px.Min() int -- instances before this seq have been forgotten
//

import "net"
import "net/rpc"
import "log"
import "os"
import "syscall"
import "sync"
import "fmt"
import "math/rand"

type Status struct {
  Value interface{}
  Done bool
}

type Paxos struct {
  mu sync.Mutex
  l net.Listener
  dead bool
  unreliable bool
  rpcCount int
  peers []string
  me int 

  // Your data here.
  acceptorStateMap map[int]AcceptorState  // map each instance to an acceptor state(np,na,va)
  statusMap map[int]Status   //log:map each instance to Status
  max int
  peersDoneValue []int
  min int


}

//The instance with seqNo has been decided
func (px *Paxos) Decide(args *DecideArgs, reply *DecideReply) error {
  px.mu.Lock()
  defer px.mu.Unlock()
  px.UpdateMax(args.SeqNo)
  if args.SeqNo >= px.min {
    px.statusMap[args.SeqNo] = Status{Value:args.Value, Done:true}
    //store relevant paxos state on disk 
    if (args.SeqNo %Interval==0){
        WriteToDisk(px.me,px.acceptorStateMap)
        WriteMin(px.me,px.min)
        WriteMax(px.me,px.max)
        WriteToDisk(px.me,px.peersDoneValue)
        WriteToDisk(px.me, px.statusMap)
    }
        
  }
  px.peersDoneValue[args.Me] = args.MaxDoneSeq
  px.CleanMemory()

  reply.Ok = true
  return nil;
}


//
// the application wants paxos to start agreement on
// instance seq, with proposed value v.
// Start() returns right away; the application will
// call Status() to find out if/when agreement
// is reached.
//
func (px *Paxos) Start(seq int, v interface{}) {
  px.mu.Lock()
  defer px.mu.Unlock()
  px.UpdateMax(seq)
  _, alreadyStart := px.statusMap[seq]
  if !alreadyStart && seq >= px.min {
     px.statusMap[seq]=Status{nil,false}
     go func(){
       px.DriveProposing(seq, v)
     }() 
  }
  return
}



func (px *Paxos) CleanMemory() {
  length := len(px.peersDoneValue)
  min := 0
  if length > 0 {
    min = px.peersDoneValue[0]
    for i := 1; i < length; i++ {
      if px.peersDoneValue[i] < min {
        min = px.peersDoneValue[i]
      }
    }
  }
  px.min = min + 1

  for key, _ := range px.acceptorStateMap {
    if key < px.min {
      delete(px.acceptorStateMap, key)
    }
  }

  for key, _ := range px.statusMap {
    if key < px.min {
      delete(px.statusMap, key)
    }
  }
  return
}

//
// the application on this machine is done with
// all instances <= seq.
//
// see the comments for Min() for more explanation.
//
func (px *Paxos) Done(seq int) {
  px.mu.Lock()
  defer px.mu.Unlock()
  px.peersDoneValue[px.me] = seq
  px.CleanMemory()
  return
}

//
// the application wants to know the
// highest instance sequence known to
// this peer.
//
func (px *Paxos) Max() int {
  px.mu.Lock()
  defer px.mu.Unlock()
  return px.max
}

//
// Min() should return one more than the minimum among z_i,
// where z_i is the highest number ever passed
// to Done() on peer i. A peers z_i is -1 if it has
// never called Done().
//
// Paxos is required to have forgotten all information
// about any instances it knows that are < Min().
// The point is to free up memory in long-running
// Paxos-based servers.
//
// Paxos peers need to exchange their highest Done()
// arguments in order to implement Min(). These
// exchanges can be piggybacked on ordinary Paxos
// agreement protocol messages, so it is OK if one
// peers Min does not reflect another Peers Done()
// until after the next instance is agreed to.
//
// The fact that Min() is defined as a minimum over
// *all* Paxos peers means that Min() cannot increase until
// all peers have been heard from. So if a peer is dead
// or unreachable, other peers Min()s will not increase
// even if all reachable peers call Done. The reason for
// this is that when the unreachable peer comes back to
// life, it will need to catch up on instances that it
// missed -- the other peers therefor cannot forget these
// instances.
// 
func (px *Paxos) Min() int {
  px.mu.Lock()
  defer px.mu.Unlock()
  return px.min
}

//
// the application wants to know whether this
// peer thinks an instance has been decided,
// and if so what the agreed value is. Status()
// should just inspect the local peer state;
// it should not contact other Paxos peers.
//
func (px *Paxos) Status(seq int) (bool, interface{}) {
  px.mu.Lock()
  defer px.mu.Unlock()
  _, ok := px.statusMap[seq]
  if ok {
    return px.statusMap[seq].Done, px.statusMap[seq].Value
  }
  return false, nil
}



//
// the application wants to create a paxos peer.
// the ports of all the paxos peers (including this one)
// are in peers[]. this servers port is peers[me].
//
func Make(peers []string, me int, rpcs *rpc.Server) *Paxos {
  px := &Paxos{}
  px.peers = peers
  px.me = me

  length := len(peers)
  
 //persistent storage
 if (Store){
   px.acceptorStateMap,_=ReadAcceptorStateMap(px.me)
   px.statusMap,_ = ReadStatusMap(px.me)
   px.max,_ = ReadMax(px.me)
   px.min,_=ReadMin(px.me)
   px.peersDoneValue,_=ReadPeersDoneValue(px.me,length)
 }else{
 
   px.acceptorStateMap = make(map[int]AcceptorState)
   px.statusMap = make(map[int]Status)   
   px.max = -1
   px.min = 0
 
   px.peersDoneValue = make([]int, length)
   for i := 0; i < length; i++ {
     px.peersDoneValue[i] = -1
   }
  
 
 }


  if rpcs != nil {
    // caller will create socket &c
    rpcs.Register(px)
  } else {
    rpcs = rpc.NewServer()
    rpcs.Register(px)

    // prepare to receive connections from clients.
    // change "unix" to "tcp" to use over a network.
    os.Remove(peers[me]) // only needed for "unix"
    l, e := net.Listen("unix", peers[me]);
    if e != nil {
      log.Fatal("listen error: ", e);
    }
    px.l = l
    
    // please do not change any of the following code,
    // or do anything to subvert it.
    
    // create a thread to accept RPC connections
    go func() {
      for px.dead == false {
        conn, err := px.l.Accept()
        if err == nil && px.dead == false {
          if px.unreliable && (rand.Int63() % 1000) < 100 {
            // discard the request.
            conn.Close()
          } else if px.unreliable && (rand.Int63() % 1000) < 200 {
            // process the request but force discard of reply.
            c1 := conn.(*net.UnixConn)
            f, _ := c1.File()
            err := syscall.Shutdown(int(f.Fd()), syscall.SHUT_WR)
            if err != nil {
              fmt.Printf("shutdown: %v\n", err)
            }
            px.rpcCount++
            go rpcs.ServeConn(conn)
          } else {
            px.rpcCount++
            go rpcs.ServeConn(conn)
          }
        } else if err == nil {
          conn.Close()
        }
        if err != nil && px.dead == false {
          fmt.Printf("Paxos(%v) accept: %v\n", me, err.Error())
        }
      }
    }()
  }
  return px
}

//
// call() sends an RPC to the rpcname handler on server srv
// with arguments args, waits for the reply, and leaves the
// reply in reply. the reply argument should be a pointer
// to a reply structure.
//
// the return value is true if the server responded, and false
// if call() was not able to contact the server. in particular,
// the replys contents are only valid if call() returned true.
//
// you should assume that call() will time out and return an
// error after a while if it does not get a reply from the server.
//
// please use call() to send all RPCs, in client.go and server.go.
// please do not change this function.
//
func call(srv string, name string, args interface{}, reply interface{}) bool {
  c, err := rpc.Dial("unix", srv)
  if err != nil {
    err1 := err.(*net.OpError)
    if err1.Err != syscall.ENOENT && err1.Err != syscall.ECONNREFUSED {
      fmt.Printf("paxos Dial() failed: %v\n", err1)
    }
    return false
  }
  defer c.Close()
    
  err = c.Call(name, args, reply)
  if err == nil {
    return true
  }

  fmt.Println(err)
  return false
}

//
// tell the peer to shut itself down.
// for testing.
// please do not change this function.
//
func (px *Paxos) Kill() {
  px.dead = true
  if px.l != nil {
    px.l.Close()
  }
}
