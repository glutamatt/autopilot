package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/glutamatt/autopilot/generator"
	"github.com/glutamatt/autopilot/model"

	"github.com/glutamatt/autopilot/ia"

	"github.com/glutamatt/autopilot/graphics"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const minTurningRadius = 10
const turnWheelInc = .02
const carWidth = 5
const blockBorder = 5
const carHeight = 2
const uiScale = 2
const groundWidth = 400
const groundHeight = 250
const adherenceMax = 2.5    // m/s/s newton force
const boostMax = 3          //m/s/s
const breakMax = 11         //m/s/s
const reverseMaxSpeed = 5.6 // m/S
const securityDistance = 1.0

var spawneFreq = 10 * time.Second

var deepLearning = true

func createVehiculeManager(from, to model.Position) *vehiculeManager {
	return &vehiculeManager{
		pathTicker: time.NewTicker(500 * time.Millisecond),
		iaTicker:   time.NewTicker(50 * time.Millisecond),
		target:     to,
		vehicule: &model.Vehicule{
			Position: from,
			Rotation: from.Angle(to),
		},
	}
}

func main() {
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		log.Println("pprof.StartCPUProfile")
		go func() {
			time.Sleep(20 * time.Second)
			pprof.StopCPUProfile()
			log.Println("Profiuling Done")
		}()
	}

	vehicules := []*vehiculeManager{}
	model.SetMinTurningRadius(minTurningRadius)
	model.SetAdherenceMax(adherenceMax)
	ia.PrepareDrives()
	model.BoostMax = boostMax
	model.SecurityDistance = securityDistance
	model.BreakMax = breakMax
	model.ReverseMaxSpeed = reverseMaxSpeed
	graphics.SetTurnInc(turnWheelInc)
	graphics.SetCarDimension(carWidth, carHeight)
	graphics.UiScale = uiScale
	graphics.BlockBorder = blockBorder
	graphics.PepareWheel()
	model.InitPathTiles(blockBorder, groundWidth, groundHeight)
	graphics.InitBlockImage()
	ia.PrepareDrives()
	ia.BlocRadius = model.InitRadiusBlock(blockBorder)
	ia.VehiculRadius = model.InitRadiusCar(carWidth, carHeight)
	var debugVisuImg *ebiten.Image

	blocks := make(map[model.Position]bool)
	spots := map[model.Position]int{}
	manualDrive := &model.Driving{}
	boostVisu := graphics.InitBoostVisu()

	vehiculeImage, _ := ebiten.NewImage(int(carWidth*uiScale), int(carHeight*uiScale), ebiten.FilterNearest)
	blocksImage, _ := ebiten.NewImage(groundWidth*uiScale, groundHeight*uiScale, ebiten.FilterNearest)

	for _, p := range model.GenerateBlocks(groundWidth, groundHeight) {
		blocks[*p] = true
		graphics.DrawBlock(p, blocksImage)
	}

	debugExport := generator.Init(blocks)

	vehiculeImage.Fill(color.NRGBA{0xFF, 0xFF, 0xFF, 0xff})

	spawner := time.NewTimer(7 * time.Second)

	update := func(screen *ebiten.Image) error {
		generator.NewFrame()
		select {
		case <-spawner.C:
			if len(spots) > 1 && len(vehicules) < 20 {
				from, to := spotPositions(spots, false)
				if freePlace(vehicules, from) {
					vehicules = append(vehicules, createVehiculeManager(from, to))
					spawner = time.NewTimer(spawneFreq / time.Duration(len(spots)/2))
				} else {
					spawner = time.NewTimer(time.Second)
				}
			} else {
				spawner = time.NewTimer(time.Second)
			}
		default:
		}

		screen.Fill(color.NRGBA{0x00, 0x00, 0x88, 0xff})

		wg := sync.WaitGroup{}
		wg.Add(len(vehicules))
		optsChan := make(chan ebiten.DrawImageOptions)
		screen.DrawImage(blocksImage, nil)

		manualDriveOn := graphics.InputControls(manualDrive)

		arrivedChan := make(chan int)

		//blocksAndCars := getBlocksAndCars(blocks, vehicules, func(v *model.Vehicule) bool { return true })
		//blocksAndSlowCars :=

		var allVehiculesfeatures generator.FeaturesByIndex

		if deepLearning {
			for iv, v := range vehicules {
				generator.AddVehicule(iv, v.vehicule, v.Target(), nil)
			}
			allVehiculesfeatures = generator.GetVehiculeFeatures()
		}

		for iv, v := range vehicules {
			if iv == 0 && manualDriveOn {
				v.futurePositions = ia.Extrapol(v.vehicule, manualDrive)
				v.vehicule.Drive(manualDrive, 1.0/60)
				wg.Done()
				continue
			}

			go func(v *vehiculeManager, iv int) {
				defer wg.Done()
				select {
				case <-v.iaTicker.C:
					if !deepLearning && len(v.pathFound) > 0 {
						drives, future := ia.Genetic(
							v.vehicule,
							v.futureDrives,
							v.Target(),
							getBlocksAndCars(blocks, vehicules, func(i int, v *model.Vehicule) bool { return i != iv }),
							futureBlockedPos(iv, vehicules),
						)
						v.futurePositions = future
						v.futureDrives = drives
					}
					if deepLearning && len(v.pathFound) > 0 {
						if features, have := allVehiculesfeatures[iv]; have {
							v.futureDrives = []*model.Driving{ia.NeuralNet(features)}
						}
					}
				case <-v.pathTicker.C:
					if found, path := model.FindPath(
						v.vehicule.Position,
						v.target,
						getBlocksAndCars(blocks, vehicules, func(i int, v *model.Vehicule) bool { return i != iv && math.Abs(v.Velocity) < 3 }),
					); found {
						if len(path) < 3 {
							arrivedChan <- iv
							return
						}
						if len(path) > 8 {
							path = path[len(path)-8:]
						}
						v.pathFound = path
					} else {
						arrivedChan <- iv
						return
					}
				default:
				}
				if v.futureDrives != nil {
					v.vehicule.Drive(v.futureDrives[0], 1.0/60)
				} else {
					v.vehicule.Drive(&model.Driving{}, 1.0/60)
				}
				if !deepLearning && v.futureDrives != nil {
					generator.AddVehicule(iv, v.vehicule, v.Target(), v.futureDrives[0])
				}
			}(v, iv)
		}

		arrived := map[int]bool{}
		go func() {
			for i := range arrivedChan {
				arrived[i] = true
			}
		}()

		wg.Wait()
		close(arrivedChan)

		remainingV := []*vehiculeManager{}
		for i, iv := range vehicules {
			if !arrived[i] {
				if i == 0 && iv.futureDrives != nil && len(iv.futureDrives) > 0 {
					graphics.SetWheelRotation(iv.futureDrives[0].Turning, screen)
					boostVisu.Render(iv.futureDrives[0].Thrust, screen)
				}
				remainingV = append(remainingV, iv)
				graphics.DrawPath(screen, 0, iv.futurePositions...) // DEBUG print future positions
				graphics.DrawPath(screen, 1, iv.Target())           // DEBUG print future positions
			} else {
				iv.pathTicker.Stop()
				iv.iaTicker.Stop()
			}
		}
		vehicules = remainingV

		//collisions := model.Collisions(vehicules, blocks)

		wg.Add(len(vehicules))
		for i, v := range vehicules {
			go func(v *model.Vehicule, i int) {
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
				if _, exist := spots[*pos]; !exist {
					spots[*pos] = len(spots)
					if len(spots) > 1 && len(spots)%2 == 0 {
						from, to := spotPositions(spots, true)
						vehicules = append(vehicules, createVehiculeManager(from, to))
					}
				}
			}
		}

		select {
		case debugVisuImg = <-debugExport:
		default:
		}
		if debugVisuImg != nil {
			debugOpt := &ebiten.DrawImageOptions{}
			debugOpt.GeoM.Translate(0, 190)
			screen.DrawImage(debugVisuImg, debugOpt)
		}

		if len(vehicules) > 0 {
			ebitenutil.DebugPrint(screen, fmt.Sprintf(
				"pos: %.1f:%.1f\n%.1f km/h\n%.2f°\nfps:%.0f\nvcount: %d",
				vehicules[0].vehicule.X, vehicules[0].vehicule.Y, vehicules[0].vehicule.Velocity/1000*60*60, vehicules[0].vehicule.Rotation*180/math.Pi, ebiten.CurrentFPS(), len(vehicules)))
		}

		return nil
	}

	if err := ebiten.Run(update, groundWidth*uiScale, groundHeight*uiScale, 1.5, "Drive my crazy!"); err != nil {
		panic(err)
	}
}

