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
const uiScale = 2
const groundWidth = 600
const groundHeight = 300

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

//Collide a vehicule with a position and the sqaure of the 2 radius sum
func (v *Vehicule) Collide(other *Position, sqrtDist float64) bool {
	return math.Abs(v.x-other.x)+math.Abs(v.y-other.y) <= sqrtDist
}

//Drive a vehicule
func (v *Vehicule) Drive(driving *Driving, seconds float64) {
	v.Velocity += driving.Thrust * seconds
	/*if v.Velocity < 0 {
		v.Velocity = 0
	}*/
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
			x: rand.Float64()*groundWidth/2 + groundWidth/4,
			y: rand.Float64()*groundHeight/-2 - groundHeight/4,
		},
		Rotation: rand.Float64() * math.Pi * 2,
	}
}

func main() {
	//rand.Seed(time.Now().Unix())
	vehicules := make([]*Vehicule, 100)
	blocks := []*Position{}
	vehiculeImage, _ := ebiten.NewImage(int(carWidth*uiScale), int(carHeight*uiScale), ebiten.FilterNearest)
	blockImage, _ := ebiten.NewImage(int(carWidth*uiScale), int(carWidth*uiScale), ebiten.FilterNearest)
	blockImage.Fill(color.NRGBA{0xBB, 0xBB, 0xBB, 0xff})
	vehiculeImage.Fill(color.NRGBA{0xFF, 0xFF, 0xFF, 0xff})
	for i := range vehicules {
		vehicules[i] = createRandomVehicule()
	}
	vehiculeVehiculeRadiusSqrt := float64(carWidth*carWidth)/4.0 + float64(carHeight*carHeight)/4.0

	drive := &Driving{}

	update := func(screen *ebiten.Image) error {
		inputControls(drive)
		screen.Fill(color.NRGBA{0x00, 0x00, 0x88, 0xff})

		/*STEPS
		1.move
		2.collide
		3. render
		*/

		wg := sync.WaitGroup{}
		wg.Add(len(vehicules))
		optsChan := make(chan ebiten.DrawImageOptions)

		for _, b := range blocks {
			opts := ebiten.DrawImageOptions{}
			opts.GeoM.Translate(carWidth*uiScale/-2, carWidth*uiScale/-2)
			opts.GeoM.Translate(b.x*uiScale, b.y*uiScale*-1)
			screen.DrawImage(blockImage, &opts)
		}

		for _, v := range vehicules {
			v.Drive(drive, 1.0/60)
		}

		collisions := make(map[int]struct{})

		for i1, v1 := range vehicules {
			for _, b := range blocks {
				//println("check", v1.Position.x)
				if v1.Collide(b, vehiculeVehiculeRadiusSqrt) {
					collisions[i1] = struct{}{}
					v1.Velocity = 0
				}
			}
			for i2, v2 := range vehicules[i1+1:] {
				if v1.Collide(&v2.Position, vehiculeVehiculeRadiusSqrt) {
					collisions[i1] = struct{}{}
					collisions[i2+i1+1] = struct{}{}
					v1.Velocity = 0
					v2.Velocity = 0
				}
			}
		}

		for i, v := range vehicules {
			go func(v *Vehicule, i int) {
				opts := ebiten.DrawImageOptions{}
				opts.GeoM.Translate(carWidth*uiScale/-2, carHeight*uiScale/-2)
				opts.GeoM.Rotate(v.Rotation * -1)
				opts.GeoM.Translate(v.x*uiScale, v.y*uiScale*-1)
				if _, collision := collisions[i]; collision {
					opts.ColorM.Translate(1, -1, -1, 0)
				}
				optsChan <- opts
				wg.Done()
			}(v, i)
		}

		go func() {
			for opts := range optsChan {
				screen.DrawImage(vehiculeImage, &opts)
			}
		}()

		wg.Wait()
		close(optsChan)

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			x, y = x/uiScale, y/-uiScale
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("click at %d:%d", x, y), 0, groundHeight*uiScale-20)
			blocks = append(blocks, &Position{float64(x), float64(y)})
		}

		ebitenutil.DebugPrint(screen, fmt.Sprintf(
			"pos: %.1f:%.1f\n%#v\n%.1f km/h\n%.2f°\nfps:%.0f\nvcount: %d",
			vehicules[0].x, vehicules[0].y, drive, vehicules[0].Velocity/1000*60*60, vehicules[0].Rotation*180/math.Pi, ebiten.CurrentFPS(), len(vehicules)))
		return nil
	}

	if err := ebiten.Run(update, groundWidth*uiScale, groundHeight*uiScale, 2, "Drive my crazy!"); err != nil {
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
