package graphics

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"math"

	"github.com/glutamatt/autopilot/model"
	"github.com/hajimehoshi/ebiten"
)

var turnInc = .02
var CarWidth float64
var CarHeight float64
var UiScale int
var BlockBorder int

func SetCarDimension(w, h float64) {
	CarWidth, CarHeight = w, h
}

//SetTurnInc to turn inc
func SetTurnInc(f float64) {
	turnInc = f
}

//InputControls to driving
func InputControls(drive *model.Driving) (keyPressed bool) {
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		keyPressed = true
		drive.Thrust = -1
	} else {
		if ebiten.IsKeyPressed(ebiten.KeyUp) {
			keyPressed = true
			drive.Thrust = 1
		} else {
			drive.Thrust = 0
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		keyPressed = true
		if drive.Turning < 1 {
			if drive.Turning < 0 {
				drive.Turning = 0
			}
			drive.Turning += turnInc
		}
	} else {
		if ebiten.IsKeyPressed(ebiten.KeyRight) {
			keyPressed = true
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

	return keyPressed
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

func GetMouseClickPos() *model.Position {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		x, y = x/UiScale, y/-UiScale
		pos := &model.Position{X: float64(x), Y: float64(y)}
		pos.Gap(BlockBorder)
		return pos
	}

	return nil
}

func DrawBlock(pos *model.Position, blocksImage *ebiten.Image) {
	opts := ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(BlockBorder*UiScale)/-2, float64(BlockBorder*UiScale)/-2)
	opts.GeoM.Translate(pos.X*float64(UiScale), pos.Y*float64(UiScale)*-1)
	blocksImage.DrawImage(blockImage, &opts)
}

func DrawPath(img *ebiten.Image, color float64, positions ...model.Position) {
	for _, pos := range positions {
		opts := ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(BlockBorder*UiScale)/-2, float64(BlockBorder*UiScale)/-2)
		opts.GeoM.Translate(pos.X*float64(UiScale), pos.Y*float64(UiScale)*-1)
		opts.ColorM.Scale(1.5, 1.5, 1.5, .1)
		opts.ColorM.RotateHue(color)
		img.DrawImage(blockImage, &opts)
	}
}

var wheel *ebiten.Image

func PepareWheel() {
	wheekBytes, err := ioutil.ReadFile("wheel.png")
	if err != nil {
		panic(fmt.Errorf("Unable to Read file wheel.png : %v", err))
	}
	img, _, err := image.Decode(bytes.NewReader(wheekBytes))
	if err != nil {
		panic(fmt.Errorf("Unable to decode image wheek bytes : %v", err))
	}
	wheel, err = ebiten.NewImageFromImage(img, ebiten.FilterDefault)
	if err != nil {
		panic(fmt.Errorf("Unable to create wheel image from image : %v", err))
	}
}

func SetWheelRotation(turning float64, screen *ebiten.Image) {
	wheelOpt := &ebiten.DrawImageOptions{}
	wheelOpt.GeoM.Translate(-50, -50)
	wheelOpt.GeoM.Rotate(-turning * (math.Pi / 2))
	wheelOpt.GeoM.Translate(50, 130)
	screen.DrawImage(wheel, wheelOpt)
}

type BoostVisu struct {
	boost *ebiten.Image
}

func InitBoostVisu() *BoostVisu {

	boost, _ := ebiten.NewImage(10, 50, ebiten.FilterDefault)
	boost.Fill(color.RGBA{0, 255, 0, 255})
	return &BoostVisu{boost: boost}
}

func (b *BoostVisu) Render(thrust float64, img *ebiten.Image) {
	o := &ebiten.DrawImageOptions{}
	o.GeoM.Scale(1, -1*thrust)
	o.GeoM.Translate(100, 250)
	o.ColorM.RotateHue(-1.55 * thrust)

	img.DrawImage(b.boost, o)
}
