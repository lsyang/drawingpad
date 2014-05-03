/* tick-tock.go */
package main
//package EPaxos

import (
        "fmt"
        "log"
        "net/http"
        "time"
        "encoding/json"
        "EPaxos"
       // "strconv"
)

type Line struct{
  Keys []int
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
        $.each([3, 5, 10, 15], function() {
          $('#colors_demo .tools').append("<a href='#colors_sketch' data-size='" + this + "' style='background: #ccc'>" + this + "</a> ");
        });
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
func handler(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, page)
}

// handler to cater AJAX requests
func handlerGetTime(w http.ResponseWriter, r *http.Request) {
        body := r.FormValue("inputVal")
        fmt.Println(body)
        fmt.Fprint(w, time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"))
}

func handlerStroke(w http.ResponseWriter, r *http.Request){
  r.ParseForm()
  keys:=r.FormValue("Key")
  keys_json:="{"+`"Keys":`+ keys +`}`
  val:=r.FormValue("Value")
  var line Line
  var lala []byte
  lala= []byte(keys_json)
  err:=json.Unmarshal(lala, &line)
  fmt.Println(err)
  fmt.Println(keys)
  fmt.Println(lala)
  fmt.Println(line.Keys)

  var m EPaxos.GetUpdateReply
  m.Has_operation=true
  for i := 0; i < len(line.Keys); i++ {
      op:=EPaxos.Operation{line.Keys[i], val,0}
      m.New_operations=append(m.New_operations, op)
  }
  b, _ := json.Marshal(m)
//put response into a json file
  fmt.Fprint(w, string(b))

  // var line Line 
  // err := json.Unmarshal(r, &line)
  // if err==nil{
  //   //call Put, get PutReply
  //   //dummmy code
  //   var m EPaxos.GetUpdateReply
  //   m.Has_operation=true
  //   for i := 0; i < len(line.Keys); i++ {
  //       op:=EPaxos.Operation{Line.keys[i], line.Val,0}
  //       m.New_operations=append(m.New_operations, op)
  // }
  // b, _ := json.Marshal(m)
  // //put response into a json file
  // fmt.Fprint(w, string(b))

  // }
}

func drawUpdate(w http.ResponseWriter, r *http.Request) {
  //lock it?
  //call getUpdate() to get GetUpdateReply
  //loop through to see if operation num is continous
  //update max_operation_num
  op1:=EPaxos.Operation{180300,"#000000",0}
  op2:=EPaxos.Operation{180400,"#000000",1}
  var m EPaxos.GetUpdateReply
  m.Has_operation=true
  m.New_operations=append(m.New_operations, op1, op2)

  b, _ := json.Marshal(m)
  //put response into a json file
  fmt.Fprint(w, string(b))
}

func main() {
        http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources")))) 
        http.HandleFunc("/time", handler)
        http.HandleFunc("/stroke", handlerStroke)
        http.HandleFunc("/drawUpdate", drawUpdate)
        http.HandleFunc("/gettime", handlerGetTime)
        log.Fatal(http.ListenAndServe(":9999", nil))
}