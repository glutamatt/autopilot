package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"sync"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const minTurningRadius = 12
const turnInc = .02
const carWidth = 5
const carHeight = 2
const uiScale = 3.0
const uiWidth = 500
const uiHeight = 500

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
func (v *Vehicule) Drive(driving *Driving, seconds float64) {
	v.Velocity += driving.Thrust * seconds
	if v.Velocity < 0 {
		v.Velocity = 0
	}
	if v.Velocity == 0 {
		return
	}
	instantDist := v.Velocity * seconds
	if driving.Turning == 0 {
		v.Position = Position{
			x: v.x + math.Cos(v.Rotation)*instantDist,
			y: v.y + math.Cos(math.Pi/2-v.Rotation)*instantDist,
		}
		return
	}
	turningRadius := minTurningRadius / driving.Turning
	turningAngle := instantDist / turningRadius
	v.Rotation = math.Mod(v.Rotation+turningAngle, 2*math.Pi)

	rotateCenterAngle := v.Rotation + math.Pi/2

	rotateCenterFromV := Position{
		x: math.Cos(rotateCenterAngle) * turningRadius,
		y: math.Sin(rotateCenterAngle) * turningRadius,
	}

	vehiculePosFromRotatePoint := Position{
		x: 0 - rotateCenterFromV.x,
		y: 0 - rotateCenterFromV.y,
	}

	s, c := math.Sin(turningAngle), math.Cos(turningAngle)
	v.Position = Position{
		x: vehiculePosFromRotatePoint.x*c - vehiculePosFromRotatePoint.y*s + rotateCenterFromV.x + v.x,
		y: vehiculePosFromRotatePoint.x*s + vehiculePosFromRotatePoint.y*c + rotateCenterFromV.y + v.y,
	}
}

func debug(v ...interface{}) {
	fmt.Println(v...)
}

func createRandomVehicule() *Vehicule {
	return &Vehicule{
		Position: Position{
			x: rand.Float64() * uiHeight,
			y: rand.Float64() * uiWidth,
		},
		Rotation: rand.Float64() * math.Pi * 2,
	}
}

func main() {
	vehicules := make([]*Vehicule, 500)
	vehiculeImage, _ := ebiten.NewImage(int(carWidth*uiScale), int(carHeight*uiScale), ebiten.FilterNearest)
	vehiculeImage.Fill(color.White)
	for i := range vehicules {
		vehicules[i] = createRandomVehicule()
	}

	drive := &Driving{}

	update := func(screen *ebiten.Image) error {
		inputControls(drive)
		screen.Fill(color.NRGBA{0x88, 0x00, 0x00, 0xff})

		wg := sync.WaitGroup{}
		wg.Add(len(vehicules))
		optsChan := make(chan ebiten.DrawImageOptions)

		for _, v := range vehicules {
			go func(v *Vehicule) {
				v.Drive(drive, 1.0/60)
				opts := ebiten.DrawImageOptions{}
				opts.GeoM.Translate(carWidth*uiScale/-4, carHeight*uiScale/-2)
				opts.GeoM.Rotate(v.Rotation * -1)
				opts.GeoM.Translate(carWidth*uiScale/4, carHeight*uiScale/2)
				opts.GeoM.Translate(v.x*uiScale, v.y*uiScale*-1+uiHeight)
				optsChan <- opts
				wg.Done()
			}(v)
		}

		go func() {
			for opts := range optsChan {
				screen.DrawImage(vehiculeImage, &opts)
			}
		}()

		wg.Wait()
		close(optsChan)

		ebitenutil.DebugPrint(screen, fmt.Sprintf(
			"%#v\n%.1f km/h\n%.2f°\nfps:%.0f\nvcount: %d",
			drive, vehicules[0].Velocity/1000*60*60, vehicules[0].Rotation*180/math.Pi, ebiten.CurrentFPS(), len(vehicules)))
		return nil
	}

	if err := ebiten.Run(update, uiWidth, uiHeight, 2, "Hello world!"); err != nil {
		panic(err)
	}
}

func inputControls(drive *Driving) {
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		drive.Thrust = -10
	} else {
		if ebiten.IsKeyPressed(ebiten.KeyUp) {
			drive.Thrust = 5
		} else {
			drive.Thrust = 0
		}
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
			if math.Abs(drive.Turning-turnInc) < turnInc*2 {
				drive.Turning = 0
			}
			if drive.Turning > 0 {
				drive.Turning -= turnInc * 2
			}
			if drive.Turning < 0 {
				drive.Turning += turnInc * 2
			}
		}
	}
}
