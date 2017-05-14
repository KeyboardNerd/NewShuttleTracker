package pkg

import (
	"fmt"
	"time"
)

func MeasureTime(start time.Time, info string) {
	fmt.Printf("cost %v: %s\n", time.Since(start), info)
}
