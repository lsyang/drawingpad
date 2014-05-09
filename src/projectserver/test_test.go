package projectserver

import "testing"
import "runtime"
import "strconv"
import "os"
import "time"
import "fmt"
import "math/rand"

// func check(t *testing.T, ck *Clerk, start int, value Stroke) {
//   v := ck.GetUpdate()
//   if v.Has_operation{
//     newop:=v.New_operations[len(v.New_operations)-1].ClientStroke
    

//     if equal(newop, value) {
//      t.Fatalf("Getupdate() -> %v, expected %v", newop, value)
//      //fmt.Printf("Get(%v) -> %v, expected %v", start,end, v, value)
//     }
//     }else{
//       t.Fatalf("Getupdate() -> is empty , expected %v", value)
//     }
// }

func check(t *testing.T, ck *Clerk, start int,end int, value string) {
  v := ck.Get(start,end)
  
  if v != value {
   t.Fatalf("Get(%v, %v) -> %v, expected %v", start,end, v, value)
   //fmt.Printf("Get(%v) -> %v, expected %v", start,end, v, value)
  }
}

func equal(op1 Stroke, op2 Stroke) bool{
  if op1.Start_x!=op2.Start_x || op1.Start_y!=op2.Start_y || op1.End_x!=op2.End_x || op1.End_y!=op2.End_y || op1.Color!=op2.Color || op1.Size!=op2.Size{
    return false
  }
  return true
}

func port(tag string, host int) string {
  s := "/var/tmp/824-"
  s += strconv.Itoa(os.Getuid()) + "/"
  os.Mkdir(s, 0777)
  s += "kv-"
  s += strconv.Itoa(os.Getpid()) + "-"
  s += tag + "-"
  s += strconv.Itoa(host)
  return s
}

func cleanup(kva []*KVPaxos) {
  for i := 0; i < len(kva); i++ {
    if kva[i] != nil {
      kva[i].kill()
    }
  }
}

func TestBasic(t *testing.T) {
  runtime.GOMAXPROCS(4)

  const nservers = 3
  var kva []*KVPaxos = make([]*KVPaxos, nservers)
  var kvh []string = make([]string, nservers)
  defer cleanup(kva)

  for i := 0; i < nservers; i++ {
    kvh[i] = port("basic", i)
  }
  for i := 0; i < nservers; i++ {
    kva[i] = StartServer(kvh, i)
  }

  ck := MakeClerk(kvh)
  var cka [nservers]*Clerk
  for i := 0; i < nservers; i++ {
    cka[i] = MakeClerk([]string{kvh[i]})
  }

  fmt.Printf("Test: Basic put/getupdate ...\n")
  op:=Stroke{1,10,20,30,"aa",10}

  ck.Put(op)
  check(t, ck, 1, 10, "aa")

  op=Stroke{1,10,20,30,"aaa",10}
  cka[1].Put(op)

  check(t, cka[2], 1, 10, "aaa")
  check(t, cka[1], 1, 10, "aaa")
  check(t, ck, 1, 10, "aaa")

  fmt.Printf("  ... Passed\n")
   time.Sleep(1 * time.Second)
  }


