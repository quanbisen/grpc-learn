package sample

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestRandom(t *testing.T) {
	rand.Seed(time.Now().Unix())
	fmt.Println(rand.Intn(10))
	fmt.Println(rand.Int())
}
