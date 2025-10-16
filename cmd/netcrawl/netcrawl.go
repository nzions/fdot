package main

import (
	"fmt"

	"github.com/nzions/dsjdb"
	"github.com/nzions/fdot/pkg/fdh/user"
)

// ssh to remote switch, show ver, show run, capture output to file
func main() {
	fmt.Println("hello!")
	jstore, err := dsjdb.NewJSDB(user.CurrentUser.NetworkDir)
	if err != nil {
		fmt.Printf("Error initializing dsjdb: %v\n", err)
		return
	}
	fmt.Println(jstore.IsCompressed())
}
