package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"sync"

	"github.com/glutamatt/autopilot/graphics"
	geom "github.com/glutamatt/autopilot/model"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const minTurningRadius = 12
const turnInc = .02
const carWidth = 5
const blockBorder = 10
const carHeight = 2
const uiScale = 2
const groundWidth = 600
const groundHeight = 300

func createRandomVehicule() *geom.Vehicule {
	return &geom.Vehicule{
		Position: geom.Position{
			X: rand.Float64()*groundWidth/2 + groundWidth/4,
			Y: rand.Float64()*groundHeight/-2 - groundHeight/4,
		},
		Rotation: rand.Float64() * math.Pi * 2,
	}
}

func main() {

	vehicules := make([]*geom.Vehicule, 300)

	geom.SetMinTurningRadius(minTurningRadius)
	graphics.SetTurnInc(turnInc)
	graphics.SetCarDimension(carWidth, carHeight)
	graphics.UiScale = uiScale
	graphics.BlockBorder = blockBorder
	graphics.InitBlockImage()
	geom.InitBlockCar(blockBorder)
	geom.InitRadiusCar(carWidth, carHeight)

	blocks := make(map[geom.Position]bool)
	vehiculeImage, _ := ebiten.NewImage(int(carWidth*uiScale), int(carHeight*uiScale), ebiten.FilterNearest)
	blocksImage, _ := ebiten.NewImage(groundWidth*uiScale, groundHeight*uiScale, ebiten.FilterNearest)
	vehiculeImage.Fill(color.NRGBA{0xFF, 0xFF, 0xFF, 0xff})
	for i := range vehicules {
		vehicules[i] = createRandomVehicule()
	}

	drive := &geom.Driving{}

	update := func(screen *ebiten.Image) error {
		graphics.InputControls(drive)
		screen.Fill(color.NRGBA{0x00, 0x00, 0x88, 0xff})

		wg := sync.WaitGroup{}
		wg.Add(len(vehicules))
		optsChan := make(chan ebiten.DrawImageOptions)
		screen.DrawImage(blocksImage, nil)

		for _, v := range vehicules {
			v.Drive(drive, 1.0/60)
		}

		collisions := geom.Collisions(vehicules, blocks)

		for i, v := range vehicules {
			go func(v *geom.Vehicule, i int) {
				_, collision := collisions[i]
				optsChan <- graphics.VehiculeImage(v, collision)
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
			if block := graphics.HandleBlockAdd(blocksImage); block != nil {
				blocks[*block] = true
			}
		}

		ebitenutil.DebugPrint(screen, fmt.Sprintf(
			"pos: %.1f:%.1f\n%#v\n%.1f km/h\n%.2f°\nfps:%.0f\nvcount: %d",
			vehicules[0].X, vehicules[0].Y, drive, vehicules[0].Velocity/1000*60*60, vehicules[0].Rotation*180/math.Pi, ebiten.CurrentFPS(), len(vehicules)))
		return nil
	}

	if err := ebiten.Run(update, groundWidth*uiScale, groundHeight*uiScale, 2, "Drive my crazy!"); err != nil {
		panic(err)
	}
}
