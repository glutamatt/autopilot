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
	//up
	if p, ok := mapTiles[Position{X: t.Position.X, Y: t.Position.Y + float64(BlockBorder)}]; ok {
		paths = append(paths, p)
	}
	//down
	if p, ok := mapTiles[Position{X: t.Position.X, Y: t.Position.Y - float64(BlockBorder)}]; ok {
		paths = append(paths, p)
	}
	//left
	if p, ok := mapTiles[Position{X: t.Position.X + float64(BlockBorder), Y: t.Position.Y}]; ok {
		paths = append(paths, p)
	}
	//right
	if p, ok := mapTiles[Position{X: t.Position.X - float64(BlockBorder), Y: t.Position.Y}]; ok {
		paths = append(paths, p)
	}
	//up right
	if p, ok := mapTiles[Position{X: t.Position.X + float64(BlockBorder), Y: t.Position.Y + float64(BlockBorder)}]; ok {
		paths = append(paths, p)
	}
	//up left
	if p, ok := mapTiles[Position{X: t.Position.X - float64(BlockBorder), Y: t.Position.Y - float64(BlockBorder)}]; ok {
		paths = append(paths, p)
	}
	//down right
	if p, ok := mapTiles[Position{X: t.Position.X - float64(BlockBorder), Y: t.Position.Y + float64(BlockBorder)}]; ok {
		paths = append(paths, p)
	}
	//down left
	if p, ok := mapTiles[Position{X: t.Position.X - float64(BlockBorder), Y: t.Position.Y - float64(BlockBorder)}]; ok {
		paths = append(paths, p)
	}

	return paths
}

func (t *Tile) PathNeighborCost(to astar.Pather) float64 {
	toT, toto := to.(*Tile)
	if !toto || toT == nil {
		return math.Inf(1)
	}
	if _, exist := (*blockMap)[toT.Position]; exist {
		return math.Inf(1)
	}

	{
		_, diag1 := (*blockMap)[Position{X: t.X, Y: toT.Y}]
		_, diag2 := (*blockMap)[Position{X: toT.X, Y: t.Y}]
		if diag1 && diag2 {
			return math.Inf(1)
		}
	}

	if t.X != toT.X && t.Y != toT.Y {
		return 1.4
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
	startTile, ok := mapTiles[*start]

	if !ok {
		return false, nil
	}

	end := &to
	end.Gap(BlockBorder)
	endTile, ok := mapTiles[*end]

	if !ok {
		return false, nil
	}

	path, _, found := astar.Path(startTile, endTile)
	if !found {
		return false, nil
	}

	pathPositions := make([]Position, len(path))
	for i, p := range path {
		p := p.(*Tile)
		pathPositions[i] = p.Position
	}

	pathPositions = pathPositions[:len(pathPositions)-2]
	if len(pathPositions) > 0 {
		return true, pathPositions
	}
	return false, nil
}
