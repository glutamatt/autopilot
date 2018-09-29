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

//Drive a vehicule
func (v *Vehicule) Drive(driving *Driving) {
	v.Velocity += driving.Thrust
	turningRadius := GetTurningRadius(driving)
	debug("turningRadius", turningRadius)
	turningAngle := TurningAngle(v.Velocity, turningRadius)
	debug("turningAngle", turningAngle)
	debug("turningAngle Deg", turningAngle*180/math.Pi)
	if turningAngle == 0 {
		v.Position.x += math.Cos(v.Rotation) * v.Velocity
		v.Position.y += math.Cos(math.Pi/2-v.Rotation) * v.Velocity
	} else {
		v.Rotation += turningAngle
		rotatePoint := RotatePoint(v.Position, turningRadius)
		Rotate(&v.Position, turningAngle, rotatePoint)
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

//RotatePoint for turn
func RotatePoint(pos Position, turningRadius float64) Position {
	angle := 90 - math.Atan(pos.x/pos.y)
	if turningRadius < 0 {
		return Position{
			x: pos.x + math.Sin(angle)*turningRadius,
			y: pos.y - math.Cos(angle)*turningRadius,
		}
	}
	return Position{
		x: pos.x - math.Sin(angle)*turningRadius,
		y: pos.y + math.Cos(angle)*turningRadius,
	}
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

		time.Sleep(50 * time.Millisecond)

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
