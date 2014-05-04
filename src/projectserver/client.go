package projectserver

import "net/rpc"
import "fmt"
import "time"
import "crypto/rand"
import "math/big"
import "sync"
import "log"
import "net/http"
import "encoding/json"
import "strconv"


type Clerk struct {
  mu sync.Mutex // one RPC at a time
  servers []string
  // You will have to modify this struct.
  me int64
  requestID int64
  max_operation_num int
  keys []int
}


func MakeClerk(servers []string) *Clerk {
  ck := new(Clerk)
  ck.servers = servers
  // You'll have to add code here.
  ck.requestID=1
  ck.me=nrand()
  ck.StartBrowser()
  return ck
}

func nrand() int64 {
  max := big.NewInt(int64(1) << 62)
  bigx, _ := rand.Int(rand.Reader, max)
  x := bigx.Int64()
  return x
}
//
// call() sends an RPC to the rpcname handler on server srv
// with arguments args, waits for the reply, and leaves the
// reply in reply. the reply argument should be a pointer
// to a reply structure.
//
// the return value is true if the server responded, and false
// if call() was not able to contact the server. in particular,
// the reply's contents are only valid if call() returned true.
//
// you should assume that call() will time out and return an
// error after a while if it doesn't get a reply from the server.
//
// please use call() to send all RPCs, in client.go and server.go.
// please don't change this function.
//
func call(srv string, rpcname string,
          args interface{}, reply interface{}) bool {
  c, errx := rpc.Dial("unix", srv)
  if errx != nil {
    return false
  }
  defer c.Close()
    
  err := c.Call(rpcname, args, reply)
  if err == nil {
    return true
  }

  fmt.Println(err)
  return false
}

//
// Get update from the server
//
//func (ck *Clerk) GetUpdate() (bool bool map[int]int []Operation) {
func (ck *Clerk) GetUpdate() GetUpdateReply{
  // ck.mu.Lock()
  // defer ck.mu.Unlock()
  var reply GetUpdateReply
  args := &GetUpdateArgs{ck.max_operation_num,ck.me,ck.requestID}
  for{
    for _, srv := range ck.servers {
      ok := call(srv, "KVPaxos.GetUpdate", args, &reply)
      if ok {
        ck.requestID++
        return reply
      }
    }
    time.Sleep(100 * time.Millisecond)
  }
  return reply
}

func (ck *Clerk) Get(key int) string {
  // ck.mu.Lock()
  // defer ck.mu.Unlock()
  //increment the operationId to be the next one
  args := &GetArgs{key, ck.me, ck.requestID}
  for {
    //try sending request for all the servers
    for _, server := range ck.servers {
      reply := &GetReply{}
      ok := call(server, "KVPaxos.Get", args, reply)
      if ok == true && reply.Err == "" {
        ck.requestID++
        return reply.Value 
      }
    }
    time.Sleep(time.Second)
  }
}

//
// Put operation by client
//
func (ck *Clerk) Put(key int, value string) PutReply {
  ck.mu.Lock()
  defer ck.mu.Unlock()
  // var new_op Operation
  // new_op.OpName="Put"
  // new_op.Key=key
  // new_op.Value=value
  // new_op.OperationId=ck.requestID
  // new_op.ClientId=ck.me

  var reply PutReply
  for{
    for _, srv := range ck.servers {
      args := &PutArgs{ck.max_operation_num,key,value,ck.me,ck.requestID}
      
      ok := call(srv, "KVPaxos.Put", args, &reply)
      if ok {
        ck.requestID++
        //TODO: check if current op is in reply
        return reply
      }
    }
    time.Sleep(100 * time.Millisecond)
  }
  return reply
}

// Content for the main html page..
var page =
`<html>
  <head>
      <meta content="text/html; charset=utf-8" http-equiv="Content-Type" />
        <title>Canvas Drawing</title>
        <link rel="stylesheet" type="text/css" href="resources/test.css" />
        <script type="text/javascript" src="resources/external_js/jquery-1.9.0.min.js"></script>
    
    <script type="text/javascript" src="resources/sketch.js"></script>


    <script type="text/javascript">
      $(function() {
        canvas = document.getElementById("colors_sketch");
        // canvas.width = document.body.clientWidth; //document.width is obsolete
        // canvas.height = document.body.clientHeight; //document.height is obsolete

        $.each(['#f00', '#ff0', '#0f0', '#0ff', '#00f', '#f0f', '#000', '#fff'], function() {
          $('#colors_demo .tools').append("<a href='#colors_sketch' data-color='" + this + "' style='width: 10px; background: " + this + ";'></a> ");
        });
        // $.each([3, 5, 10, 15], function() {
        //   $('#colors_demo .tools').append("<a href='#colors_sketch' data-size='" + this + "' style='background: #ccc'>" + this + "</a> ");
        // });
        $('#colors_sketch').sketch();
      });
    </script>

  </head>

  <body>

  <h2>Go Timer (ticks every second!), the system clock</h2>
             <div id="output"></div>
  <div id="colors_demo">
    <div class="tools">
      <a href="#colors_sketch" data-download="png" style="float: right; width: 100px;">Download</a>
    </div>
  </div>
  <canvas id="colors_sketch" width="600" height="600"></canvas>
  
  </body>

</html>`


// handler for the main page.
func (ck *Clerk) handler(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, page)
}

// handler to cater AJAX requests
func (ck *Clerk) handlerGetTime(w http.ResponseWriter, r *http.Request) {
        body := r.FormValue("inputVal")
        fmt.Println(body)
        fmt.Fprint(w, time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"))
}

func (ck *Clerk) handlerStroke(w http.ResponseWriter, r *http.Request){
  r.ParseForm()
  keys:=r.FormValue("Key")
  //keys_json:="{"+`"Keys":`+ keys +`}`
  val:=r.FormValue("Value")
  // var line Line
  // var lala []byte
 // lala= []byte(keys_json)
  //json.Unmarshal(lala, &line)
  key_int,_:=strconv.Atoi(keys)
  ck.keys=append(ck.keys,key_int)
  m:=ck.Put(key_int,val)
  
  // op:=EPaxos.Operation{key_int, val,0}
  // m.New_operations=append(m.New_operations, op)
  // for i := 0; i < len(line.Keys); i++ {
  //     op:=EPaxos.Operation{line.Keys[i], val,0}
  //     m.New_operations=append(m.New_operations, op)
  // }
  b, _ := json.Marshal(m)
//put response into a json file
  fmt.Fprint(w, string(b))
}

func (ck *Clerk) drawUpdate(w http.ResponseWriter, r *http.Request) {
  //lock it?
  //call getUpdate() to get GetUpdateReply
  //loop through to see if operation num is continous
  //update max_operation_num
  m:=ck.GetUpdate()
  operations:=m.New_operations
  fmt.Println(len(operations))
  if m.Has_operation{
   ck.max_operation_num=operations[len(operations)-1].SeqNum
   fmt.Println("max seq Num,  %v",ck.max_operation_num)
  }
  // m.Has_operation=true
  // m.New_operations=append(m.New_operations, op1, op2)
  b, _ := json.Marshal(m)
  //put response into a json file
  fmt.Fprint(w, string(b))
}

func (ck *Clerk) StartBrowser() {
  fmt.Println("starting browser")
//  http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources")))) 
    http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources")))) 
  http.HandleFunc("/time", ck.handler)
  http.HandleFunc("/stroke", ck.handlerStroke)
  http.HandleFunc("/drawUpdate", ck.drawUpdate)
  http.HandleFunc("/gettime", ck.handlerGetTime)
  log.Fatal(http.ListenAndServe(":9999", nil))
}

