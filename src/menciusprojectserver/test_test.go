package projectserver

import "testing"
import "runtime"
import "strconv"
import "os"
import "time"
import "fmt"
import "math/rand"
import "mencius"

//Check a single stroke
func check(t *testing.T, ck *Clerk, start int,end int, value string) {
  v := ck.Get(start,end)
  
  if v != value {
   t.Fatalf("Get(%v) -> %v, expected %v", start,end, v, value)
   //fmt.Printf("Get(%v) -> %v, expected %v", start,end, v, value)
  }
}

//Check the history of clients to see if it equals the given list of strokes
//Skip noop operations
func checkHistory(t *testing.T, ck *Clerk, correct_strokes []mencius.Stroke) {
   v := ck.GetUpdateFrom(-1)
   if v.Has_operation{
      missed_ops:= v.New_operations
      j := 0
      for i := 0; i < len(missed_ops); i++{
           if (j <len(correct_strokes)){
	          client_stroke :=missed_ops[i].ClientStroke 
              correct_stroke :=correct_strokes[j]
              if !isNoOp(client_stroke){	
                  j=j+1    
                  if !strokeEqual(client_stroke, correct_stroke) { 
			          t.Fatalf("Getupdate()'s stroke #%v, expected %v, got %v", j, correct_stroke, client_stroke)
		           }               
	          }
          }
      }
     // if j!=len(correct_strokes){
       //  t.Fatalf("Getupdate()'s return %v noop stroke, should be %v strokes", j, len(correct_strokes))
    //  }
   }
}

//Check whether the history of two clients is a prefix of one another
func checkClientHisEqual(t *testing.T, ck1 *Clerk, ck2 *Clerk){
   update1 := ck1.GetUpdateFrom(-1)
   update2 := ck2.GetUpdateFrom(-1)
   if update1.Has_operation && update2.Has_operation{
       v1 := update1.New_operations
       v2 := update2.New_operations
       //First,check if the length of two lists are equal
       v :=v1
       if len(v1)>=len(v2){
           v=v2
       }
       for i := 0; i < len(v); i++{
           s1 :=v1[i].ClientStroke 
           s2 :=v2[i].ClientStroke 
           if !strokeEqual(s1, s2) { 
			    t.Fatalf("Two clients have different history, for %v instance, one had %v, one had %v", i, s1, s2)
		   }       
       }    
   }
}


//test if two strokes are equal
func strokeEqual(s1 mencius.Stroke, s2 mencius.Stroke) bool {
  result := (s1.Start_x==s2.Start_x && s1.Start_y==s2.Start_y && s1.End_x==s2.End_x && s1.End_y==s2.End_y && s1.Color==s2.Color && s1.Size==s2.Size)
  return result
}

func isNoOp(s1 mencius.Stroke) bool {
  result := (s1.Start_x==0 && s1.Start_y==0 && s1.End_x==0 && s1.End_y==0 && s1.Color=="" && s1.Size==0)
  return result
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
 // defer deleteStorage(nservers)
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

  fmt.Printf("Test: Basic put/puthash/get ...\n")
   
  s1 := mencius.Stroke{1,1,3,5, "aa",1}
  s2 := mencius.Stroke{1,1,3,5, "aaa",13}
  s3 := mencius.Stroke{1,1,3,5, "b",1}
  l1 := make([]mencius.Stroke,1)
  l2 := make([]mencius.Stroke,2)
  l3 := make([]mencius.Stroke,3)
  l4 := make([]mencius.Stroke,4)
  ck.Put(s1)
  time.Sleep(1000*time.Millisecond)
  l1[0]=s1
  l2[0]=s1
  l3[0]=s1
  l4[0]=s1
  checkHistory(t, ck, l1)
  
  cka[1].Put(s2)
  time.Sleep(1000*time.Millisecond)
  l2[1]=s2
  l3[1]=s2
  l4[1]=s2
  checkHistory(t, ck, l2)
  checkHistory(t, cka[1], l2)
  checkHistory(t, cka[2], l2)
  
  cka[2].Put(s3)
  l3[2]=s3
  l4[2]=s3
  time.Sleep(1000*time.Millisecond)
   
  checkHistory(t, ck, l3) 
  checkHistory(t, cka[1],l3) 
  checkHistory(t, cka[2],l3)

  fmt.Printf("  ... Passed\n")
  
  fmt.Printf("Test: A new client join ...\n")

  new_client := MakeClerk(kvh)
  time.Sleep(100 * time.Millisecond)
  checkHistory(t, new_client, l3) 

  fmt.Printf("  ... Passed\n")


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
          myck.Put(mencius.Stroke{2,2,3,5, strconv.Itoa(rand.Int()),1})
        } else {
          myck.Get(2,2)
        }
      }(nth)
    }
    for nth := 0; nth < npara; nth++ {
      <- ca[nth]
    }
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



