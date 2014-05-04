package projectserver


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
type PutReply struct{ //same as GetUpdateReply
    Has_operation bool
    New_operations []Operation
	Err string
}

type GetArgs struct{
    SeqNum int
    ClientID int64
    RequestID int64
}
type GetReply struct{ //same as GetUpdateReply
    Err string
    ClientStroke Stroke
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
}

type Node struct {
op *Operation
seq int
}
