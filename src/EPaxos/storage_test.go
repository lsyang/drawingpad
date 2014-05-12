package epaxos

import "testing"
import "runtime"
import "time"
import "fmt"
import "os"

func TestWriteStatusMap(t *testing.T) {
  fmt.Printf("TestWriteStatusMap. \n")
  	os.Remove(StatusMap+"1")
	os.Remove(StatusMap+"2")    
  runtime.GOMAXPROCS(4)
      val,bool:=ReadStatusMap(1)
    if (len(val)!=0) || (bool){
       t.Fatalf("readStatusMap initialization failed")
    }    
    for i := 0; i < 40; i++ {
        newStatus:=Status{Done:false,SeqNum:i, Key:i}
        val[i]=newStatus
		WriteToDisk(1,val)
    	newVal,newBool:=ReadStatusMap(1)
    	for j:=0;j<=i;j++{
    	   status:=newVal[j]
    	    if (status.Key !=j )|| (status.SeqNum!=j)||(status.Done)|| (!newBool){
    	    fmt.Println(status,j)
       		//t.Fatalf("readStatus for ",j," failed")
    	  }  
    	} 
    }

     WriteToDisk(2,val)
     Val3,Bool3:=ReadStatusMap(2)
     if (!Bool3)||(len(Val3)!=40){
       t.Fatalf("readStatus for 2 failed")
   	}   	
     for i := 0; i < 40; i++ {
           status:=Val3[i]
    	    if (status.Key !=i )|| (status.SeqNum!=i)||(status.Done){
       		t.Fatalf("readStatus for ",i," failed")
    	  }    
    }
       
	os.Remove(StatusMap+"1")
	os.Remove(StatusMap+"2")     
    fmt.Printf("  ... Passed\n") 
	time.Sleep(1 * time.Millisecond)	
}


func TestWriteAcceptorStateMap(t *testing.T) {
  fmt.Printf("TestAcceptorStateMap. \n")
  	os.Remove(AcceptorStateMap+"1")
	os.Remove(AcceptorStateMap+"2")  
  runtime.GOMAXPROCS(4)
      val,bool:=ReadAcceptorStateMap(1)
    if (len(val)!=0) || (bool){
       t.Fatalf("readAcceptorStateMap initialization failed")
    }    
    for i := 0; i < 40; i++ {
        newStatus:=AcceptorState{i,i,i}
        val[i]=newStatus
		WriteToDisk(1,val)
    	newVal,newBool:=ReadAcceptorStateMap(1)
    	for j:=0;j<=i;j++{
    	   status:=newVal[j]
    	    if (status.N_a !=j )|| (status.V_a!=j)||(status.N_p!=j)|| (!newBool){
    	    fmt.Println(status,j)
       		//t.Fatalf("readAcceptorStateMap for ",j," failed")
    	  }  
    	} 
    }
     WriteToDisk(2,val)
     Val3,Bool3:=ReadAcceptorStateMap(2)
     if (!Bool3)||(len(Val3)!=40){
       t.Fatalf("readAcceptorStateMap for 2 failed")
   	}   	
     for i := 0; i < 40; i++ {
           status:=Val3[i]
    	    if (status.N_a !=i )|| (status.N_p!=i)|| (status.V_a!=i){
       		t.Fatalf("readAcceptorStateMap for ",i," failed")
    	  }    
    }
       
	os.Remove(AcceptorStateMap+"1")
	os.Remove(AcceptorStateMap+"2")     
    fmt.Printf("  ... Passed\n") 
	time.Sleep(1 * time.Millisecond)
}

func TestPeersDone(t *testing.T) {
  fmt.Printf("TestPeersDone. \n")
  	os.Remove(PeersDoneValue+"1")
	os.Remove(PeersDoneValue+"2")  
  runtime.GOMAXPROCS(4)
      val,bool:=ReadPeersDoneValue(1,10)
    if (len(val)!=10) || (bool){
       t.Fatalf("readPeersDone initialization failed")
    }    
    
    for i:=0;i<10;i++{
     if (val[i]!=-1) {
       t.Fatalf("readPeersDoneValue initialization failed")
    } 
    }
    
    for i := 0; i < 10; i++ {
        val[i]=i
		WriteToDisk(1,val)
    	newVal,newBool:=ReadPeersDoneValue(1,10)
    	for j:=0;j<=i;j++{
    	
    	    if (newVal[j] !=j )|| (!newBool){
       		t.Fatalf("readPeersDoneValue for ",j," failed")
    	  }  
    	} 
    }
     WriteToDisk(2,val)
     Val3,Bool3:=ReadPeersDoneValue(2,10)
     if (!Bool3)||(len(Val3)!=10){
       t.Fatalf("readPeersDoneValue for 2 failed")
   	}   	
     for i := 0; i < 10; i++ {
    	    if (Val3[i]!=i ){
       		t.Fatalf("readPeersDoneValue for ",i," failed")
    	  }    
    }
       
	os.Remove(PeersDoneValue+"1")
	os.Remove(PeersDoneValue+"2")     
    fmt.Printf("  ... Passed\n") 
	time.Sleep(1 * time.Millisecond)
}



func TestMax(t *testing.T) {
    fmt.Printf("TestMax... \n")
        os.Remove(Max+"1")
     os.Remove(Max+"2")
  runtime.GOMAXPROCS(4)
    //try reading opNum first
    val,bool:=ReadMax(1)
    if (val!=-1) || (bool){
       fmt.Println(val,bool)
       t.Fatalf("readMax initialization failed")
    }
    
    for i := 0; i < 15; i++ {
		WriteMax(1,i)
    	newVal,newBool:=ReadMax(1)
    	if (newVal!=i) || (!newBool){
       		t.Fatalf("readMax for ",i," failed")
    	}   
    }        
     //try write without initilization
     WriteMax(2,23)
     Val3,Bool3:=ReadMax(2)
    	if (Val3!=23) || (!Bool3){
       t.Fatalf("readMax for 23 failed")
   	}
         
    os.Remove(Max+"1")
     os.Remove(Max+"2")
    fmt.Printf("  ... Passed\n")
 
  time.Sleep(1 * time.Millisecond)
}

func TestMin(t *testing.T) {
    fmt.Printf("TestMin... \n")
        os.Remove(Min+"1")
     os.Remove(Min+"2")
  runtime.GOMAXPROCS(4)
    //try reading opNum first
    val,bool:=ReadMin(1)
    if (val!=0) || (bool){
       fmt.Println(val,bool)
       t.Fatalf("readMin initialization failed")
    }
    
    for i := 0; i < 15; i++ {
		WriteMin(1,i)
    	newVal,newBool:=ReadMin(1)
    	if (newVal!=i) || (!newBool){
       		t.Fatalf("readMin for ",i," failed")
    	}   
    }        
     //try write without initilization
     WriteMin(2,23)
     Val3,Bool3:=ReadMin(2)
    	if (Val3!=23) || (!Bool3){
       t.Fatalf("readMin for 23 failed")
   	}
         
    os.Remove(Min+"1")
     os.Remove(Min+"2")
    fmt.Printf("  ... Passed\n")
 
  time.Sleep(1 * time.Millisecond)
}

