package ia

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/glutamatt/autopilot/model"
)

//var driveSequenceLen = 5
//var intervalTime = (400 * time.Millisecond).Seconds()
var randomPool = sync.Pool{
	New: func() interface{} {
		return rand.New(rand.NewSource(time.Now().UnixNano()))
	},
}
var distanceToLook = 30.0
var distanceTPredict = 20.0
var angleOkThreshold = math.Pi / 4
var VehiculRadius float64
var BlocRadius float64

type session struct {
	vehicule                 *model.Vehicule
	driveSequenceLen         int
	drivesInterval           time.Duration
	previousDrives           []*model.Driving
	target                   model.Position
	blocks                   *[]model.Position
	vehiculesFuturePositions []map[model.Position]bool
	sequences                []*sequence
	costF                    costFunc
}

func (sess *session) computeSequences() {
	wg := sync.WaitGroup{}
	wg.Add(len(sess.sequences))
	for _, seq := range sess.sequences {
		go func(s *sequence) {
			s.compute(sess)
			wg.Done()
		}(seq)
	}
	wg.Wait()
	sort.Slice(sess.sequences, func(i, j int) bool { return sess.sequences[i].cost < sess.sequences[j].cost })
}

func (sess *session) naturalSelection(forceXtremDrives bool) {
	newSequences := []*sequence{}
	newSequences = append(newSequences, crossOver(sess, 10, &sess.sequences, sess.vehicule)...)
	newSequences = append(newSequences, mutateSequences(sess, 5, &sess.sequences, sess.vehicule, forceXtremDrives)...)
	newSequences = append(newSequences, generateSequences(sess.driveSequenceLen, 10, sess.vehicule, forceXtremDrives)...)
	sess.sequences = newSequences
}

func driveInterval(velocity float64) time.Duration {
	if velocity < 0 {
		velocity *= -1
	}
	if velocity < 5 {
		velocity = 5
	}
	return time.Duration(float64(time.Second) * VehiculRadius / velocity)
}

func Genetic(
	vehicule *model.Vehicule,
	previousDrives []*model.Driving,
	target model.Position,
	blocks *map[model.Position]bool,
	vehiculesFuturePositions []map[model.Position]bool,
) ([]*model.Driving, []model.Position) {

	sess := session{
		vehicule:                 vehicule,
		blocks:                   filterBlocks(vehicule.Position, blocks),
		vehiculesFuturePositions: vehiculesFuturePositions,
		target:                   target,
		costF:                    costByFarTargetDistance,
		driveSequenceLen:         int(distanceTPredict/VehiculRadius) + 1,
		drivesInterval:           driveInterval(vehicule.Velocity),
	}

	forceXtremDrives := false
	anglePositions := vehicule.Position.Angle(target)
	angleToCheck := math.Abs(math.Mod(vehicule.Rotation, 2*math.Pi) - anglePositions)
	if angleToCheck > math.Pi {
		angleToCheck = (2 * math.Pi) - angleToCheck
	}
	fmt.Printf("angle to target : %.0f ", angleToCheck*360/2/math.Pi)

	if vehicule.Velocity < 1.5 && angleToCheck > angleOkThreshold {
		println("must correct")
		forceXtremDrives = true
		sess.costF = func(s *sequence, target model.Position) float64 {
			invert := 0.0
			if (vehicule.Velocity < 0 && s.drives[0].Thrust > 0) || (vehicule.Velocity > 0 && s.drives[0].Thrust < 0) {
				invert = 10
			}
			angle := math.Abs(math.Mod(s.vehicule.Rotation, 2*math.Pi) - anglePositions)
			if angle > math.Pi {
				angle = (2 * math.Pi) - angle
			}
			return angle + invert
		}
	} else {
		println("")
	}

	sess.sequences = generateSequences(sess.driveSequenceLen, 100, vehicule, forceXtremDrives)
	if len(previousDrives) > 0 {
		if len(previousDrives) < sess.driveSequenceLen {
			previousDrives = append(previousDrives, driveSequence(sess.driveSequenceLen-len(previousDrives), forceXtremDrives)...)
		}
		sess.sequences = append(sess.sequences, &sequence{drives: previousDrives[:sess.driveSequenceLen], vehicule: copyVehicule(vehicule)})
	}

	sess.computeSequences()
	i := 20
	for {
		i--
		sess.naturalSelection(forceXtremDrives)
		sess.computeSequences()
		if i == 0 {
			fmt.Printf("predict angle : %.0f\n", sess.sequences[0].cost*360/2/math.Pi)
			return sess.sequences[0].drives, sess.sequences[0].positions
		}
	}
}

