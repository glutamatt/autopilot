package ia

import (
	"fmt"
	"math"
	"os"

	"github.com/glutamatt/autopilot/model"
)

func Basic(vehicule *model.Vehicule, path *[]model.Position) *model.Driving {
	//targets := (*path)[0]
	target := (*path)[len(*path)-1]
	targetFromV := model.Position{X: target.X - vehicule.X, Y: target.Y - vehicule.Y}
	angleTargetFromV := math.Atan2(targetFromV.Y, targetFromV.X)
	angleToDo := angleTargetFromV - vehicule.Rotation
	turning := math.Mod(math.Tanh(angleToDo), 2*math.Pi)
	fmt.Fprintf(os.Stderr, "angle to do : %.2f -- turning : %.2f\n", angleToDo, turning)
	return &model.Driving{Turning: turning}
}
