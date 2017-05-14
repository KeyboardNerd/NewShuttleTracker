package main

import (
	"fmt"
	"os"

	"github.com/keyboardnerd/yastserver"
)

func main() {
	fmt.Print("YAST v0.5\n")
	if len(os.Args) != 2 {
		panic("usage: ./yast <config file>")
	}
	config := YAST.Loadconfig(os.Args[1])
	YAST.Boot(config)
}
