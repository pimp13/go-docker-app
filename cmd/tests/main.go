package main

import (
	"fmt"
	"math/rand/v2"
	"time"
)

func main() {
	for i := 0; i <= 100; i++ {
		randomID := rand.IntN(99999)
		fmt.Printf("✅ Successfully updated via Docker (ID %d) and is now ready for operation.\n", randomID)
		if randomID%3 == 0 {
			fmt.Printf("❌ Failed to update via Docker (ID %d) and is not ready for operation.\n", randomID)
			time.Sleep(1 * time.Second)
			continue
		}
		time.Sleep(1 * time.Second)
	}
}