func TestUnreliable(t *testing.T) {
  runtime.GOMAXPROCS(4)

  const nservers = 3
  var kva []*KVPaxos = make([]*KVPaxos, nservers)
  var kvh []string = make([]string, nservers)
  defer cleanup(kva)

  for i := 0; i < nservers; i++ {
    kvh[i] = port("un", i)
  }
  for i := 0; i < nservers; i++ {
    kva[i] = StartServer(kvh, i)
    kva[i].unreliable = true
  }
  
  ck := MakeClerk(kvh)
 
  var cka [nservers]*Clerk
  for i := 0; i < nservers; i++ {
    cka[i] = MakeClerk([]string{kvh[i]})
  }


  fmt.Printf("Test: Basic put/get, unreliable ...\n")
op:=Stroke{2,2,20,30,"aa",10}
  ck.Put(op)
  check(t, ck, 2,2, "aa")

op=Stroke{2,2,20,30,"aaa",10}
  cka[1].Put(op)

  check(t, cka[2], 2,2, "aaa")
  check(t, cka[1], 2,2, "aaa")
  check(t, ck,2, 2,"aaa")

  fmt.Printf("  ... Passed\n")

  fmt.Printf("Test: Sequence of puts, unreliable ...\n")

  for iters := 0; iters < 6; iters++ {
  const ncli = 5
    var ca [ncli]chan bool
    for cli := 0; cli < ncli; cli++ {
      ca[cli] = make(chan bool)
      go func(me int) {
        ok := false
        defer func() { ca[me] <- ok }()
        sa := make([]string, len(kvh))
        copy(sa, kvh)
        for i := range sa {
          j := rand.Intn(i+1)
          sa[i], sa[j] = sa[j], sa[i]
        }
        myck := MakeClerk(sa)
        key := me
        op:=Stroke{key,key,20,30,"0",10}
        myck.Put(op)
        pv := myck.Get(key,key)
        if pv!="0" {
          t.Fatalf("wrong value; expected %s but got %s", pv, "0")
        }
        op=Stroke{key,key,20,30,"1",10}
        myck.Put(op)
        pv = myck.Get(key,key)
        if pv != "1" {
          t.Fatalf("wrong value; expected %s but got %s", pv, "1")
        }
        op=Stroke{key,key,20,30,"2",10}
        myck.Put(op)
     

        time.Sleep(100 * time.Millisecond)
        if myck.Get(key,key) != "2" {
          t.Fatalf("wrong value")
        }

        ok = true
      }(cli)
    }
    for cli := 0; cli < ncli; cli++ {
      x := <- ca[cli]
      if x == false {
        t.Fatalf("failure")
      }
    }
  }

  fmt.Printf("  ... Passed\n")

  fmt.Printf("Test: Concurrent clients, unreliable ...\n")

  for iters := 0; iters < 15; iters++ {
    const ncli = 15
    var ca [ncli]chan bool
    for cli := 0; cli < ncli; cli++ {
      ca[cli] = make(chan bool)
      go func(me int) {
        defer func() { ca[me] <- true }()
        sa := make([]string, len(kvh))
        copy(sa, kvh)
        for i := range sa {
          j := rand.Intn(i+1)
          sa[i], sa[j] = sa[j], sa[i]
        }
        myck := MakeClerk(sa)
        if (rand.Int() % 1000) < 500 {
          op:=Stroke{3,3,20,30,strconv.Itoa(rand.Int()),10}
          myck.Put(op)
        } else {
          myck.Get(3,3)
        }
      }(cli)
    }
    for cli := 0; cli < ncli; cli++ {
      <- ca[cli]
    }

    var va [nservers]string
    for i := 0; i < nservers; i++ {
      va[i] = cka[i].Get(3,3)
      if va[i] != va[0] {
        t.Fatalf("mismatch; 0 got %v, %v got %v", va[0], i, va[i])
      }
    }
  }

  fmt.Printf("  ... Passed\n")

time.Sleep(1 * time.Second)
}

func SingleCrash(t *testing.T) {
  runtime.GOMAXPROCS(4)

  const nservers = 3
  var kva []*KVPaxos = make([]*KVPaxos, nservers)
  var kvh []string = make([]string, nservers)
  defer cleanup(kva)

  for i := 0; i < nservers; i++ {
    kvh[i] = port("basic", i)
  }
  for i := 0; i < nservers; i++ {
    kva[i] = StartServer(kvh, i)
  }

  //ck := MakeClerk(kvh)
  var cka [nservers]*Clerk
  for i := 0; i < nservers; i++ {
    cka[i] = MakeClerk([]string{kvh[i]})
  }

  fmt.Printf("Test: Concurrent clients ...\n")

  for iters := 0; iters < 20; iters++ {
    const npara = 15
    var ca [npara]chan bool
    for nth := 0; nth < npara; nth++ {
      ca[nth] = make(chan bool)
      go func(me int) {
        defer func() { ca[me] <- true }()
        ci := (rand.Int() % nservers)
        myck := MakeClerk([]string{kvh[ci]})
        if (rand.Int() % 1000) < 500 {
          op:=Stroke{2,2,20,30,strconv.Itoa(rand.Int()),10}
            kva[0].kill() //kill server 0
  kva[1].kill() //kill server 0
  kva[2].kill() //kill server 0
          myck.Put(op)
          
        } else {
          myck.Get(2,2)
        }
      }(nth)
    }
    for nth := 0; nth < npara; nth++ {
      <- ca[nth]
    }
    kvh[0] = port("basic", 0)  //restarting server
    kva[0] = StartServer(kvh, 0)
    var va [nservers]string
    for i := 0; i < nservers; i++ {
      va[i] = cka[i].Get(2,2)
      if va[i] != va[0] {
        t.Fatalf("mismatch, va[0] is ", va[0], "BUT va[i] is ", va[i])
      }
    } 
  }

  fmt.Printf("  ... Passed\n")
  time.Sleep(1 * time.Second)
}



