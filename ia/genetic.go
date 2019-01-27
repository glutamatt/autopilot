package ia

import (
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/glutamatt/autopilot/model"
)

var randomPool = sync.Pool{
	New: func() interface{} {
		return rand.New(rand.NewSource(time.Now().UnixNano()))
	},
}

var distanceTPredict = 30.0
var angleOkThreshold = math.Pi / 6
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

	anglePosition := vehicule.Position.Angle(target)
	angleToCheck := math.Abs(math.Mod(vehicule.Rotation, 2*math.Pi) - anglePosition)
	if angleToCheck > math.Pi {
		angleToCheck = (2 * math.Pi) - angleToCheck
	}

	if vehicule.Velocity < 1.5 && angleToCheck > angleOkThreshold {
		sess.costF = costByAngleToTarget(vehicule, anglePosition)
	}

	sess.sequences = generateSequences(sess.driveSequenceLen, 200, vehicule)
	if len(previousDrives) > 0 {
		if len(previousDrives) < sess.driveSequenceLen {
			previousDrives = append(previousDrives, driveSequence(sess.driveSequenceLen-len(previousDrives))...)
		}
		sess.sequences = append(sess.sequences, &sequence{drives: previousDrives[:sess.driveSequenceLen], vehicule: copyVehicule(vehicule)})
	}

	sess.computeSequences()
	i := 30
	for {
		i--
		sess.naturalSelection()
		sess.computeSequences()
		if i == 0 {
			return sess.sequences[0].drives, sess.sequences[0].positions
		}
	}
}

func costByAngleToTarget(vehicule *model.Vehicule, anglePositions float64) costFunc {
	return func(s *sequence, target model.Position) float64 {
		invert := 0.0
		if (vehicule.Velocity < 0 && s.drives[0].Thrust > 0) || (vehicule.Velocity > 0 && s.drives[0].Thrust < 0) {
			invert = 20
		}
		angle := math.Abs(math.Mod(s.vehicule.Rotation, 2*math.Pi) - anglePositions)
		if angle > math.Pi {
			angle = (2 * math.Pi) - angle
		}
		return angle + invert
	}
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

func (sess *session) naturalSelection() {
	newSequences := []*sequence{}
	newSequences = append(newSequences, crossOver(sess, 10, &sess.sequences, sess.vehicule)...)
	newSequences = append(newSequences, mutateSequences(sess, 5, &sess.sequences, sess.vehicule)...)
	newSequences = append(newSequences, generateSequences(sess.driveSequenceLen, 10, sess.vehicule)...)
	sess.sequences = newSequences
}

func mutateSequences(sess *session, crossedLen int, sequences *[]*sequence, vehicule *model.Vehicule) []*sequence {
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
		crossed[i].drives[random.Intn(sess.driveSequenceLen)] = gene()
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

func generateSequences(driveSequenceLen int, len int, vehicule *model.Vehicule) []*sequence {
	sequences := make([]*sequence, len)
	for i := range sequences {
		sequences[i] = &sequence{drives: driveSequence(driveSequenceLen), vehicule: copyVehicule(vehicule)}
	}
	return sequences
}

func filterBlocks(vehicule model.Position, blocks *map[model.Position]bool) *[]model.Position {
	b := []model.Position{}
	for p := range *blocks {
		if p.ManDist(vehicule) < distanceTPredict {
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

type costFunc func(*sequence, model.Position) float64

func costByFarTargetDistance(s *sequence, target model.Position) float64 {
	return target.ManDist(s.vehicule.Position)
}

func (s *sequence) compute(sess *session) {
	s.positions = make([]model.Position, len(s.drives))
	for i, d := range s.drives {
		s.vehicule.Drive(d, sess.drivesInterval.Seconds())
		if s.vehicule.Velocity > 16 {
			s.cost += 20
		}
		for _, pos := range *sess.blocks {
			if s.vehicule.Collide(&pos, VehiculRadius+BlocRadius) {
				s.cost += 100
			}
		}
		if len(sess.vehiculesFuturePositions) >= i+1 {
			for pos := range sess.vehiculesFuturePositions[i] {
				if s.vehicule.Collide(&pos, VehiculRadius+VehiculRadius) {
					s.cost += 100
				}
			}
		}
		s.positions[i] = s.vehicule.Position
		s.cost += sess.costF(s, sess.target)
	}
}

func driveSequence(driveSequenceLen int) []*model.Driving {
	s := make([]*model.Driving, driveSequenceLen)
	for i := 0; i < driveSequenceLen; i++ {
		s[i] = gene()
	}
	return s
}

func gene() *model.Driving {
	random := randomPool.Get().(*rand.Rand)
	defer randomPool.Put(random)
	d := &model.Driving{
		Turning: math.Tanh(random.Float64()*2 - 1),
		Thrust:  math.Tanh(random.Float64()*2 - 1),
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
