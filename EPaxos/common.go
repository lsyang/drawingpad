package EPaxos

type GetUpdateArgs struct{
	Client_newest_op_num int
	ClientID int64
	RequestID int64
}
type GetUpdateReply struct{
	Has_update bool 
	Board map[int]int //key: pixel index, value: color value
	New_operations []Operation
}

type PutArgs struct{
	New_operation Operation
	ClientID int64
	RequestID int64
}
type PutReply struct{ //same as GetUpdateReply
	Has_update bool 
	Board map[int]int 
	New_operations []Operation
}
type Operation struct{
	Key int //pixel index
	Value int //color value
	Operation_num int
}