package rand

import (
	"math/rand"
)

// Int returns a pseudorandom int between min and max (inclusive)
func Int(min, max int) int {
	return rand.Intn((max+1)-min) + min
}
