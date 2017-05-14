package main

import (
	"fmt"
	"os"

	yast "github.com/keyboardnerd/yastserver"
	"github.com/keyboardnerd/yastserver/api"
)

func main() {
	fmt.Print("YAST v0.5\n")
	if len(os.Args) != 2 {
		panic("usage: ./yast <config file>")
	}
	config := api.Loadconfig(os.Args[1])
	yast.Boot(config)
}
