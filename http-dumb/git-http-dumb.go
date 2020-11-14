// Usage: ./http-dumb -repo=/xxx/xxx/xxx/ -port=8881
package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {
	repo := flag.String("repo", "/Users/zoker/Tmp/repositories", "Specify a repositories root dir.")
	port := flag.String("port", "8881", "Specify a port to start process.")
	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir(*repo)))
	fmt.Printf("Dumb http server start at port %s on dir %s \n", *port, *repo)
	_ = http.ListenAndServe(fmt.Sprintf(":%s", *port), nil)
}