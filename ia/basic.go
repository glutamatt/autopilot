package ia

import (
	"fmt"
	"math"

	"github.com/glutamatt/autopilot/model"
)

func Basic(vehicule *model.Vehicule, path *[]model.Position) *model.Driving {
	//targets := (*path)[0]
	target := (*path)[len(*path)-1]
	targetFromV := model.Position{X: target.X - vehicule.X, Y: target.Y - vehicule.Y}
	fmt.Printf("target: %v\ntargetFromV : %v\n", target, targetFromV)
	angleTargetFromV := math.Mod(math.Atan(targetFromV.Y/targetFromV.X), 2.0*math.Pi)
	if angleTargetFromV > math.Pi {
		angleTargetFromV = 2.0*math.Pi - angleTargetFromV
	}
	angleToDo := angleTargetFromV - vehicule.Rotation
	turning := 1/(1+math.Pow(math.E, angleToDo)) - .5
	/*
		180 -> pi
		? -> ang
		? = ang * 180 / pi

		tan = o / a


	*/
	fmt.Printf("angleTargetFromV : %.2f  : turning : %.2f  : angle to do : %.2f\n", angleTargetFromV*180/math.Pi, turning, angleToDo*180/math.Pi)
	return &model.Driving{Turning: turning}
}
