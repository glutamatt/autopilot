package ia

import (
	"math"
	"math/rand"
	"time"

	"github.com/glutamatt/autopilot/model"
)

var driveSequenceLen = 2
var intervalTime = 500 * time.Millisecond
var random = rand.New(rand.NewSource(time.Now().UnixNano()))
var VehiculRadius float64
var BlocRadius float64

func Genetic(vehicule *model.Vehicule, path *[]model.Position, blocks *map[model.Position]bool) *model.Driving {
	var bestSeq *sequence
	filteredBlocks := filterBlocks(vehicule.Position, blocks)
	for i := 0; i < 1000; i++ {
		seq := sequence{
			drives: driveSequence(),
			vehicule: model.Vehicule{
				Position: vehicule.Position,
				Rotation: vehicule.Rotation,
				Velocity: vehicule.Velocity,
			},
		}
		seq.compute(intervalTime, filteredBlocks, path)
		if bestSeq == nil || bestSeq.cost > seq.cost {
			bestSeq = &seq
		}
	}

	return bestSeq.drives[0]
}

func filterBlocks(vehicule model.Position, blocks *map[model.Position]bool) *[]model.Position {
	b := []model.Position{}
	for p := range *blocks {
		if p.ManDist(vehicule) < 100 {
			b = append(b, p)
		}
	}
	return &b
}

type sequence struct {
	drives   []*model.Driving
	vehicule model.Vehicule
	cost     float64
}

func (s *sequence) compute(interval time.Duration, blocks *[]model.Position, path *[]model.Position) {
	for _, d := range s.drives {
		s.vehicule.Drive(d, interval.Seconds())
		for _, pos := range *blocks {
			if s.vehicule.Collide(&pos, VehiculRadius+BlocRadius) {
				s.cost = math.Inf(1)
				return
			}
		}
	}

	s.cost = (*path)[0].ManDist(s.vehicule.Position)
}

func driveSequence() []*model.Driving {
	s := make([]*model.Driving, driveSequenceLen)
	for i := 0; i < driveSequenceLen; i++ {
		s[i] = gene()
	}
	return s
}

func gene() *model.Driving {
	//max break -11 m/s/s
	//max boost 2.3 m/s/s
	return &model.Driving{
		Turning: random.Float64()*2 - 1,
		Thrust:  -11.0 + (2.3+11.0)*random.Float64(),
	}
}
