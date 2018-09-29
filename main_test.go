package main

import (
	"fmt"
	"math"
	"testing"
)

func TestCoordLocalTurning(t *testing.T) {
	fmt.Printf("%#v\n", CoordLocalTurning(10, math.Pi/8))
	fmt.Printf("%#v\n", CoordLocalTurning(10, math.Pi/-8))
	fmt.Printf("%#v\n", CoordLocalTurning(10, math.Pi/8+math.Pi))
}
