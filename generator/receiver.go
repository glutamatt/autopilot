package generator

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten"

	"github.com/glutamatt/autopilot/generator/graphics"

	"github.com/glutamatt/autopilot/model"
)

var gameBlocks map[model.Position]bool
var chanFrame chan struct{}
var chanVehicule chan vehiculeState
var sightDistance = 100.0
var metersPerIndex = 5.0
var indicesPerRow = int(sightDistance/metersPerIndex) + 2
var debugVisu chan *ebiten.Image

type outputLine struct {
	current vehiculeState
	others  map[model.Position]vehiculeState
	blocks  map[model.Position]bool
}

func posToIndices(p model.Position) (int, int) {
	p.Y -= sightDistance / 2
	p.Y *= -1
	x := int(p.X / metersPerIndex)
	if math.Mod(p.X, metersPerIndex) > metersPerIndex/2 {
		x++
	}
	y := int(p.Y / metersPerIndex)
	if math.Mod(p.Y, metersPerIndex) > metersPerIndex/2 {
		y++
	}
	return x, y
}

func (o outputLine) Floats() []float64 {
	line := []float64{}
	line = append(line, o.current.vehicule.Velocity)
	line = append(line, o.current.target.X/sightDistance)
	line = append(line, o.current.target.Y/sightDistance)

	//add others
	others := make([][]*vehiculeState, indicesPerRow)
	for i := range others {
		others[i] = make([]*vehiculeState, indicesPerRow)
	}
	for p, v := range o.others {
		x, y := posToIndices(p)
		others[y][x] = &v
	}
	for _, r := range others {
		for _, v := range r {
			if v == nil {
				line = append(line, 0)
				line = append(line, 0)
				line = append(line, 0)
				line = append(line, 0)
				line = append(line, 0)
				line = append(line, 0)
			} else {
				line = append(line, v.drive.Turning)
				line = append(line, v.drive.Thrust)
				line = append(line, v.vehicule.Velocity)
				line = append(line, v.vehicule.Rotation)
				line = append(line, v.target.X/sightDistance)
				line = append(line, v.target.Y/sightDistance)
			}
		}
	}
	blocks := make([][]bool, indicesPerRow)
	for i := range blocks {
		blocks[i] = make([]bool, indicesPerRow)
	}
	for p := range o.blocks {
		x, y := posToIndices(p)
		blocks[y][x] = true
	}
	for _, r := range blocks {
		for _, b := range r {
			if b {
				line = append(line, 1)
			} else {
				line = append(line, 0)
			}
		}
	}

	line = append(line, o.current.drive.Turning)
	line = append(line, o.current.drive.Thrust)

	return line
}

type vehiculeState struct {
	index    int
	vehicule model.Vehicule
	target   model.Position
	drive    model.Driving
}

func Init(blocks map[model.Position]bool) chan *ebiten.Image {
	gameBlocks = blocks
	chanFrame = make(chan struct{})
	chanVehicule = make(chan vehiculeState)
	graphics.InitConstants(100, sightDistance, indicesPerRow)
	debugVisu = make(chan *ebiten.Image)
	dataFilePath := os.TempDir() + string(os.PathSeparator) + time.Now().Format("autopilot_2006_01_02__15_04_05.csv")
	dataFile, err := os.Create(dataFilePath)
	if err != nil {
		panic(fmt.Errorf("Unable to create the data file %s : %v", dataFilePath, err))
	}
	csvWriter := csv.NewWriter(dataFile)
	flushTimer := time.Tick(3 * time.Second)
	saveTimer := time.Tick(200 * time.Millisecond)
	go func() {
		vehicules := []vehiculeState{}
		for {
			select {
			case <-flushTimer:
				csvWriter.Flush()
			case <-chanFrame:
				if len(vehicules) > 0 {
					select {
					case <-saveTimer:
						saveVehicules(vehicules, csvWriter)
					default:
					}
					for i, v := range vehicules {
						if v.index == 0 {
							if img, err := graphics.DrawExport(processVehicule(i, vehicules).Floats()); err == nil {
								go func(img *ebiten.Image) {
									select {
									case debugVisu <- img:
									case <-time.After(time.Second / 30):
									}
								}(img)
							}
							break
						}
					}

					vehicules = []vehiculeState{}
				}
			case v := <-chanVehicule:
				vehicules = append(vehicules, v)
			}
		}
	}()

	return debugVisu
}

