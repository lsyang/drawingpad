package projectserver

/*save a snapshot of key value storage after we execute snapshot_interval operations.
func (kv *KVPaxos) SaveSnapshot(op Operation, seq int) {
  if (seq%kv.snapshot_interval==0){
     kv.board_snapshot=kv.board
  }
}


//clean memory
func (kv *KVPaxos) cleanMemory(ins_num int) {
  for key, _ := range kv.opLogs {
    if key <= ins_num{
      delete(kv.opLogs, key)
    }
  }
}
*/