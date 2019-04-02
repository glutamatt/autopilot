package graphics

import (
	"image/color"

	"github.com/glutamatt/autopilot/model"

	"github.com/hajimehoshi/ebiten"
)

var visualizationSize int
var targetImg *ebiten.Image
var wallImg *ebiten.Image
var carImg *ebiten.Image
var rotateImg *ebiten.Image
var sightDistance float64
var rearDistance float64
var targetRatioSize = 20
var wallRatioSize = 30
var carRatioSize = 25
var visuRatio float64
var indicesPerRow int

func InitConstants(visualSize int, sightSize, rearSize float64, mapSize int) {
	visualizationSize = visualSize
	sightDistance = sightSize
	rearDistance = rearSize
	visuRatio = float64(visualizationSize) / sightDistance
	indicesPerRow = mapSize

	targetImg, _ = ebiten.NewImage(visualizationSize/targetRatioSize, visualizationSize/targetRatioSize, ebiten.FilterDefault)
	targetImg.Fill(color.NRGBA{0xFF, 0xBB, 0xBB, 0xff})
	wallImg, _ = ebiten.NewImage(visualizationSize/wallRatioSize, visualizationSize/wallRatioSize, ebiten.FilterDefault)
	wallImg.Fill(color.NRGBA{0xCC, 0xCC, 0xCC, 0xff})
	carImg, _ = ebiten.NewImage(visualizationSize/carRatioSize, visualizationSize/carRatioSize, ebiten.FilterDefault)
	carImg.Fill(color.NRGBA{0x22, 0x22, 0xEE, 0xff})
	rotateImg, _ = ebiten.NewImage(visualizationSize/carRatioSize*3, 1, ebiten.FilterDefault)
	rotateImg.Fill(color.Black)
}

func DrawExport(export []float64) (*ebiten.Image, error) {
	img, err := ebiten.NewImage(visualizationSize, visualizationSize, ebiten.FilterDefault)
	img.Fill(color.NRGBA{0x00, 0x88, 0x00, 0xff})

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(visualizationSize/targetRatioSize)/-2, float64(visualizationSize/targetRatioSize)/-2)
	opts.GeoM.Translate(rearDistance*visuRatio, sightDistance/2*visuRatio)
	opts.ColorM.Apply(color.Black)
	img.DrawImage(carImg, opts)

	opts = &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(visualizationSize/targetRatioSize)/-2, float64(visualizationSize/targetRatioSize)/-2)
	opts.GeoM.Translate((export[1]*sightDistance+rearDistance)*visuRatio, export[2]*sightDistance*visuRatio*-1+sightDistance/2*visuRatio)
	img.DrawImage(targetImg, opts)

	wallOffset := 3 + 6*indicesPerRow*indicesPerRow
	// 6 values per car for each point
	//velocity, target X and Y

	walls := export[wallOffset : wallOffset+indicesPerRow*indicesPerRow]
	for i, w := range walls {
		if w > .5 {
			p := iToPos(i)
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(float64(visualizationSize/wallRatioSize)/-2, float64(visualizationSize/wallRatioSize)/-2)
			opts.GeoM.Translate(p.X*visuRatio, p.Y*visuRatio)
			img.DrawImage(wallImg, opts)
		}
	}

	for i := 0; i < indicesPerRow*indicesPerRow; i++ {
		if export[3+i*6+2] > 0 {
			p := iToPos(i)
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(float64(visualizationSize/carRatioSize)/-2, float64(visualizationSize/carRatioSize)/-2)
			opts.GeoM.Rotate(export[3+i*6+3] * -1)
			opts.GeoM.Translate(p.X*visuRatio, p.Y*visuRatio)
			img.DrawImage(rotateImg, opts)
			img.DrawImage(carImg, opts)
		}
	}
	//println(not)

	return img, err
}

func iToPos(i int) model.Position {
	return model.Position{
		X: (float64(i%indicesPerRow) + .5) * sightDistance / float64(indicesPerRow),
		Y: (float64(i/indicesPerRow) + .5) * sightDistance / float64(indicesPerRow),
	}
}
