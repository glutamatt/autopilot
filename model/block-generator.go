package model

import (
	"math/rand"
	"time"
)

func GenerateBlocks(w, h float64) []*Position {
	rSource := rand.New(rand.NewSource(time.Now().Unix()))

	blockCnt := 140
	positions := []*Position{}
	for i := 0; i < blockCnt; i++ {
		bw, bh := rSource.Float64()*20+5, rSource.Float64()*20+5
		bStart := &Position{rSource.Float64() * w, rSource.Float64() * -h}
		wBCount, hBCount := int(bw)/BlockBorder, int(bh)/BlockBorder

		for hi := 0; hi <= hBCount; hi++ {
			for wi := 0; wi <= wBCount; wi++ {
				pos := &Position{bStart.X - bw/2 + float64(wi*BlockBorder), bStart.Y + bh/2 - float64(hi*BlockBorder)}
				pos.Gap(BlockBorder)
				positions = append(positions, pos)
			}
		}
	}

	return positions
}
