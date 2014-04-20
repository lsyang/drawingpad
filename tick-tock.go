/* tick-tock.go */
package main

import (
        "fmt"
        "log"
        "net/http"
        "time"
)

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
               $(document).ready(function () {
                 $("#output").append("Waiting for system time..");
                 setInterval("delayedPost()", 1000);
               });
               function delayedPost() {
                 $.post("http://localhost:9999/gettime?x=2&y=3&color=red", {inputVal: "lala"}, function(data, status) {
                 $("#output").empty();
                 $("#output").append(data);
                 });
               }
    </script>

    <script type="text/javascript">
      $(function() {
        canvas = document.getElementById("colors_sketch");
        canvas.width = document.body.clientWidth; //document.width is obsolete
        canvas.height = document.body.clientHeight; //document.height is obsolete

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
  <canvas id="colors_sketch" width="800" height="300"></canvas>
  
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
  x :=r.FormValue("x")
  y :=r.FormValue("y")
  fmt.Println(x, y)
  fmt.Fprint(w, page)
}

func main() {
        http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources")))) 
        http.HandleFunc("/time", handler)
        http.HandleFunc("/stroke", handlerStroke)
        http.HandleFunc("/gettime", handlerGetTime)
        log.Fatal(http.ListenAndServe(":9999", nil))
}