func mutateSequences(sess *session, crossedLen int, sequences *[]*sequence, vehicule *model.Vehicule, forceXtremDrives bool) []*sequence {
	crossed := make([]*sequence, crossedLen)
	sequencesLen := len(*sequences)
	random := randomPool.Get().(*rand.Rand)
	defer randomPool.Put(random)
	for i := 0; i < crossedLen; i++ {
		crossed[i] = &sequence{
			drives:   make([]*model.Driving, sess.driveSequenceLen),
			vehicule: copyVehicule(vehicule),
		}
		copy(crossed[i].drives, (*sequences)[i%sequencesLen].drives)
		crossed[i].drives[random.Intn(sess.driveSequenceLen)] = gene(forceXtremDrives)
	}
	return crossed
}

func crossOver(sess *session, crossedLen int, sequences *[]*sequence, vehicule *model.Vehicule) []*sequence {
	crossed := make([]*sequence, crossedLen)
	random := randomPool.Get().(*rand.Rand)
	defer randomPool.Put(random)
	sequencesLen := len(*sequences)
	for i := 0; i < crossedLen; i++ {
		fatherLen := random.Intn(sess.driveSequenceLen)
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

func generateSequences(driveSequenceLen int, len int, vehicule *model.Vehicule, forceXtremDrives bool) []*sequence {
	sequences := make([]*sequence, len)
	for i := range sequences {
		sequences[i] = &sequence{drives: driveSequence(driveSequenceLen, forceXtremDrives), vehicule: copyVehicule(vehicule)}
	}
	return sequences
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

func killSequence(s *sequence, i int) {
	s.cost = math.Inf(1)
	s.positions = s.positions[:i]
}

type costFunc func(*sequence, model.Position) float64

func costByFarTargetDistance(s *sequence, target model.Position) float64 {
	return target.ManDist(s.vehicule.Position)
}

func (s *sequence) compute(sess *session) {
	s.positions = make([]model.Position, len(s.drives))
	for i, d := range s.drives {
		s.vehicule.Drive(d, sess.drivesInterval.Seconds())
		if s.vehicule.Velocity > 13.80 {
			killSequence(s, i)
			return
		}
		for _, pos := range *sess.blocks {
			if s.vehicule.Collide(&pos, VehiculRadius+BlocRadius) {
				killSequence(s, i)
				return
			}
		}
		if len(sess.vehiculesFuturePositions) >= i+1 {
			for pos := range sess.vehiculesFuturePositions[i] {
				if s.vehicule.Collide(&pos, VehiculRadius+VehiculRadius) {
					killSequence(s, i)
					return
				}
			}
		}
		s.positions[i] = s.vehicule.Position
	}

	s.cost = sess.costF(s, sess.target)
}

func driveSequence(driveSequenceLen int, forceXtremDrives bool) []*model.Driving {
	s := make([]*model.Driving, driveSequenceLen)
	for i := 0; i < driveSequenceLen; i++ {
		s[i] = gene(forceXtremDrives)
	}
	return s
}

func gene(forceXtremDrives bool) *model.Driving {
	random := randomPool.Get().(*rand.Rand)
	defer randomPool.Put(random)
	d := &model.Driving{
		Turning: math.Tanh(random.Float64()*2 - 1),
		Thrust:  math.Tanh(random.Float64()*2 - 1),
	}
	if forceXtremDrives {
		if d.Thrust > 0 {
			d.Thrust = 1
		} else {
			d.Thrust = -1
		}
		if d.Turning > 0 {
			d.Turning = 1
		} else {
			d.Turning = -1
		}
	}
	return d
}

//Extrapol the future positions of the vehicule from future drivings
func Extrapol(vehicule *model.Vehicule, drive *model.Driving) []model.Position {
	driveSequenceLen := int(distanceTPredict/VehiculRadius) + 1
	drivesInterval := driveInterval(vehicule.Velocity)
	pos := make([]model.Position, driveSequenceLen)
	v := copyVehicule(vehicule)
	for i := 0; i < driveSequenceLen; i++ {
		v.Drive(drive, drivesInterval.Seconds())
		pos[i] = v.Position
	}
	return pos
}
