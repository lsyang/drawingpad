package projectserver

import "testing"
import "runtime"
import "time"
import "fmt"
import "os"
import "strconv"

//Test read and write maxExecutedOpNum variable
func TestMaxExecutedOpNum(t *testing.T) {
    fmt.Printf("TestMaxExecutedOpNum... \n")
             
    os.Remove(MaxExecuted+"1")
    os.Remove(MaxExecuted+"2")
    runtime.GOMAXPROCS(4)
    val,bool:=ReadMaxExecutedOpNum(1)
    if (val!=-1) || (bool){
       fmt.Println(val,bool)
       t.Fatalf("readMaxExecutedOpNum initialization failed")
    }
    
    for i := 0; i < 15; i++ {
	WriteToDisk(1,i)
    	newVal,newBool:=ReadMaxExecutedOpNum(1)
    	if (newVal!=i) || (!newBool){
       		t.Fatalf("readMaxExecutedOpNum for ",i," failed")
    	}   
    }
         
     //try write without initilization
     WriteToDisk(2,23)
     Val3,Bool3:=ReadMaxExecutedOpNum(2)
     if (Val3!=23) || (!Bool3){
       t.Fatalf("readMaxExecutedOpNum for 23 failed")
    }
         
    os.Remove(MaxExecuted+"1")
    os.Remove(MaxExecuted+"2")
    fmt.Printf("  ... Passed\n")
 
  time.Sleep(2 * time.Millisecond)
}


func TestCachedRequestState(t *testing.T) {
   fmt.Printf("TestCachedRequestState... \n")
   os.Remove(CachedRequest+"1")
   os.Remove(CachedRequest+"2")
   runtime.GOMAXPROCS(4)

   val,bool:=ReadCachedRequestState(1)
   if (len(val)!=0) || (bool){
       t.Fatalf("readCachedRequestState initialization failed")
   }
    
   for i := 0; i < 30; i++ {
        val[int64(i)]=CachedRequestState{int64(i),strconv.Itoa(i)}
	WriteToDisk(1,val)
    	newVal,newBool:=ReadCachedRequestState(1)
    	for j:=0;j<=i;j++{
    	   request:=newVal[int64(j)]
    	   if (request.OperationId !=int64(j) )|| (request.Result!=strconv.Itoa(j))|| (!newBool){
       		t.Fatalf("readMaxCachedRequestState for ",j," failed")
    	  }  
    	} 
    }


     WriteToDisk(2,val)
     Val3,Bool3:=ReadCachedRequestState(2)
     if (!Bool3)||(len(Val3)!=30){
       t.Fatalf("readCachedRequestState for 2 failed")
     }
   	
     for i := 0; i <30; i++ {
        request:=Val3[int64(i)]
    	if (request.OperationId !=int64(i) )|| (request.Result!=strconv.Itoa(i)){
       		t.Fatalf("readMaxCachedRequestState for ",i," failed")
	}  	  
    }
       
    os.Remove(CachedRequest+"1")
    os.Remove(CachedRequest+"2")
     
    fmt.Printf("  ... Passed\n") 
    time.Sleep(1 * time.Millisecond)
}

func TestOpLogs(t *testing.T) {
    fmt.Printf("TestOpLogs... \n")
    os.Remove(OperationLogs+"1")
    os.Remove(OperationLogs+"2")
    runtime.GOMAXPROCS(4)

    val,bool:=ReadOpLogs(1)
    if (len(val)!=0) || (bool){
       t.Fatalf("readOpLogs initialization failed")
    }
    
    for i := 0; i < 40; i++ {
        stroke:=Stroke{i,i+1,i-1,i,strconv.Itoa(i),i}
        val[i]=Operation{SeqNum:i,ClientStroke:stroke}
	WriteToDisk(1,val)
    	newVal,newBool:=ReadOpLogs(1)
    	for j:=0;j<=i;j++{
    	   op:=newVal[j]
    	   newStroke:=op.ClientStroke
    	    if (op.SeqNum !=j )|| (newStroke.Start_x!=j)||(newStroke.Start_y!=j+1)|| (newStroke.Size!=j)|| (!newBool){
       		t.Fatalf("readOpLogs for ",j," failed")
    	  }  
    	} 
    }

     //try write without initilization
     WriteToDisk(2,val)
     Val3,Bool3:=ReadOpLogs(2)
     if (!Bool3)||(len(Val3)!=40){
       t.Fatalf("readOpLogs for 2 failed")
     }
   	
     for i := 0; i < 40; i++ {
           op:=Val3[i]
    	   newStroke:=op.ClientStroke
    	  
    	   if (op.SeqNum !=i)|| (newStroke.Start_x!=i)||((newStroke.Start_y!=i+1))|| ((newStroke.Size!=i)){
       		t.Fatalf("readOpLogs for ",i," failed")
    	  }    
    }
       
    os.Remove(OperationLogs+"1")
    os.Remove(OperationLogs+"2")
     
    fmt.Printf("  ... Passed\n") 
    time.Sleep(1 * time.Millisecond)
}

