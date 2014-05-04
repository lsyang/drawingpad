package EPaxos

type GetUpdateArgs struct{
	Client_newest_op_num int
	ClientID int64
	RequestID int64
}

type GetUpdateReply struct{
	Has_map bool 
	Has_operation bool 
	Board []string //key: pixel index, value: color value
	New_operations []Operation
}

type PutArgs struct{
	Client_newest_op_num int
	New_operation Operation
	ClientID int64
	RequestID int64
}
type PutReply struct{ //same as GetUpdateReply
	Has_map bool 
	Has_operation bool
	Board []string
	New_operations []Operation
}
type Operation struct{
	Key int //pixel index
	Value string //color value
	Operation_num int
}