func TestDifferentBasic(t *testing.T) {
  time.Sleep(2 * time.Second)
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

  fmt.Printf("Test: Basic put/get for different indices ...\n")

  ck.Put(mencius.Stroke{1,2, 10,10,"aa",1})
  check(t, ck, 1,2, "aa")

  ck.Put(mencius.Stroke{2, 2, 10,10,"aaa",1})
  check(t, ck, 2,2, "aaa")
  check(t, cka[1], 2,2, "aaa")
  check(t, cka[2],2,2, "aaa")

  ck.Put(mencius.Stroke{3,3,10,10, "b",1})
  check(t, ck, 3,3, "b") 
  check(t, cka[2], 3,3, "b")
  check(t, cka[1], 3,3, "b") 

  time.Sleep(1 * time.Second)
  checkClientHisEqual(t, cka[1], cka[2])  
  checkClientHisEqual(t, ck, cka[1])
  checkClientHisEqual(t, ck, cka[2]) 

  fmt.Printf("  ... Passed\n")

  time.Sleep(1 * time.Second)
}


func pp(tag string, src int, dst int) string {
  s := "/var/tmp/824-"
  s += strconv.Itoa(os.Getuid()) + "/"
  s += "kv-" + tag + "-"
  s += strconv.Itoa(os.Getpid()) + "-"
  s += strconv.Itoa(src) + "-"
  s += strconv.Itoa(dst)
  return s
}

func cleanpp(tag string, n int) {
  for i := 0; i < n; i++ {
    for j := 0; j < n; j++ {
      ij := pp(tag, i, j)
      os.Remove(ij)
    }
  }
}

func part(t *testing.T, tag string, npaxos int, p1 []int, p2 []int, p3 []int) {
  cleanpp(tag, npaxos)

  pa := [][]int{p1, p2, p3}
  for pi := 0; pi < len(pa); pi++ {
    p := pa[pi]
    for i := 0; i < len(p); i++ {
      for j := 0; j < len(p); j++ {
        ij := pp(tag, p[i], p[j])
        pj := port(tag, p[j])
        err := os.Link(pj, ij)
        if err != nil {
          t.Fatalf("os.Link(%v, %v): %v\n", pj, ij, err)
        }
      }
    }
  }
}


