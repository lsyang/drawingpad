//import epaxos
package main
// import "projectserver"
// import "strconv"
// import "runtime"
// import "fmt"
// import "os"


import (
        "fmt"
        "log"
        "net/http"
        "encoding/json"
        "projectserver"
        "mencius"
        "strconv"
        "runtime"
        "os"
)

var ClientList []*projectserver.Clerk
const nservers = 3
var kva []*projectserver.KVPaxos = make([]*projectserver.KVPaxos, nservers)
var kvh []string = make([]string, nservers)

func main() {


  runtime.GOMAXPROCS(4)

  //defer cleanup(kva)

  for i := 0; i < nservers; i++ {
    kvh[i] = port("basic", i)
  }
  for i := 0; i < nservers; i++ {
    kva[i] = projectserver.StartServer(kvh, i)
  }
  fmt.Println("making servers")
  StartBrowser()

  //ck := projectserver.MakeClerk(kvh)
  //const nclients=3
  //var cka [nclients]*projectserver.Clerk
   // go  projectserver.MakeClerk(kvh,":9999")
   // go projectserver.MakeClerk(kvh,":9998")
   // projectserver.MakeClerk(kvh)
  // for i := 0; i < nclients; i++ {
  //   cka[i] = projectserver.MakeClerk(kvh)
  // }
  //fmt.Println("making clients")
	
}

func port(tag string, host int) string {
  s := "/var/tmp/824-"
  s += strconv.Itoa(os.Getuid()) + "/"
  os.Mkdir(s, 0777)
  s += "kv-"
  s += strconv.Itoa(os.Getpid()) + "-"
  s += tag + "-"
  s += strconv.Itoa(host)
  return s
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
        // canvas.width = document.body.clientWidth-100; //document.width is obsolete
        // canvas.height = document.body.clientHeight-200; //document.height is obsolete

        $.each(['#f00', '#ff0', '#0f0', '#0ff', '#00f', '#f0f', '#000', '#fff'], function() {
          $('#colors_demo .tools').append("<a href='#colors_sketch' data-color='" + this + "' style='width: 10px; background: " + this + ";'></a> ");
        });
        $.each([3, 5, 10, 15], function() {
          $('#colors_demo .tools').append("<a href='#colors_sketch' data-size='" + this + "' style='background: #ccc'>" + this + "</a> ");
        });
        $('#colors_sketch').sketch();
      });
    </script>

  </head>

  <body>

  <h2>Collaborative Drawing Pad </h2>
             <div id="output"></div>
  <div id="colors_demo">
    <div class="tools">
     
    </div>
  </div>
  <canvas id="colors_sketch" width="600" height="600"></canvas>
  
  </body>

</html>`


// handler for the main page.
func handler(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, page)
}

func handlerStroke(w http.ResponseWriter, r *http.Request){
  r.ParseForm()
  id,_:=strconv.Atoi(r.FormValue("id"))
  x1,_:=strconv.Atoi(r.FormValue("startx"))
  y1,_:=strconv.Atoi(r.FormValue("starty"))
  x2,_:=strconv.Atoi(r.FormValue("endx"))
  y2,_:=strconv.Atoi(r.FormValue("endy"))
  col:=r.FormValue("color")
  size,_:=strconv.Atoi(r.FormValue("size"))
  op:=mencius.Stroke{x1,y1,x2,y2,col,size}
  //fmt.Println("%v,%v,%v,%v", x1,y1,x2,y2)
  ClientList[id].Put(op) //strokes won't be in order
  //fmt.Println(id)
  //ClientList[id].PutChan(op)
}

func drawUpdate(w http.ResponseWriter, r *http.Request) {
  //lock it?
  //call getUpdate() to get GetUpdateReply
  //loop through to see if operation num is continous
  //update max_operation_num
  id,_:=strconv.Atoi(r.FormValue("id"))
  m:=ClientList[id].GetUpdate()
  //m:=ClientList[id].GetChan()
  b, _ := json.Marshal(m)
  fmt.Fprint(w, string(b))
  
}

func handlerRegister(w http.ResponseWriter, r *http.Request){
  //create a new client, need lock
  id,_:=strconv.Atoi(r.FormValue("id"))
  if id==-1{
    id=len(ClientList)
    ck:=projectserver.MakeClerk([]string{kvh[(id%nservers)]})
    ClientList=append(ClientList, ck)
  }
  fmt.Fprint(w, id)
}

func StartBrowser() {
  fmt.Println("starting browser")  
  http.HandleFunc("/", handler)
  http.HandleFunc("/stroke", handlerStroke)
  http.HandleFunc("/drawUpdate", drawUpdate)
  http.HandleFunc("/register", handlerRegister)
  http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources")))) 
  log.Fatal(http.ListenAndServe(":9999", nil))

}
