package main

import (
	"bufio"
	"fmt"
	"image/color"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten"
)

const minTurningRadius = 10

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

//GlobalAngle of a pos
func (v Position) GlobalAngle() float64 {
	if v.y == 0 {
		return 0
	}
	return math.Atan(v.y / v.x)
}

//Drive a vehicule
func (v *Vehicule) Drive(driving *Driving) {
	v.Velocity += driving.Thrust
	turningRadius := GetTurningRadius(driving)
	debug("turningRadius", turningRadius)
	turningAngle := TurningAngle(v.Velocity, turningRadius)
	debug("turningAngle", turningAngle)
	debug("turningAngle Deg", turningAngle*180/math.Pi)
	if turningAngle == 0 {
		globalAngle := v.GlobalAngle()
		v.Position.x += math.Cos(globalAngle) * v.Velocity
		v.Position.y += math.Cos(math.Pi/2-globalAngle) * v.Velocity
	} else {
		coordLocalTurning := CoordLocalTurning(turningRadius, turningAngle)
		debug("coordLocalTurning", coordLocalTurning)
		v.Position = LocalToGlobal(coordLocalTurning, turningRadius, turningAngle, v)
		debug("LocalToGlobal", v.Position)
		v.Rotation += turningAngle
	}
}

//TurningAngle from velocity and turningRadius
func TurningAngle(velocity, turningRadius float64) float64 {
	if velocity == 0 || turningRadius == 0 {
		return 0
	}
	return velocity / turningRadius
}

//LocalToGlobal transpose local to global position
func LocalToGlobal(localPos Position, turningRadius, turningAngle float64, v *Vehicule) Position {
	globalAngle := v.GlobalAngle()
	debug("globalAngle", globalAngle)

	return Position{
		x: math.Cos(globalAngle-turningAngle)*turningRadius + v.x,
		y: math.Cos(math.Pi/2+turningAngle*2-globalAngle)*turningRadius + v.y,
	}
}

//CoordLocalTurning on a local reference
func CoordLocalTurning(turningRadius, turningAngle float64) Position {
	return Position{
		x: math.Abs(turningRadius) * math.Sin(turningAngle),
		y: math.Abs(turningRadius) * math.Cos(turningAngle),
	}
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

	v := Vehicule{}
	drives := make(chan Driving)

	update := func(screen *ebiten.Image) error {
		screen.Fill(color.NRGBA{0xff, 0x00, 0x00, 0xff})

		if square == nil {
			square, _ = ebiten.NewImage(3, 3, ebiten.FilterNearest)
		}
		square.Fill(color.White)

		opts := &ebiten.DrawImageOptions{}

		select {
		case d := <-drives:
			v.Drive(&d)
		default:
			v.Drive(&Driving{})
		}

		debug("v", fmt.Sprintf("rot: %.1f", v.Rotation))

		scale := 10.0
		opts.GeoM.Reset()
		opts.GeoM.Translate(v.x/scale, v.y/scale)
		screen.DrawImage(square, opts)

		time.Sleep(100 * time.Millisecond)

		return nil
	}

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			in := scanner.Text()
			if in != "" {
				d := Driving{}
				pieces := strings.Split(in, " ")
				turning, err := strconv.ParseFloat(pieces[0], 64)
				if err != nil {
					panic(err)
				}
				d.Turning = turning
				if len(pieces) > 1 {
					thurst, err := strconv.ParseFloat(pieces[1], 64)
					if err != nil {
						panic(err)
					}
					d.Thrust = thurst
				}
				drives <- d
			}
		}
	}()

	if err := ebiten.Run(update, 500, 500, 2, "Hello world!"); err != nil {
		panic(err)
	}

}
