package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/glutamatt/autopilot/ia"

	"github.com/glutamatt/autopilot/graphics"
	geom "github.com/glutamatt/autopilot/model"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const minTurningRadius = 11
const turnWheelInc = .02
const carWidth = 5
const blockBorder = 3
const carHeight = 2
const uiScale = 2
const groundWidth = 500
const groundHeight = 300
const adherenceMax = 2.5 // m/s/s newton force

func createVehiculeManager(spots []geom.Position) *vehiculeManager {
	spotKey := rand.Intn(len(spots))
	targetKey := 0
	for {
		targetKey = rand.Intn(len(spots))
		if targetKey != spotKey {
			break
		}
	}
	return &vehiculeManager{
		pathTicker: time.Tick(300 * time.Millisecond),
		target:     spots[targetKey],
		vehicule: &geom.Vehicule{
			Position: spots[spotKey],
			Rotation: math.Atan2(spots[targetKey].Y-spots[spotKey].Y, spots[targetKey].X-spots[spotKey].X),
		},
	}
}

func main() {

	vehicules := []*vehiculeManager{}
	geom.SetMinTurningRadius(minTurningRadius)
	geom.SetAdherenceMax(adherenceMax)
	graphics.SetTurnInc(turnWheelInc)
	graphics.SetCarDimension(carWidth, carHeight)
	graphics.UiScale = uiScale
	graphics.BlockBorder = blockBorder
	graphics.PepareWheel()
	geom.InitPathTiles(blockBorder, groundWidth, groundHeight)
	graphics.InitBlockImage()
	ia.BlocRadius = geom.InitRadiusBlock(blockBorder)
	ia.VehiculRadius = geom.InitRadiusCar(carWidth, carHeight)

	blocks := make(map[geom.Position]bool)
	spots := map[geom.Position]bool{}

	vehiculeImage, _ := ebiten.NewImage(int(carWidth*uiScale), int(carHeight*uiScale), ebiten.FilterNearest)
	blocksImage, _ := ebiten.NewImage(groundWidth*uiScale, groundHeight*uiScale, ebiten.FilterNearest)

	for _, p := range geom.GenerateBlocks(groundWidth, groundHeight) {
		blocks[*p] = true
		graphics.DrawBlock(p, blocksImage)
	}
	vehiculeImage.Fill(color.NRGBA{0xFF, 0xFF, 0xFF, 0xff})

	spawner := time.NewTicker(4 * time.Second)

	update := func(screen *ebiten.Image) error {

		select {
		case <-spawner.C:
			if len(spots) > 1 {
				vehicules = append(vehicules, createVehiculeManager(spotPositios(spots)))
			}
		default:
		}

		screen.Fill(color.NRGBA{0x00, 0x00, 0x88, 0xff})

		wg := sync.WaitGroup{}
		wg.Add(len(vehicules))
		optsChan := make(chan ebiten.DrawImageOptions)
		screen.DrawImage(blocksImage, nil)

		//DEBUG
		/*
			draw := []geom.Position{}
			for _, v := range vehicules {
				draw = append(draw, v.futurePositions...)
			}
			graphics.DrawPath(screen, draw...)
		*/

		manualDrive := &geom.Driving{}
		manualDriveOn := graphics.InputControls(manualDrive)
		fmt.Printf("%#v\n", manualDrive)

		arrived := map[int]bool{}
		for iv, v := range vehicules {
			if iv == 0 && manualDriveOn {
				v.futurePositions = ia.Extrapol(v.vehicule, manualDrive)
				v.vehicule.Drive(manualDrive, 1.0/60)
				wg.Done()
				continue
			}
			go func(v *vehiculeManager, iv int) {
				defer wg.Done()
				blocked := blockedPos(iv, vehicules, blocks)
				select {
				case <-v.pathTicker:
					if found, path := geom.FindPath(v.vehicule.Position, v.target, blocked); found {
						if found && len(path) < 3 {
							arrived[iv] = true
							return
						}
						if len(path) > 10 {
							path = path[len(path)-10:]
						}
						v.pathFound = path
					}
				default:
				}
				if len(v.pathFound) > 0 {
					drive, future := ia.Genetic(v.vehicule, &v.pathFound, blocked)
					v.vehicule.Drive(drive, 1.0/60)
					v.futurePositions = future
				}
			}(v, iv)
		}
		wg.Wait()

		remainingV := []*vehiculeManager{}
		for i, iv := range vehicules {
			if !arrived[i] {
				remainingV = append(remainingV, iv)
			}
		}
		vehicules = remainingV

		//collisions := geom.Collisions(vehicules, blocks)

		wg.Add(len(vehicules))
		for i, v := range vehicules {
			go func(v *geom.Vehicule, i int) {
				//_, collision := collisions[i]
				optsChan <- graphics.VehiculeImageOptions(v, false)
				wg.Done()
			}(v.vehicule, i)
		}

		go func() {
			for opts := range optsChan {
				screen.DrawImage(vehiculeImage, &opts)
			}
		}()

		wg.Wait()
		close(optsChan)

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			if pos := graphics.GetMouseClickPos(); pos != nil {
				spots[*pos] = true
			}
		}

		if len(vehicules) > 0 {
			ebitenutil.DebugPrint(screen, fmt.Sprintf(
				"pos: %.1f:%.1f\n%.1f km/h\n%.2f°\nfps:%.0f\nvcount: %d",
				vehicules[0].vehicule.X, vehicules[0].vehicule.Y, vehicules[0].vehicule.Velocity/1000*60*60, vehicules[0].vehicule.Rotation*180/math.Pi, ebiten.CurrentFPS(), len(vehicules)))
		}

		return nil
	}

	if err := ebiten.Run(update, groundWidth*uiScale, groundHeight*uiScale, 2, "Drive my crazy!"); err != nil {
		panic(err)
	}
}

type vehiculeManager struct {
	vehicule        *geom.Vehicule
	pathTicker      <-chan time.Time
	target          geom.Position
	pathFound       []geom.Position
	futurePositions []geom.Position
}

func spotPositios(spots map[geom.Position]bool) []geom.Position {
	p := make([]geom.Position, len(spots))
	i := 0
	for q := range spots {
		p[i] = q
		i++
	}
	return p
}

func blockedPos(vehiculeKey int, vehicules []*vehiculeManager, blocks map[geom.Position]bool) *map[geom.Position]bool {
	blocked := map[geom.Position]bool{}

	for i, v := range vehicules {
		if i != vehiculeKey {
			for _, p := range v.futurePositions {
				blocked[p] = true
			}
		}
	}

	for p, b := range blocks {
		blocked[p] = b
	}
	return &blocked
}
