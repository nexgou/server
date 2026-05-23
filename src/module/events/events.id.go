package events

import (
	"fmt"
	"sync/atomic"
)

var idCounter uint64

func generateID() string {
	n := atomic.AddUint64(&idCounter, 1)
	return fmt.Sprintf("sub_%d", n)
}
