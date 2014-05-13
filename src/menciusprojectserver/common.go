package menciusprojectserver
import "sync"
import "mencius"

const (
  MaxExecuted="maxExecutedOpNum"
  CachedRequest="cachedRequestState"
  OperationLogs="opLogs"
  Interval=50
  
)

type Lock struct{
   mu sync.Mutex
}

type GetUpdateArgs struct{
    Client_newest_op_num int
    ClientID int64
    RequestID int64
}

type GetUpdateReply struct{
    Has_operation bool 
    New_operations []mencius.Operation
}

type PutArgs struct{
    Client_newest_op_num int
    ClientStroke mencius.Stroke
    ClientID int64
    RequestID int64
}
type PutReply struct{
    Has_operation bool
    New_operations []mencius.Operation
	Err string
}

type GetArgs struct{
    Start_x int
	  Start_y int
    ClientID int64
    RequestID int64
}
type GetReply struct{ 
    Err string
    Value string
}



type Node struct {
op *mencius.Operation
seq int
}