func TestPartition(t *testing.T) {
time.Sleep(2 * time.Second)
  runtime.GOMAXPROCS(4)

  tag := "partition"
  const nservers = 5
  var kva []*KVPaxos = make([]*KVPaxos, nservers)

  defer cleanup(kva)
  defer cleanpp(tag, nservers)
    
  for i := 0; i < nservers; i++ {
    var kvh []string = make([]string, nservers)
    for j := 0; j < nservers; j++ {
      if j == i {
        kvh[j] = port(tag, i)
      } else {
        kvh[j] = pp(tag, i, j)
      }
    }
    kva[i] = StartServer(kvh, i)
  }
  defer part(t, tag, nservers, []int{}, []int{}, []int{})

  var cka [nservers]*Clerk
  for i := 0; i < nservers; i++ {
    cka[i] = MakeClerk([]string{port(tag, i)})
  }

  fmt.Printf("Test: No partition ...\n")

  part(t, tag, nservers, []int{0,1,2,3,4}, []int{}, []int{})
  cka[0].Put(mencius.Stroke{1,2, 10,10,"12",1})
  cka[2].Put(mencius.Stroke{1, 2, 10,10,"13",1})
  check(t, cka[3], 1, 2,"13")
  
  time.Sleep(2*time.Second)
  checkClientHisEqual(t, cka[1], cka[2])  
  checkClientHisEqual(t, cka[0], cka[1])
  checkClientHisEqual(t, cka[0], cka[2]) 

  fmt.Printf("  ... Passed\n")

  fmt.Printf("Test: Progress in majority ...\n")

  part(t, tag, nservers, []int{2,3,4}, []int{0,1}, []int{})
  cka[2].Put(mencius.Stroke{1,2, 10,10,"14",1})
  check(t, cka[4], 1, 2,"14")

  fmt.Printf("  ... Passed\n")

  fmt.Printf("Test: No progress in minority ...\n")

  done0 := false
  done1 := false
  go func() {
    cka[0].Put(mencius.Stroke{1, 2, 10,10,"15",1})
    done0 = true
  }()
  go func() {
    cka[1].Get(1,2)
    done1 = true
  }()
  time.Sleep(time.Second)
  if done0 {
    t.Fatalf("Put in minority completed")
  }
  if done1 {
    t.Fatalf("Get in minority completed")
  }
  check(t, cka[4], 1, 2,"14")
  cka[3].Put(mencius.Stroke{1, 2, 10, 10, "16", 1} )
  check(t, cka[4], 1,2, "16")

  fmt.Printf("  ... Passed\n")
  //2, 3, 4 has 16, 0,1, has 12


  /*
  fmt.Printf("Test: Completion after heal ...\n")

  part(t, tag, nservers, []int{0,2,3,4}, []int{1}, []int{})
  for iters := 0; iters < 30; iters++ {
    if done0 {
      break
    }
    time.Sleep(100 * time.Millisecond)
  }
  if done0 == false {
    t.Fatalf("Put did not complete")
  }
  if done1 {
    t.Fatalf("Get in minority completed")
  }

  check(t, cka[4], 1, 2,"15")

  check(t, cka[0], 1,2, "15")
 
  part(t, tag, nservers, []int{0,1,2}, []int{3,4}, []int{})
  for iters := 0; iters < 100; iters++ {
    if done1 {
      break
    }
    time.Sleep(100 * time.Millisecond)
  }
  if done1 == false {
    t.Fatalf("Get did not complete")
  }
     
  check(t, cka[1], 1,2, "15")

  fmt.Printf("  ... Passed\n")
*/
}


func TestUnreliable(t *testing.T) {
time.Sleep(2 * time.Second)
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

  ck.Put(mencius.Stroke{2, 2, 10, 10, "aa",1})
  check(t, ck, 2,2, "aa")

  cka[1].Put(mencius.Stroke{2,2, 10, 10,"aaa",1})

  check(t, cka[2], 2,2, "aaa")
  check(t, cka[1], 2,2, "aaa")
  check(t, ck,2, 2,"aaa")

  fmt.Printf("  ... Passed\n")
  /*
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

        myck.Put(mencius.Stroke{key,key,10,10, "0",1})
        pv := myck.Get(key,key)
        if pv!="0" {
          t.Fatalf("wrong value; expected %s but got %s", pv, "0")
        }
        myck.Put(mencius.Stroke{key, key,10,10, "1",1})
        pv = myck.Get(key,key)
        if pv != "1" {
          t.Fatalf("wrong value; expected %s but got %s", pv, "1")
        }
        myck.Put(mencius.Stroke{key,key,10,10,  "2",1})
     

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
          myck.Put(mencius.Stroke{3, 3, 10,10, strconv.Itoa(rand.Int()),1})
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
*/

}