type vehiculeManager struct {
	vehicule        *model.Vehicule
	pathTicker      *time.Ticker
	iaTicker        *time.Ticker
	target          model.Position
	pathFound       []model.Position
	futurePositions []model.Position
	futureDrives    []*model.Driving
}

func (v *vehiculeManager) Target() model.Position {
	if v.pathFound == nil || len(v.pathFound) == 0 {
		return model.Position{}
	}
	vel := math.Max(v.vehicule.Velocity, 0)
	i := int(vel*6/13.8 + 3)
	if i > len(v.pathFound) {
		i = len(v.pathFound)
	}
	pathI := len(v.pathFound) - i
	return v.pathFound[pathI]
}

func spotPositions(spots map[model.Position]int, last bool) (model.Position, model.Position) {
	p := make([]model.Position, len(spots))
	for q, i := range spots {
		p[i] = q
	}
	if last {
		return p[len(p)-2], p[len(p)-1]
	}
	k := rand.Intn(len(spots)/2) * 2
	return p[k], p[k+1]
}

func futureBlockedPos(vehiculeKey int, vehicules []*vehiculeManager) []map[model.Position]bool {
	posLen := len(vehicules[0].futurePositions)
	ret := make([]map[model.Position]bool, posLen)
	for i := 0; i < posLen; i++ {
		ret[i] = make(map[model.Position]bool)
		for iv, v := range vehicules {
			if iv != vehiculeKey && len(v.futurePositions) >= i+1 {
				ret[i][v.futurePositions[i]] = true
			}
		}
	}
	return ret
}

func getBlocksAndCars(blocks map[model.Position]bool, vehicules []*vehiculeManager, carFilter func(int, *model.Vehicule) bool) *map[model.Position]bool {
	newBlocks := make(map[model.Position]bool, len(blocks))
	for p, v := range blocks {
		newBlocks[p] = v
	}
	for i, v := range vehicules {
		if carFilter(i, v.vehicule) {
			p := v.vehicule.Position
			p.Gap(blockBorder)
			newBlocks[p] = true
			//newBlocks = model.BlocksArround(newBlocks, v.vehicule.Position, blockBorder, blockBorder, 2.0)
		}
	}
	return &newBlocks
}

func freePlace(vehicules []*vehiculeManager, place model.Position) bool {
	for _, v := range vehicules {
		if v.vehicule.Position.ManDist(place) < carWidth {
			return false
			break
		}
	}
	return true
}

/*
func blocksButPos(blocks map[model.Position]bool, p model.Position) *map[model.Position]bool {
	newBlocks := make(map[model.Position]bool, len(blocks))
	p.Gap(blockBorder)
	for pos, v := range blocks {
		if pos != p {
			newBlocks[pos] = v
		}
	}
	return &newBlocks
}
*/
