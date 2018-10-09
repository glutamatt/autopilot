package graphics

import (
	"image/color"
	"math"

	"github.com/glutamatt/autopilot/model"
	"github.com/hajimehoshi/ebiten"
)

var turnInc = .02
var CarWidth float64
var CarHeight float64
var UiScale int
var BlockBorder = 5

func SetCarDimension(w, h float64) {
	CarWidth, CarHeight = w, h
}

//SetTurnInc to turn inc
func SetTurnInc(f float64) {
	turnInc = f
}

//InputControls to driving
func InputControls(drive *model.Driving) {
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

func VehiculeImageOptions(v *model.Vehicule, collision bool) (opts ebiten.DrawImageOptions) {
	opts.GeoM.Translate(CarWidth*float64(UiScale)/-2.0, CarHeight*float64(UiScale)/-2.0)
	opts.GeoM.Rotate(v.Rotation * -1)
	opts.GeoM.Translate(v.X*float64(UiScale), v.Y*float64(UiScale)*-1)
	if collision {
		opts.ColorM.Translate(1, -1, -1, 0)
	}
	return
}

var blockImage *ebiten.Image

func InitBlockImage() {
	blockImage, _ = ebiten.NewImage(int(BlockBorder*UiScale), int(BlockBorder*UiScale), ebiten.FilterNearest)
	blockImage.Fill(color.NRGBA{0xBB, 0xBB, 0xBB, 0xff})
}

func HandleBlockAdd(blocksImage *ebiten.Image) *model.Position {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		x, y = x/UiScale, y/-UiScale
		pos := &model.Position{X: float64(x), Y: float64(y)}
		pos.Gap(BlockBorder)
		opts := ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(BlockBorder*UiScale)/-2, float64(BlockBorder*UiScale)/-2)
		opts.GeoM.Translate(pos.X*float64(UiScale), pos.Y*float64(UiScale)*-1)
		blocksImage.DrawImage(blockImage, &opts)
		return pos
	}

	return nil
}

func DrawPath(img *ebiten.Image, positions ...model.Position) {
	for _, pos := range positions {
		opts := ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(BlockBorder*UiScale)/-2, float64(BlockBorder*UiScale)/-2)
		opts.GeoM.Translate(pos.X*float64(UiScale), pos.Y*float64(UiScale)*-1)
		opts.ColorM.Scale(1.5, 1.5, 1.5, .1)
		img.DrawImage(blockImage, &opts)
	}
}
