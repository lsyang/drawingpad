//import epaxos
package main
import "projectserver"
import "strconv"
import "runtime"
import "fmt"
import "os"

func main() {


  runtime.GOMAXPROCS(4)

  const nservers = 2
  var kva []*projectserver.KVPaxos = make([]*projectserver.KVPaxos, nservers)
  var kvh []string = make([]string, nservers)
  //defer cleanup(kva)

  for i := 0; i < nservers; i++ {
    kvh[i] = port("basic", i)
  }
  for i := 0; i < nservers; i++ {
    kva[i] = projectserver.StartServer(kvh, i)
  }
  fmt.Println("making servers")

  //ck := projectserver.MakeClerk(kvh)
  const nclients=3
  var cka [nclients]*projectserver.Clerk
  for i := 0; i < nclients; i++ {
    cka[i] = projectserver.MakeClerk(kvh)
  }
  fmt.Println("making clients")
	
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