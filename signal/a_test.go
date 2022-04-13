package signal

import (
	"math/rand"
	"testing"
	"time"
)

func TestRandSleep(t *testing.T) {
	rand.Seed(time.Now().Unix())
	RandSleep(100 * time.Millisecond)
}
