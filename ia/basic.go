package ia

import (
	"math"

	"github.com/glutamatt/autopilot/model"
)

func Basic(vehicule *model.Vehicule, path []model.Position) *model.Driving {
	target := path[len(path)-1]
	targetFromV := model.Position{X: target.X - vehicule.X, Y: target.Y - vehicule.Y}
	angleTargetFromV := math.Atan2(targetFromV.Y, targetFromV.X)
	angleToDo := angleTargetFromV - vehicule.Rotation
	if angleToDo > math.Pi {
		angleToDo -= math.Pi * 2
	}
	if angleToDo < -math.Pi {
		angleToDo += math.Pi * 2
	}
	return &model.Driving{Turning: math.Tanh(angleToDo)}
}
