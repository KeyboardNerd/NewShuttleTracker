package YAST

import (
	"fmt"
	"time"
)

func measureTime(start time.Time, info string) {
	fmt.Printf("cost %v: %s\n", time.Since(start), info)
}
