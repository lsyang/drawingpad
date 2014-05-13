package projectserver
import "sync"
const (
  MaxExecuted="maxExecutedOpNum"
  CachedRequest="cachedRequestState"
  OperationLogs="opLogs"
  Store = true
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
    New_operations []Operation
}

type PutArgs struct{
    Client_newest_op_num int
    ClientStroke Stroke
    ClientID int64
    RequestID int64
}
type PutReply struct{
    Has_operation bool
    New_operations []Operation
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


type Operation struct{
  OpName string
  ClientStroke Stroke
  OperationId int64
  ClientId int64
  SeqNum int
  Dep []int
  Status string
  Index, Lowlink int
}

type Stroke struct{
  Start_x int
  Start_y int
  End_x int
  End_y int
  Color string
  Size int
}

type Node struct {
op *Operation
seq int
}
