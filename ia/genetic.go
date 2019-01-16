package ia

import (
	"math/rand"
	"time"

	"github.com/glutamatt/autopilot/model"
)

func Genetic() *model.Driving {
	return gene()
}

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func gene() *model.Driving {
	return &model.Driving{
		Turning: random.Float64()*2 - 1,
		Thrust:  random.Float64() * 2.3,
	}
}
