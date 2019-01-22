package model

import "math"

//Position for Items
type Position struct {
	X, Y float64
}

func (p *Position) Gap(dist int) {
	p.X = math.Round(p.X/float64(dist)) * float64(dist)
	p.Y = math.Round(p.Y/float64(dist)) * float64(dist)
}

func (p Position) ManDist(to Position) float64 {
	return math.Abs(p.X-to.X) + math.Abs(p.Y-to.Y)
}

//Driving instruction
type Driving struct {
	Turning float64
	Thrust  float64
}

//Vehicule on the map
type Vehicule struct {
	Position
	Rotation float64
	Velocity float64
}

var minTurningRadius float64 = 5

//SetMinTurningRadius on ini
func SetMinTurningRadius(r float64) {
	minTurningRadius = r
}

var SecurityDistance float64

//Collide a vehicule : is my vehicule farrer than dist from other
func (v *Vehicule) Collide(other *Position, dist float64) bool {
	dist += SecurityDistance
	xdist := math.Abs(v.X - other.X)

	if xdist > dist {
		return false
	}

	ydist := math.Abs(v.Y - other.Y)

	if xdist > dist {
		return false
	}

	return math.Sqrt(xdist*xdist+ydist*ydist) <= dist
}

var adherenceMaxDotgravity float64

func SetAdherenceMax(adherenceMax float64) {
	adherenceMaxDotgravity = adherenceMax * 9.81
}

var BoostMax float64
var BreakMax float64
var ReverseMaxSpeed float64

//Drive a vehicule
func (v *Vehicule) Drive(driving *Driving, seconds float64) (wheelTurn float64) {
	thrustCoef := BoostMax
	if (driving.Thrust < 0 && v.Velocity > 0) || (driving.Thrust > 0 && v.Velocity < 0) {
		thrustCoef = BreakMax
	}
	v.Velocity += driving.Thrust * thrustCoef * seconds
	if v.Velocity < -ReverseMaxSpeed {
		v.Velocity = -ReverseMaxSpeed
	}
	if v.Velocity == 0 {
		return
	}
	instantDist := v.Velocity * seconds
	if driving.Turning == 0 {
		v.Position = Position{
			X: v.X + math.Cos(v.Rotation)*instantDist,
			Y: v.Y + math.Cos(math.Pi/2-v.Rotation)*instantDist,
		}
		return
	}

	minRadiusRelVel := math.Max(minTurningRadius, v.Velocity*v.Velocity/(adherenceMaxDotgravity))
	wheelTurn = driving.Turning * minTurningRadius / minRadiusRelVel

	turningRadius := minRadiusRelVel / driving.Turning
	turningAngle := instantDist / turningRadius
	v.Rotation = math.Mod(v.Rotation+turningAngle, 2*math.Pi)

	rotateCenterAngle := v.Rotation + math.Pi/2

	rotateCenterFromV := Position{
		X: math.Cos(rotateCenterAngle) * turningRadius,
		Y: math.Sin(rotateCenterAngle) * turningRadius,
	}

	vehiculePosFromRotatePoint := Position{
		X: 0 - rotateCenterFromV.X,
		Y: 0 - rotateCenterFromV.Y,
	}

	s, c := math.Sin(turningAngle), math.Cos(turningAngle)
	v.Position = Position{
		X: vehiculePosFromRotatePoint.X*c - vehiculePosFromRotatePoint.Y*s + rotateCenterFromV.X + v.X,
		Y: vehiculePosFromRotatePoint.X*s + vehiculePosFromRotatePoint.Y*c + rotateCenterFromV.Y + v.Y,
	}

	return
}

var vehiculRadius float64
var blocRadius float64

func InitRadiusCar(carWidth, carHeight int) float64 {
	vehiculRadius = math.Sqrt(float64(carWidth*carWidth)/4.0 + float64(carHeight*carHeight)/4.0)
	return vehiculRadius
}

func InitRadiusBlock(blockBorder int) float64 {
	blocRadius = math.Sqrt(float64(blockBorder*blockBorder) / 2)
	return blocRadius
}

func Collisions(vehicules []*Vehicule, blocks map[Position]bool) map[int]struct{} {
	collisions := make(map[int]struct{})

	for i1, v1 := range vehicules {
		for b := range blocks {
			if v1.Collide(&b, vehiculRadius+blocRadius) {
				collisions[i1] = struct{}{}
			}
		}
		for i2, v2 := range vehicules[i1+1:] {
			if v1.Collide(&v2.Position, vehiculRadius*2) {
				collisions[i1] = struct{}{}
				collisions[i2+i1+1] = struct{}{}
			}
		}
	}

	return collisions
}
