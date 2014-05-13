package mencius

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
import "time"
import "encoding/gob"


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
  me int // index into peers[]

  // Your data here.
  acceptorStateMap map[int]AcceptorState
  statusMap map[int]Status  
  max int
  peersDoneValue []int
  min int

  // Added for mencius
  next_ins int //index of the server. At initialization, is me (the index of server in peers).
  num_srv int 
  recentPingTime []time.Time  //index to the last ping time from each server
  peersCoordinatedDone []int //the minimum instance that is coordinated by q and learned by this server
}


//
// the application wants server to suggest value v.
// Return the instance number
func (px *Paxos) Suggest(v interface{}) int{
  px.mu.Lock()
  defer  px.mu.Unlock()
  seq :=px.next_ins
  px.UpdateMax(seq)
  _, alreadyStart := px.statusMap[seq]
  if !alreadyStart && seq >= px.min {// Should not be already started. Since the server only suggests an instance once
     px.statusMap[seq] = Status{nil, false}
     px.next_ins += px.num_srv
    //Let the server suggest the value
    //fmt.Printf("Replica %v. px.next_ins=%v.... there are %v replicas\n", px.me, px.next_ins, px.num_srv)
    go func(){
      px.DriveSuggesting(seq, v)
    }()
  }
  return seq
}


// Try to insert a noop at instance with number seq
func (px *Paxos) InsertNoOp(seq int, v interface{}){
  //fmt.Printf("%v .InsertNoOp( %v) called. \n", px.me, seq)
  px.mu.Lock()
  defer px.mu.Unlock()
  px.UpdateMax(seq)
  _, alreadyStart := px.statusMap[seq]
  //fmt.Printf("%v do %v has statusMap[seq]. %v>=? %v \n", px.me, alreadyStart,seq,px.min)
  if !alreadyStart && seq >= px.min {// Should not be already started. Since the server only suggests an instance once
     //fmt.Printf("Condition met \n")
     px.statusMap[seq] = Status{nil, false}
     //fmt.Printf("before go DriveRevoking() \n")
     //Let the server revoke
     go func(){
       //fmt.Printf("inside go func() \n")
       px.DriveRevoking(seq,v) //Operation{OpName:"nop", OperationId:-1, ClientId :0}
     }()
  }
  return
}

//Let the leader Skips at instance with number seq
//Don't need to hold lock since the only place that call this is in accepting suggest message, already holds the lock
func (px *Paxos) Skip(seq int){
  px.UpdateMax(seq)
  _, alreadyStart := px.statusMap[seq]
  if !alreadyStart && seq >= px.min {// Should not be already started. Since the server only suggests an instance once
     px.statusMap[seq] = Status{nil, false}   
     go  func(){
        op := Operation{OpName:"nop", OperationId:-1, ClientId :-1,Index:0, Lowlink:0}
        px.DriveSkipping(seq, op)
     }()
  }
  return 
}


//suggest for the sequence only if seq=px.next_ins
func (px *Paxos) Start(seq int, v interface{}) {
  px.mu.Lock()
  defer px.mu.Unlock()
  px.UpdateMax(seq)
  _, alreadyStart := px.statusMap[seq]
  gob.Register(Operation{})
  //fmt.Println(px.me, "receives ", seq," my next_ins is ",px.next_ins)
  if !alreadyStart && seq >= px.min && seq==px.next_ins {
      px.statusMap[seq] = Status{nil, false}
      px.next_ins = px.next_ins + px.num_srv
      go func(){
         px.DriveSuggesting(seq, v)  
      }()
  }
  return
}


//Added for mencius: check if I am the leader for instance seq
func (px *Paxos) IsLeader(seq int) bool{
    return (seq % px.num_srv == px.me)
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
    //fmt.Printf("seq number is %d, the result is %t, I am %d\n", seq, px.statusMap[seq].done, px.me)
    return px.statusMap[seq].Done, px.statusMap[seq].Value
  }
  //fmt.Printf("seq number is %d, I come here weirdly, I am %d\n", seq, px.me)
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


  // Your initialization code here.~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
  px.acceptorStateMap = make(map[int]AcceptorState)
  px.statusMap = make(map[int]Status)
  px.max = -1
  px.min = 0
  length := len(peers)
  px.peersDoneValue = make([]int, length)
  px.next_ins = me
  px.num_srv = len(peers)
  px.recentPingTime = make([]time.Time, length)
  px.peersCoordinatedDone= make([]int, length)
  for i :=0; i <length; i++{
       px.recentPingTime[i]=time.Now()
  }

  for i := 0; i < length; i++ {
    px.peersDoneValue[i] = -1
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

    // create a thread to call tick() periodically.
  go func() {
    for !px.dead {
      px.tick()
      time.Sleep(PingInterval)

    }
  }()
  //create a thread to check periodically
  go func() {
     for !px.dead {
       px.CheckAlive()
       k := (1+px.me) *100
       time.Sleep(PingInterval + time.Duration(k) * time.Millisecond  )
     }
   }()
  return px
}

// tick() is called once per PingInterval; it should send Ping RPC to all servers,
func (px *Paxos) tick() {
  for _, peer := range px.peers {
    args := &PingArgs{Me:px.me}
    var reply PingReply
    if px.peers[px.me] == peer{
       px.Ping(args, &reply)
    }else{
       call(peer, "Paxos.Ping", args, &reply)
    }
  }
} 

// server Ping RPC handler:update the recent ping time map
func (px *Paxos) Ping(args *PingArgs, reply *PingReply) error {
  px.mu.Lock()
  defer px.mu.Unlock()
  px.recentPingTime[args.Me] = time.Now()
  return nil
}


//Rule 3: revoke q for all instances in the range [Cq,Ip] that q coordinates.
//Where Cq is the smallest instance that is coordinated by q and not learned by p(self)
func (px *Paxos) CheckAlive(){
   for server, _ := range px.recentPingTime {
      if time.Now().After(px.recentPingTime[server].Add(10* PingInterval) ) {
         //Rule 3
         //fmt.Printf("Replica # %v is considered dead by replica # %v", server, px.me)
         i := px.peersCoordinatedDone[server] + px.num_srv
         for i <= px.next_ins-px.num_srv {
            op := Operation{OpName:"nop", OperationId:-1, ClientId :-1,Index:0, Lowlink:0}
            px.InsertNoOp(i,op) 
            i += px.num_srv
        }
      }
   }
}


/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////       Do not change below       ////////////////////////////////////////////////
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

//
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

  //fmt.Printf("So Now my min is %d, I am %d\n", px.min, px.me)

  return
}
