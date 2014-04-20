package main
import ("http";"io";"runtime")
func HelloServer(w http.ResponseWriter, req *http.Request) {
        io.WriteString(w, "hello, world!\n")
}
func main() {
  runtime.GOMAXPROCS(1)
        http.HandleFunc("/", HelloServer)
        http.ListenAndServe(":8080", nil)
}