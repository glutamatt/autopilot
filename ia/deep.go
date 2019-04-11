package ia

import "github.com/glutamatt/autopilot/model"

func NeuralNet(features []float64) *model.Driving {
	return &model.Driving{
		Thrust:  1.0,
		Turning: .5,
	}
}
