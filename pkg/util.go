package pkg

import (
	"log"
	"time"
)

func MeasureTime(start time.Time, info string) {
	log.Printf("cost %v: %s\n", time.Since(start), info)
}
