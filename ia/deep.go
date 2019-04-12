package ia

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/glutamatt/autopilot/model"
)

type Instance struct {
	Raw []float64 `json:"raw"`
}

type Req struct {
	SignatureName string     `json:"signature_name"`
	Instances     []Instance `json:"instances"`
}

type Predictions struct {
	Preds [][]float64 `json:"predictions"`
}

func NeuralNet(features []float64) *model.Driving {
	modelName, version := "model_9", "9"

	serviceURL := fmt.Sprintf("http://localhost:8501/v1/models/%s/versions/%s:predict", modelName, version)

	body, err := json.Marshal(&Req{
		SignatureName: "raw",
		Instances: []Instance{
			Instance{Raw: features},
		},
	})

	if err != nil {
		log.Fatal("json.Marshal error", err)
	}

	resp, err := http.Post(serviceURL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatal("http post error", err)
	}

	predictions := &Predictions{}
	driving := model.Driving{Thrust: .5}

	if err := json.NewDecoder(resp.Body).Decode(predictions); err == nil {
		driving.Turning = predictions.Preds[0][0]
		driving.Thrust = predictions.Preds[0][1]
	}

	return &driving
}
