package model

import "math/rand"

func GenerateBlocks(w, h float64) []*Position {
	blockCnt := 100
	pos := make([]*Position, blockCnt)
	for i := 0; i < blockCnt; i++ {
		pos[i] = &Position{rand.Float64() * w, rand.Float64() * -h}
		pos[i].Gap(BlockBorder)
	}

	return pos
}
