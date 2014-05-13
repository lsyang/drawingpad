package projectserver

import "testing"
import "fmt"
import "math/rand"

const Benchservers=3

  func BenchmarkPut(b *testing.B)  { benchmark(Benchservers, b) }
  func BenchmarkConcurrentClientPut(b *testing.B)  { benchmarkConcurrentClients(Benchservers, b) }


  func BenchmarkGetUpdate3Server500Puts(b *testing.B)  { benchmarkgetupdate(Benchservers, 500, b) }
 // func BenchmarkServerDie3Server500Puts(b *testing.B)  { benchmarkServerDie(Benchservers, 500, b) }



var result string
func benchmark(nservers int, b *testing.B) {
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
  var cka []*Clerk=make([]*Clerk, nservers)
  for i := 0; i < nservers; i++ {
    cka[i] = MakeClerk([]string{kvh[i]})
  }

  op:=Stroke{1,10,20,30,"aa",10}
  var r string

  for n := 0; n < b.N; n++ {
    r=ck.Put(op)
  }
  result=r
}

func benchmarkConcurrentClients(nservers int, b *testing.B) {
  var kva []*KVPaxos = make([]*KVPaxos, nservers)
  var kvh []string = make([]string, nservers)
  defer cleanup(kva)

  for i := 0; i < nservers; i++ {
    kvh[i] = port("basic", i)
  }
  for i := 0; i < nservers; i++ {
    kva[i] = StartServer(kvh, i)
  }

  op:=Stroke{1,10,20,30,"aa",10}
  var r string
  

  for n := 0; n < b.N; n++ {
    for iters := 0; iters < 20; iters++ {
      const npara = 15
      var ca [npara]chan bool
      for nth := 0; nth < npara; nth++ {
        ca[nth] = make(chan bool)
        go func(me int) {
          defer func() { ca[me] <- true }()
          ci := (rand.Int() % nservers)
          myck := MakeClerk([]string{kvh[ci]})
          r=myck.Put(op)
        }(nth)
      }
      for nth := 0; nth < npara; nth++ {
        <- ca[nth]
      }
    }
  }
  result=r
}

func benchmarkgetupdate(nservers int, nput int, b *testing.B) {
  var kva []*KVPaxos = make([]*KVPaxos, nservers)
  var kvh []string = make([]string, nservers)
  defer cleanup(kva)

  for i := 0; i < nservers; i++ {
    kvh[i] = port("basic", i)
  }
  for i := 0; i < nservers; i++ {
    kva[i] = StartServer(kvh, i)
  }

  var cka []*Clerk=make([]*Clerk, nservers)
  for i := 0; i < nservers; i++ {
    cka[i] = MakeClerk([]string{kvh[i]})
  }

  op:=Stroke{1,10,20,30,"aa",10}
  var r string

  for n := 0; n < b.N; n++ {
    for i:=0; i<nput; i++{
      r=cka[0].Put(op)
    }
    updates:=cka[1].GetUpdate()
    fmt.Println("length of new op is ," ,len(updates.New_operations))
  }
  result=r
}


//var update GetUpdateReply
func benchmarkServerDie(nservers int, nput int, b *testing.B) {
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
  var cka []*Clerk=make([]*Clerk, nservers)
  for i := 0; i < nservers; i++ {
    cka[i] = MakeClerk([]string{kvh[i]})
  }

  kva[1].kill()

  op:=Stroke{1,10,20,30,"aa",10}
  var r string

  for n := 0; n < b.N; n++ {
    for i:=0; i<nput; i++{
      r=cka[0].Put(op)
    }
    kvh[1] = port("basic", 1)
    kva[1] = StartServer(kvh, 1)
    
    updates:=cka[1].GetUpdate()
    fmt.Println("length of new op is ," ,len(updates.New_operations))
  }
  result=r
}



