package projectserver

type GetUpdateArgs struct{
    Client_newest_op_num int
    ClientID int64
    RequestID int64
}

type GetUpdateReply struct{
    Has_map bool 
    Has_operation bool 
    Board [1000*1000]string //key: pixel index, value: color value
    New_operations []Operation
}

type PutArgs struct{
    Client_newest_op_num int
    Key int
    Value string
    ClientID int64
    RequestID int64
}
type PutReply struct{ //same as GetUpdateReply
    Has_map bool 
    Has_operation bool
    Board [1000*1000]string
    New_operations []Operation
	Err string
}

type GetArgs struct{
    Key int
    ClientID int64
    RequestID int64
}
type GetReply struct{ //same as GetUpdateReply
    Err string
    Value string
}


type Operation struct{
  OpName string
  Key int
  Value string
  OperationId int64
  ClientId int64
  SeqNum int
  Dep []int
  Status string
  Index, Lowlink int
}

type Node struct {
op *Operation
seq int
}