func processVehicule(vehiculeIndex int, vehicules []vehiculeState) outputLine {
	vC := vehicules[vehiculeIndex]
	output := outputLine{current: vC}
	cos := math.Cos(-vC.vehicule.Rotation)
	sin := math.Sin(-vC.vehicule.Rotation)
	output.current.target.X -= vC.vehicule.X
	output.current.target.Y -= vC.vehicule.Y
	output.current.target.X, output.current.target.Y = output.current.target.X*cos-output.current.target.Y*sin, output.current.target.X*sin+output.current.target.Y*cos
	output.blocks = map[model.Position]bool{}
	for p := range filterCloseBlocks(vC.vehicule.Position, gameBlocks) {
		p.X -= vC.vehicule.X
		p.Y -= vC.vehicule.Y
		p := model.Position{
			X: p.X*cos - p.Y*sin,
			Y: p.X*sin + p.Y*cos,
		}
		if isInOutputRange(p) {
			output.blocks[p] = true
		}
	}
	output.others = make(map[model.Position]vehiculeState, len(vehicules)-1)
	for _, vO := range filterCloseVehicules(vC.vehicule.Position, vehiculeIndex, vehicules) {
		vO.target.X -= vC.vehicule.X
		vO.target.Y -= vC.vehicule.Y
		vO.vehicule.X -= vC.vehicule.X
		vO.vehicule.Y -= vC.vehicule.Y
		vO.vehicule.Rotation -= vC.vehicule.Rotation
		vO.vehicule.X, vO.vehicule.Y = vO.vehicule.X*cos-vO.vehicule.Y*sin, vO.vehicule.X*sin+vO.vehicule.Y*cos
		vO.target.X, vO.target.Y = vO.target.X*cos-vO.target.Y*sin, vO.target.X*sin+vO.target.Y*cos
		if isInOutputRange(vO.vehicule.Position) {
			output.others[vO.vehicule.Position] = vO
		}
	}

	return output
}

func saveVehicules(vehicules []vehiculeState, csvWriter *csv.Writer) {
	if len(vehicules) == 0 {
		return
	}
	for iC := range vehicules {
		floats := processVehicule(iC, vehicules).Floats()
		str := make([]string, len(floats))
		for i, f := range floats {
			str[i] = fmt.Sprint(f)
		}
		csvWriter.Write(str)
	}
}

func isInWideRange(from, to model.Position) bool {
	return math.Abs(from.X-to.X) <= sightDistance && math.Abs(from.Y-to.Y) <= sightDistance
}

func isInOutputRange(p model.Position) bool {
	return p.X > 0 && p.X <= sightDistance &&
		math.Abs(p.Y) <= sightDistance/2
}

func filterCloseVehicules(base model.Position, baseIndex int, vehicules []vehiculeState) []vehiculeState {
	m := []vehiculeState{}
	for i, p := range vehicules {
		if baseIndex != i && isInWideRange(p.vehicule.Position, base) {
			m = append(m, p)
		}
	}
	return m
}

func filterCloseBlocks(base model.Position, b map[model.Position]bool) map[model.Position]bool {
	m := map[model.Position]bool{}
	for p := range b {
		if isInWideRange(p, base) {
			m[p] = true
		}
	}
	return m
}

func NewFrame() {
	chanFrame <- struct{}{}
}

func AddVehicule(iv int, vehicule *model.Vehicule, target model.Position, drive *model.Driving) {
	v := *vehicule
	chanVehicule <- vehiculeState{
		index:    iv,
		vehicule: v,
		target:   target,
		drive:    *drive,
	}
}