/*
func TestHole(t *testing.T) {
time.Sleep(2 * time.Second)
  runtime.GOMAXPROCS(4)

  fmt.Printf("Test: Tolerates holes in paxos sequence ...\n")

  tag := "hole"
  const nservers = 5

  var kva []*KVPaxos = make([]*KVPaxos, nservers)
  defer cleanup(kva)
  defer cleanpp(tag, nservers)

  for i := 0; i < nservers; i++ {
    var kvh []string = make([]string, nservers)
    for j := 0; j < nservers; j++ {
      if j == i {
        kvh[j] = port(tag, i)
      } else {
        kvh[j] = pp(tag, i, j)
      }
    }
    kva[i] = StartServer(kvh, i)
  }
  defer part(t, tag, nservers, []int{}, []int{}, []int{})

  for iters := 0; iters < 5; iters++ {
    part(t, tag, nservers, []int{0,1,2,3,4}, []int{}, []int{})

    ck2 := MakeClerk([]string{port(tag, 2)})
    ck2.Put(mencius.Stroke{ 15, 2, 10, 10, "q",1})

    done := false
    const nclients = 10
    var ca [nclients]chan bool
    for xcli := 0; xcli < nclients; xcli++ {
      ca[xcli] = make(chan bool)
      go func(cli int) {
        ok := false
        defer func() { ca[cli] <- ok }()
        var cka [nservers]*Clerk
        for i := 0; i < nservers; i++ {
          cka[i] = MakeClerk([]string{port(tag, i)})
        }
        key := cli
        last := ""
        cka[0].Put(mencius.Stroke{key,key, 10,10,last,1})
        for done == false {
          ci := (rand.Int() % 2)
          if (rand.Int() % 1000) < 500 {
            nv := strconv.Itoa(rand.Int())
            cka[ci].Put(mencius.Stroke{key,key,10,10, nv,1})
            last = nv
          } else {
            v := cka[ci].Get(key,key)
            if v != last {
              t.Fatalf("%v: wrong value, key %v, wanted %v, got %v",
                cli, key, last, v)
            }
          }
        }
        ok = true
      } (xcli)
    }

    time.Sleep(3 * time.Second)

    part(t, tag, nservers, []int{2,3,4}, []int{0,1}, []int{})

    // can majority partition make progress even though
    // minority servers were interrupted in the middle of
    // paxos agreements?
    check(t, ck2, 15, 2,"q")
    ck2.Put(mencius.Stroke{15,2, 10,10, "qq", 1})
    check(t, ck2, 15,2, "qq")
      
    // restore network, wait for all threads to exit.
    part(t, tag, nservers, []int{0,1,2,3,4}, []int{}, []int{})
    done = true
    ok := true
    for i := 0; i < nclients; i++ {
      z := <- ca[i]
      ok = ok && z
    }
    if ok == false {
      t.Fatal("something is wrong")
    }
    check(t, ck2, 15, 2,"qq")
  }

  fmt.Printf("  ... Passed\n")
}


func TestManyPartition(t *testing.T) {
time.Sleep(2 * time.Second)
  runtime.GOMAXPROCS(4)

  fmt.Printf("Test: Many clients, changing partitions ...\n")

  tag := "many"
  const nservers = 5
  var kva []*KVPaxos = make([]*KVPaxos, nservers)
  defer cleanup(kva)
  defer cleanpp(tag, nservers)
  
  for i := 0; i < nservers; i++ {
    var kvh []string = make([]string, nservers)
    for j := 0; j < nservers; j++ {
      if j == i {
        kvh[j] = port(tag, i)
      } else {
        kvh[j] = pp(tag, i, j)
      }
    }
    kva[i] = StartServer(kvh, i)
    kva[i].unreliable = true
  }
  defer part(t, tag, nservers, []int{}, []int{}, []int{})
  part(t, tag, nservers, []int{0,1,2,3,4}, []int{}, []int{})

  done := false

  // re-partition periodically
  ch1 := make(chan bool)
  go func() {
    defer func() { ch1 <- true } ()
    for done == false {
      var a [nservers]int
      for i := 0; i < nservers; i++ {
        a[i] = (rand.Int() % 3)
      }
      pa := make([][]int, 3)
      for i := 0; i < 3; i++ {
        pa[i] = make([]int, 0)
        for j := 0; j < nservers; j++ {
          if a[j] == i {
            pa[i] = append(pa[i], j)
          }
        }
      }
      part(t, tag, nservers, pa[0], pa[1], pa[2])
      time.Sleep(time.Duration(rand.Int63() % 200) * time.Millisecond)
    }
  }()

  const nclients = 10
  var ca [nclients]chan bool
  for xcli := 0; xcli < nclients; xcli++ {
    ca[xcli] = make(chan bool)
    go func(cli int) {
      ok := false
      defer func() { ca[cli] <- ok }()
      sa := make([]string, nservers)
      for i := 0; i < nservers; i++ {
        sa[i] = port(tag, i)
      }
      for i := range sa {
        j := rand.Intn(i+1)
        sa[i], sa[j] = sa[j], sa[i]
      }
      myck := MakeClerk(sa)
      key := cli
      last := ""
      myck.Put(mencius.Stroke{key, key, 10,10,last,1})
     
      for done == false {
        if (rand.Int() % 1000) < 500 {
          nv := strconv.Itoa(rand.Int())
          myck.Put(mencius.Stroke{key, key, 10,10,nv,1})
          v:=myck.Get(key,key)
          if v != nv {
            t.Fatalf("%v: puthash wrong value, key %v, wanted %v, got %v",
              cli, key, last, v)
          }
          last = v
        } else {
          v := myck.Get(key,key)
          if v != last {
            t.Fatalf("%v: get wrong value, key %v, wanted %v, got %v",
              cli, key, last, v)
          }
        }
      }
     // fmt.Println(">>>>>>>>>>>>>>>>>>>>")
      ok = true
    } (xcli)
  }


  time.Sleep(20 * time.Second)
  done = true
 
  <- ch1
  part(t, tag, nservers, []int{0,1,2,3,4}, []int{}, []int{})
 //fmt.Println("-------------------------------------")


  
  ok := true
  for i := 0; i < nclients; i++ {
  //i=0
   //fmt.Println("????????????????")
    z := <- ca[i]
  //   fmt.Println("??????????????????")
    ok = ok && z    
  }

  if ok {
    fmt.Printf("  ... Passed\n")
  }
}
*/






//Test:a single server crashed/get killed
func TestSingleCrash(t *testing.T) {
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

  fmt.Printf("Test: One server crash ...\n")
  l1 := make([]mencius.Stroke,Interval)
  //kill the server
  kva[0].kill()
  
  //do Interval Puts
  for i :=0;i<Interval+1; i++{
     s1 :=mencius.Stroke{i,i,20, 20, "firstround",1}
     ck.Put(s1) 
     //l1[i]=s1
     time.Sleep(10*time.Millisecond)   
     if i<Interval {l1[i]=s1}
  }

  time.Sleep(2*time.Second)
  checkHistory(t, cka[1], l1)
  checkHistory(t, cka[2], l1) 
 
 //restart the server
  kvh[0] = port("basic", 0)  
  kva[0] = StartServer(kvh, 0)
  time.Sleep(2*time.Second)
  
  checkHistory(t, ck, l1)
  checkHistory(t, cka[1], l1)
  checkHistory(t, cka[2], l1) 
  
  checkClientHisEqual(t, cka[1], cka[2])  
  checkClientHisEqual(t, ck, cka[1])
  checkClientHisEqual(t, ck, cka[2]) 
  
  fmt.Printf("  ... Passed ... \n")
  time.Sleep(1 * time.Second)
}


