package main

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten"
)

const minTurningRadius = 20
const slowDownDefault = .5
const turnInc = .001

//Position for Items
type Position struct {
	x, y float64
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

//Drive a vehicule
func (v *Vehicule) Drive(driving *Driving) {
	v.Velocity += driving.Thrust
	v.Velocity -= slowDownDefault
	if v.Velocity < 0 {
		v.Velocity = 0
	}
	turningRadius := GetTurningRadius(driving)
	debug("turningRadius", turningRadius)
	turningAngle := TurningAngle(v.Velocity, turningRadius)
	debug("turningAngle", turningAngle)
	debug("turningAngle Deg", turningAngle*180/math.Pi)
	if turningAngle == 0 {
		v.Position = Position{
			x: v.x + math.Cos(v.Rotation)*v.Velocity,
			y: v.y + math.Cos(math.Pi/2-v.Rotation)*v.Velocity,
		}
	} else {
		v.Rotation += turningAngle
		targetRotate := Position{
			x: v.x + math.Cos(v.Rotation)*turningRadius,
			y: v.y + math.Cos(math.Pi/2-v.Rotation)*turningRadius,
		}
		Rotate(&targetRotate, math.Pi/-2, v.Position)
		Rotate(&v.Position, turningAngle, targetRotate)
	}
}

//Rotate point around
func Rotate(point *Position, angle float64, around Position) {
	s, c := math.Sin(angle), math.Cos(angle)
	point.x -= around.x
	point.y -= around.y

	xNew := point.x*c - point.y*s
	yNew := point.x*s + point.y*c
	point.x = xNew + around.x
	point.y = yNew + around.y
}

//TurningAngle from velocity and turningRadius
func TurningAngle(velocity, turningRadius float64) float64 {
	if velocity == 0 || turningRadius == 0 {
		return 0
	}
	return velocity / turningRadius
}

//GetTurningRadius from Driving.Turning
func GetTurningRadius(driving *Driving) float64 {
	if driving.Turning == 0 {
		return 0
	}

	cap := 1.0 * minTurningRadius
	if driving.Turning < 0 {
		cap *= -1
	}
	return 1/driving.Turning + cap
}

func debug(v ...interface{}) {
	fmt.Println(v...)
}

var square *ebiten.Image

func main() {

	v := Vehicule{Position: Position{y: 100}}
	drive := &Driving{}

	update := func(screen *ebiten.Image) error {

		println("----------------")

		if ebiten.IsKeyPressed(ebiten.KeyUp) {
			if drive.Thrust == 0 {
				drive.Thrust = 1
			}
			drive.Thrust += 1 / drive.Thrust
		} else {
			drive.Thrust = 0
		}

		if ebiten.IsKeyPressed(ebiten.KeyLeft) {
			if drive.Turning < 1 {
				if drive.Turning < 0 {
					drive.Turning = 0
				}
				drive.Turning += turnInc
			}
		} else {
			if ebiten.IsKeyPressed(ebiten.KeyRight) {
				if drive.Turning > -1 {
					if drive.Turning > 0 {
						drive.Turning = 0
					}
					drive.Turning -= turnInc
				}
			} else {
				drive.Turning = 0
			}
		}

		debug("drive.Turning", drive.Turning)

		screen.Fill(color.NRGBA{0xff, 0x00, 0x00, 0xff})
		if square == nil {
			square, _ = ebiten.NewImage(3, 3, ebiten.FilterNearest)
		}
		square.Fill(color.White)

		opts := &ebiten.DrawImageOptions{}
		v.Drive(drive)
		debug("v.Rotation", fmt.Sprintf("rot: %.3f", v.Rotation))
		scale := 10.0
		opts.GeoM.Reset()
		opts.GeoM.Translate(v.x/scale, v.y/scale)
		screen.DrawImage(square, opts)

		time.Sleep(100 * time.Millisecond)

		return nil
	}

	if err := ebiten.Run(update, 500, 500, 2, "Hello world!"); err != nil {
		panic(err)
	}

}
