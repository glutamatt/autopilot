package ia

import (
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/glutamatt/autopilot/model"
)

var driveSequenceLen = 4
var intervalTime = 500 * time.Millisecond
var random = rand.New(rand.NewSource(time.Now().UnixNano()))
var VehiculRadius float64
var BlocRadius float64

func Genetic(vehicule *model.Vehicule, path *[]model.Position, blocks *map[model.Position]bool) *model.Driving {
	filteredBlocks := filterBlocks(vehicule.Position, blocks)

	sequences := generateSequences(100, vehicule)
	computeSequences(&sequences, filteredBlocks, path)
	sort.Slice(sequences, func(i, j int) bool { return sequences[i].cost < sequences[j].cost })

	timer := time.NewTimer(time.Second/60 - 10*time.Millisecond)
	for {
		newSequences := []*sequence{}
		newSequences = append(newSequences, crossOver(10, &sequences, vehicule)...)
		newSequences = append(newSequences, mutateSequences(5, &sequences, vehicule)...)
		newSequences = append(newSequences, generateSequences(10, vehicule)...)
		computeSequences(&newSequences, filteredBlocks, path)
		sort.Slice(newSequences, func(i, j int) bool { return newSequences[i].cost < newSequences[j].cost })
		select {
		case <-timer.C:
			return newSequences[0].drives[0]
		default:
			sequences = newSequences
		}
	}
}

func mutateSequences(crossedLen int, sequences *[]*sequence, vehicule *model.Vehicule) []*sequence {
	crossed := make([]*sequence, crossedLen)
	sequencesLen := len(*sequences)
	for i := 0; i < crossedLen; i++ {
		crossed[i] = &sequence{
			drives:   make([]*model.Driving, driveSequenceLen),
			vehicule: copyVehicule(vehicule),
		}
		copy(crossed[i].drives, (*sequences)[i%sequencesLen].drives)
		crossed[i].drives[random.Intn(driveSequenceLen)] = gene()
	}
	return crossed
}

func crossOver(crossedLen int, sequences *[]*sequence, vehicule *model.Vehicule) []*sequence {
	crossed := make([]*sequence, crossedLen)
	sequencesLen := len(*sequences)
	for i := 0; i < crossedLen; i++ {
		fatherLen := random.Intn(driveSequenceLen)
		crossed[i] = &sequence{
			drives:   append((*sequences)[i%sequencesLen].drives[:fatherLen], (*sequences)[random.Intn(sequencesLen)].drives[fatherLen:]...),
			vehicule: copyVehicule(vehicule)}
	}
	return crossed
}

func copyVehicule(vehicule *model.Vehicule) model.Vehicule {
	return model.Vehicule{
		Position: vehicule.Position,
		Rotation: vehicule.Rotation,
		Velocity: vehicule.Velocity,
	}
}

func generateSequences(len int, vehicule *model.Vehicule) []*sequence {
	sequences := make([]*sequence, len)
	for i := range sequences {
		sequences[i] = &sequence{drives: driveSequence(), vehicule: copyVehicule(vehicule)}
	}
	return sequences
}

func computeSequences(sequences *[]*sequence, filteredBlocks *[]model.Position, path *[]model.Position) {
	wg := sync.WaitGroup{}
	wg.Add(len(*sequences))
	for _, seq := range *sequences {
		go func(s *sequence) { s.compute(intervalTime, filteredBlocks, path); wg.Done() }(seq)
	}
	wg.Wait()
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
