package projectserver

import "testing"
import "runtime"
import "time"
import "fmt"

//Check the history of clients to see if it equals the given list of strokes
//Skip noop operations
func checkHistory(t *testing.T, ck *Clerk, correct_strokes []Stroke) {
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
      if j!=len(correct_strokes){
         t.Fatalf("Getupdate()'s return %v noop stroke, should be %v strokes", j, len(correct_strokes))
      }
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
func strokeEqual(s1 Stroke, s2 Stroke) bool {
  result := (s1.Start_x==s2.Start_x && s1.Start_y==s2.Start_y && s1.End_x==s2.End_x && s1.End_y==s2.End_y && s1.Color==s2.Color && s1.Size==s2.Size)
  return result
}

func isNoOp(s1 Stroke) bool {
  result := (s1.Start_x==0 && s1.Start_y==0 && s1.End_x==0 && s1.End_y==0 && s1.Color=="" && s1.Size==0)
  return result
}

//Test:a single server crashed/get killed
func TestSingleCrash(t *testing.T) {
  const nservers = 3

  deleteStorage(nservers)
  deletePaxosStorage(nservers)

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
  l1 := make([]Stroke,Interval)
  //kill the server
  kva[0].kill()
  
  //do Interval Puts
  for i :=0;i<Interval+1; i++{
     s1 :=Stroke{i,i,20, 20, "firstround",1}
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


func TestSingleCrashUnreliable (t *testing.T){
 const nservers = 3
  
  deleteStorage(nservers)
  deletePaxosStorage(nservers)
 
  var kva []*KVPaxos = make([]*KVPaxos, nservers)
  var kvh []string = make([]string, nservers)
  defer cleanup(kva)

  for i := 0; i < nservers; i++ {
    kvh[i] = port("basic", i)
  }
  for i := 0; i < nservers; i++ {
    kva[i] = StartServer(kvh, i)
	kva[i].unreliable=true
  }

  ck := MakeClerk(kvh)
  var cka [nservers]*Clerk
  for i := 0; i < nservers; i++ {
    cka[i] = MakeClerk([]string{kvh[i]})
  }
  fmt.Printf(" Two servers unreliable, one server crashes for a period... \n")
  //kill the server
  kva[0].kill()
  
  //do Interval Puts
  for i :=0;i<Interval+1; i++{
     s1 :=Stroke{i,i,20, 20, "unreliable",1}
     ck.Put(s1)  
  }

 //restart the server
  kvh[0] = port("basic", 0)  
  kva[0] = StartServer(kvh, 0)
  time.Sleep(2*time.Second)
  
  checkClientHisEqual(t, cka[1], cka[2])  
  checkClientHisEqual(t, ck, cka[1])
  checkClientHisEqual(t, ck, cka[2]) 

  fmt.Printf("  ... Passed ... \n")
  time.Sleep(1 * time.Second)
}


//Test: all servers crashed/get killed
func TestAllCrash(t *testing.T) {
  runtime.GOMAXPROCS(4)
  const nservers = 3

  deleteStorage(nservers)
  deletePaxosStorage(nservers)

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

  fmt.Printf("Test: All servers crash ...\n")
  l1 := make([]Stroke,Interval)
  
  //do Interval Puts
  for i :=0;i<Interval+1; i++{
     s1 :=Stroke{i,i,20, 20, "firstround",1}
     ck.Put(s1) 
     time.Sleep(10*time.Millisecond)   
     if i<Interval {l1[i]=s1}
  }

  time.Sleep(3*time.Second)
  checkHistory(t, ck, l1)
  checkHistory(t, cka[1], l1)
  checkHistory(t, cka[2], l1) 
  
  //kill all servers
  for i := 0; i < nservers; i++ {
     kva[i].kill()
  }

  time.Sleep(1*time.Second)
  //restart all servers
  for i := 0; i < nservers; i++ {
     kvh[i] = port("basic", i)  //restarting server
     kva[i] = StartServer(kvh, i)
  }
  
  checkClientHisEqual(t, cka[1], cka[2])  
  checkClientHisEqual(t, ck, cka[1])
  checkClientHisEqual(t, ck, cka[2]) 
  fmt.Printf("  ... Passed\n")
}


func TestAllCrashUnreliable (t *testing.T){
 runtime.GOMAXPROCS(4)
  const nservers = 3

  deleteStorage(nservers)
  deletePaxosStorage(nservers)

  var kva []*KVPaxos = make([]*KVPaxos, nservers)
  var kvh []string = make([]string, nservers)
  defer cleanup(kva)

  for i := 0; i < nservers; i++ {
    kvh[i] = port("basic", i)
  }
  for i := 0; i < nservers; i++ {
    kva[i] = StartServer(kvh, i)
    kva[i].unreliable=true
  }	

  ck := MakeClerk(kvh)
  var cka [nservers]*Clerk
  for i := 0; i < nservers; i++ {
    cka[i] = MakeClerk([]string{kvh[i]})
  }

  fmt.Printf("Test: All unreliable servers crash ...\n")
  
  //do Interval Puts
  for i :=0;i<Interval+1; i++{
     s1 :=Stroke{i,i,20, 20, "firstround",1}
     ck.Put(s1) 
     time.Sleep(10*time.Millisecond)   
  }

  //kill all servers
  for i := 0; i < nservers; i++ {
     kva[i].kill()
  }

  time.Sleep(1*time.Second)
  //restart all servers
  for i := 0; i < nservers; i++ {
     kvh[i] = port("basic", i)  //restarting server
     kva[i] = StartServer(kvh, i)
  }
  checkClientHisEqual(t, cka[1], cka[2])  
  checkClientHisEqual(t, ck, cka[1])
  checkClientHisEqual(t, ck, cka[2]) 
  fmt.Printf("  ... Passed\n")
  time.Sleep(1*time.Second)
}


// Test: a single server crash and lose disk content
func TestSingleLoseDisk (t *testing.T){
runtime.GOMAXPROCS(4)

  const nservers = 3

  deleteStorage(nservers)
  deletePaxosStorage(nservers)

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

  fmt.Printf("Test: One server crashes and loses disk content ...\n")
  l1 := make([]Stroke,Interval)
  
  //do Interval Puts
  
  for i :=0;i<Interval+1; i++{   
     s1 :=Stroke{i,i,20, 20, "hp",1}
     ck.Put(s1) 
     time.Sleep(10*time.Millisecond)   
     if i<Interval {l1[i]=s1}
  }

  time.Sleep(3*time.Second)
  checkHistory(t, cka[1], l1)
  checkHistory(t, cka[2], l1) 
  
  //kill server o
  kva[0].kill()
  //delete its storage
  deleteStorage(1)

  time.Sleep(1*time.Second)
  //restart server o
  kvh[0] = port("basic", 0) 
  kva[0] = StartServer(kvh, 0)
  
  ck.Put(Stroke{101,101,20,20,"after delete", 2}) 

  time.Sleep(1*time.Second)
  checkHistory(t, ck, l1)
  checkHistory(t, cka[1], l1)
  checkHistory(t, cka[2], l1) 
  
  fmt.Printf("  ... Passed \n")

}

// Test: a single server crash and lose disk content.
// All servers are unreliable
func TestSingleLoseDiskUnreliable (t *testing.T){
runtime.GOMAXPROCS(4)

  const nservers = 3

  deleteStorage(nservers)
  deletePaxosStorage(nservers)

  var kva []*KVPaxos = make([]*KVPaxos, nservers)
  var kvh []string = make([]string, nservers)
  defer cleanup(kva)

  for i := 0; i < nservers; i++ {
    kvh[i] = port("basic", i)
  }
  for i := 0; i < nservers; i++ {
    kva[i] = StartServer(kvh, i)
    kva[i].unreliable=true
  }

  ck := MakeClerk(kvh)
  var cka [nservers]*Clerk
  for i := 0; i < nservers; i++ {
    cka[i] = MakeClerk([]string{kvh[i]})
  }

  fmt.Printf("Test: One server crashes and loses disk content, unreliable ...\n")
  
  //do Interval Puts 
  for i :=0;i<Interval+1; i++{   
     s1 :=Stroke{i,i,20, 20, "hp",1}
     ck.Put(s1) 
  }
  time.Sleep(1*time.Second)
  //kill server
  kva[0].kill()
  //delete its storage
  deleteStorage(1)

  time.Sleep(1*time.Second)
  //restart server o
  kvh[0] = port("basic", 0) 
  kva[0] = StartServer(kvh, 0)
  
  ck.Put(Stroke{101,101,20,20,"after delete", 2}) 

  time.Sleep(2*time.Second)
  
  checkClientHisEqual(t, cka[1], cka[2])  
  checkClientHisEqual(t, ck, cka[1])
  checkClientHisEqual(t, ck, cka[2]) 
  fmt.Printf("  ... Passed \n")

}
