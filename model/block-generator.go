package model

import (
	"math/rand"
	"time"
)

func GenerateBlocks(w, h float64) []*Position {
	rSource := rand.New(rand.NewSource(time.Now().Unix()))
	blockCnt := 200
	forbidDistance := float64(5 * BlockBorder)
	positions := []*Position{}
	forbidenBlocks := map[Position]bool{}
	for i := 0; i < blockCnt; i++ {
		bw, bh := rSource.Float64()*50+15, rSource.Float64()*50+15
		bStart := &Position{rSource.Float64() * w, rSource.Float64() * -h}
		wBCount, hBCount := int(bw)/BlockBorder, int(bh)/BlockBorder

		for hi := 0; hi <= hBCount; hi++ {
			for wi := 0; wi <= wBCount; wi++ {
				pos := &Position{bStart.X - bw/2 + float64(wi*BlockBorder), bStart.Y + bh/2 - float64(hi*BlockBorder)}
				pos.Gap(BlockBorder)
				if forbiden := forbidenBlocks[*pos]; forbiden {
					continue
				}
				positions = append(positions, pos)
			}
		}

		forbidenBlocks = BlocksArround(forbidenBlocks, *bStart, bw, bh, forbidDistance)
	}

	return positions
}

func BlocksArround(forbidenBlocks map[Position]bool, bStart Position, bw, bh, forbidDistance float64) map[Position]bool {
	forbidPos := Position{bStart.X - bw/2 - forbidDistance, bStart.Y - bh/2 - forbidDistance}
	for forbidPos.X < bStart.X+bw/2+forbidDistance {
		for forbidPos.Y < bStart.Y+bh/2+forbidDistance {
			forbidPos.Gap(BlockBorder)
			forbidenBlocks[forbidPos] = true
			forbidPos.Y += float64(BlockBorder)
		}
		forbidPos.Y = bStart.Y - bh/2 - forbidDistance
		forbidPos.X += float64(BlockBorder)
	}

	return forbidenBlocks
}
