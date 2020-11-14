// Usage: ./git-http-dumb -repo=/xxx/xxx/xxx/
package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {
	dir := flag.String("repo", "/Users/zoker/Tmp/xxxx333/testrepo", "Specify a repositories root dir.")
	flag.Parse()
	http.Handle("/", http.FileServer(http.Dir(*dir)))
	fmt.Printf("Dumb http server start at port 8881 on dir %s \n", *dir)
	_ = http.ListenAndServe(":8881", nil)
}