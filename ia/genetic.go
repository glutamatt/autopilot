package ia

import (
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/glutamatt/autopilot/model"
)

var driveSequenceLen = 5
var intervalTime = 300 * time.Millisecond
var randomPool = sync.Pool{
	New: func() interface{} {
		return rand.New(rand.NewSource(time.Now().UnixNano()))
	},
}
var distanceToLook = 30.0
var VehiculRadius float64
var BlocRadius float64

func Extrapol(vehicule *model.Vehicule, drive *model.Driving) []model.Position {
	pos := make([]model.Position, driveSequenceLen)
	v := copyVehicule(vehicule)
	for i := 0; i < driveSequenceLen; i++ {
		v.Drive(drive, intervalTime.Seconds())
		pos[i] = v.Position
	}
	return pos
}

func Genetic(vehicule *model.Vehicule, previousDrives []*model.Driving, path *[]model.Position, blocks *map[model.Position]bool) ([]*model.Driving, []model.Position) {
	filteredBlocks := filterBlocks(vehicule.Position, blocks)

	sequences := generateSequences(100, vehicule)
	if len(previousDrives) > 0 {
		sequences = append(sequences, &sequence{drives: previousDrives, vehicule: copyVehicule(vehicule)})
	}
	computeSequences(&sequences, filteredBlocks, path)
	sort.Slice(sequences, func(i, j int) bool { return sequences[i].cost < sequences[j].cost })

	//timer := time.NewTimer(time.Second/60 - 10*time.Millisecond)
	i := 20
	for {
		i--
		newSequences := []*sequence{}
		newSequences = append(newSequences, crossOver(10, &sequences, vehicule)...)
		newSequences = append(newSequences, mutateSequences(5, &sequences, vehicule)...)
		newSequences = append(newSequences, generateSequences(10, vehicule)...)
		computeSequences(&newSequences, filteredBlocks, path)
		sort.Slice(newSequences, func(i, j int) bool { return newSequences[i].cost < newSequences[j].cost })

		if i == 0 {
			return newSequences[0].drives, newSequences[0].positions
		}
		/*
			select {
			case <-timer.C:
				return newSequences[0].drives, newSequences[0].positions
			default:
				//sequences = newSequences
			}
		*/
		sequences = newSequences
	}
}

func mutateSequences(crossedLen int, sequences *[]*sequence, vehicule *model.Vehicule) []*sequence {
	crossed := make([]*sequence, crossedLen)
	sequencesLen := len(*sequences)
	random := randomPool.Get().(*rand.Rand)
	defer randomPool.Put(random)
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
	random := randomPool.Get().(*rand.Rand)
	defer randomPool.Put(random)
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
		if p.ManDist(vehicule) < distanceToLook {
			b = append(b, p)
		}
	}
	return &b
}

type sequence struct {
	drives    []*model.Driving
	positions []model.Position
	vehicule  model.Vehicule
	cost      float64
}

func (s *sequence) compute(interval time.Duration, blocks *[]model.Position, path *[]model.Position) {
	s.positions = make([]model.Position, len(s.drives))
	for i, d := range s.drives {
		s.vehicule.Drive(d, interval.Seconds())
		for _, pos := range *blocks {
			if s.vehicule.Collide(&pos, VehiculRadius+BlocRadius) {
				s.cost = math.Inf(1)
				s.positions = s.positions[:i]
				return
			}
		}
		s.positions[i] = s.vehicule.Position
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
	random := randomPool.Get().(*rand.Rand)
	defer randomPool.Put(random)
	return &model.Driving{
		Turning: random.Float64()*2 - 1,
		Thrust:  random.Float64()*2 - 1,
	}
}
