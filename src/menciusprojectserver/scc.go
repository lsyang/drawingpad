package projectserver

//import "sort"
//import "time"
import "mencius"
//import "fmt"

type SCComponent struct {
nodes []mencius.Operation
color int8
}




// Execute one specific op(i.e. all the operations in the dependency chain), return the result of execution

func (kv *KVPaxos) findSCC(ins_num int) string {
	op :=kv.opLogs[ins_num]
	//if already executed, could directly return the value in the log.
    if (op.Status== "EXECUTED"){
	    return op.ClientStroke.Color
	}
	//index := 1
	//find SCCs using Tarjan's algorithm
	//kv.stack = kv.stack[0:0]	
	//_ , result:=kv.strongconnect(&op, &index, ins_num)
	result :=kv.ExecuteOp(ins_num)
	return result
}


/*build depedency graph by adding operation and all operations in its drpendency list			 
find the strongly connected components, sort them topologivally
in inverse topological order, for each strongly connected compponent:
sort all commands in strongly connected components by their sequence number
execute every un execured commands in increasing seqeunce number order, marking them MarkAsExecuted

func (kv *KVPaxos) strongconnect(v *mencius.Operation, index *int,seq int) (bool,string) {
    kv.stackMu.Lock()

	//val:=kv.ExecuteOp(seq)
	//return true, val

	v.Index = *index
	v.Lowlink = *index
	//index is incremented on each strongconnect call
	*index = *index + 1

	l := len(kv.stack)
	if l == cap(kv.stack) {
	newSlice := make([]Node, l, 2*l)
	copy(newSlice, kv.stack)
	kv.stack = newSlice
	}
	kv.stack = kv.stack[0 : l+1]
	kv.stack[l] = Node{v,seq}
        kv.stackMu.Unlock()
	
	inst := v.Dep
	for i := 0; i <len(inst); i++ {
		// wait for the operation to commit.
	 	kv.logsMu.Lock()
		w,ok := kv.opLogs[inst[i]]		
	 	kv.logsMu.Unlock()
	    for (!ok){
	        w,ok = kv.opLogs[inst[i]]
            time.Sleep(1000 * 1000)
        }
              
        //if this is v         	
	    if w.Index == 0 {
            Bool,Result:=kv.strongconnect(&w, index,i)
			if !Bool {
					for j := l; j < len(kv.stack); j++ {
						kv.stack[j].op.Index = 0
					}
					kv.stack = kv.stack[0:l]
					return false,Result
				}
			if w.Lowlink < v.Lowlink {
				v.Lowlink = w.Lowlink
			}
			} else { 
				if w.Index < v.Lowlink {
					v.Lowlink = w.Index
				}
			}
		}
	
	    kv.stackMu.Lock()	    
		Result:=""
		if v.Lowlink == v.Index {
			//found SCC
			list := kv.stack[l:len(kv.stack)]
			//execute commands in the increasing order of the Seq field
			sort.Sort(nodeArray(list))
			for _, w := range list {
			//execute operation with instance number
			Result=kv.ExecuteOp(w.seq)
			}
			kv.stack = kv.stack[0:l]
		}
			kv.stackMu.Unlock()
	return true,Result
}

func (kv *KVPaxos) inStack(w Node) bool {
	for _, u := range kv.stack {
		if w.op == u.op && w.seq==u.seq {
			return true
		}
	}
	return false
}
*/
type nodeArray []Node

func (na nodeArray) Len() int {
return len(na)
}

func (na nodeArray) Less(i, j int) bool {
return na[i].op.SeqNum < na[j].op.SeqNum
}

func (na nodeArray) Swap(i, j int) {
na[i], na[j] = na[j], na[i]
}
