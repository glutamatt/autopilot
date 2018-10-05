package model

import (
	"math"

	astar "github.com/beefsack/go-astar"
)

type Tile struct {
	Position
}

var blockMap *map[Position]bool
var BlockBorder = 5

var mapTiles map[Position]*Tile

func InitPathTiles(blockBorder, w, h int) {
	BlockBorder = blockBorder
	mapTiles = make(map[Position]*Tile)
	for x := 0; x < w; x += BlockBorder {
		for y := 0; y > h*-1; y -= BlockBorder {
			p := &Position{X: float64(x), Y: float64(y)}
			p.Gap(BlockBorder)
			mapTiles[*p] = &Tile{Position: *p}
		}
	}
}

func (t *Tile) PathNeighbors() []astar.Pather {
	paths := []astar.Pather{}
	if p, ok := mapTiles[Position{X: t.Position.X, Y: t.Position.Y + float64(BlockBorder)}]; ok {
		paths = append(paths, p)
	}
	if p, ok := mapTiles[Position{X: t.Position.X, Y: t.Position.Y - float64(BlockBorder)}]; ok {
		paths = append(paths, p)
	}
	if p, ok := mapTiles[Position{X: t.Position.X + float64(BlockBorder), Y: t.Position.Y}]; ok {
		paths = append(paths, p)
	}
	if p, ok := mapTiles[Position{X: t.Position.X - float64(BlockBorder), Y: t.Position.Y}]; ok {
		paths = append(paths, p)
	}

	return paths
}

func (t *Tile) PathNeighborCost(to astar.Pather) float64 {
	toT, toto := to.(*Tile)
	if !toto || toT == nil {
		return math.Inf(1)
	}
	//fmt.Printf("PathNeighborCost > blockMap: %v\n", blockMap)
	//fmt.Printf("PathNeighborCost > toT: %v\n", toT)
	if _, exist := (*blockMap)[toT.Position]; exist {
		return math.Inf(1)
	}
	return 1
}

func (t *Tile) PathEstimatedCost(to astar.Pather) float64 {
	toT, toto := to.(*Tile)
	if !toto || toT == nil {
		return math.Inf(1)
	}
	return math.Log(math.Abs(t.X-toT.X) + math.Abs(t.Y-toT.Y))
}

func FindPath(from, to Position, blocks *map[Position]bool) (bool, []Position) {
	blockMap = blocks

	start := &from
	start.Gap(BlockBorder)
	startTile := mapTiles[*start]

	end := &to
	end.Gap(BlockBorder)
	endTile := mapTiles[*end]

	path, _, found := astar.Path(startTile, endTile)
	if !found {
		return false, nil
	}

	pathPositions := make([]Position, len(path))
	for i, p := range path {
		p := p.(*Tile)
		pathPositions[i] = p.Position
	}

	return true, pathPositions
}
