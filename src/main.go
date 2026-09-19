package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
)

// Game window resolution
const screenWidth = 256
const screenHeight = 256

// Tile size:
const tileSize = 32

// Tile constant to map grid IDs to semantic meanings
const (
	TileGrass = iota 	// 0
	TileWall					// 1
)

const (
	DirDown = iota	// 0
	DirLeft					// 1
	DirRight				// 2
	DirUp						// 3
)

// Variable: tile image that contains all the available tiles
// Will be initialized by init() function
var tileSetImage *ebiten.Image
var playerSetImage *ebiten.Image

// Function: init
// special, predefined function that executes automatically when a package is
// initialized. Takes no parameters, returns no vale
// This function is used to read in Tiles.png and convert it to an image data
// structure
func init() {
	// Decode and image from the image file's byte slice.
	img, _, err := image.Decode(bytes.NewReader(images.Tiles_png))
	if err != nil {
		log.Fatal(err)
	}
	tileSetImage = ebiten.NewImageFromImage(img)

	// Decode runner image
	img, _, err = image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		log.Fatal(err)
	}
	playerSetImage = ebiten.NewImageFromImage(img)
}

// Tilemap structure: stores layout matrix and handles boundary logic
type Tilemap struct {
	Grid [][]int
}

// Tilemap Method: Width returns the map width in grid units
func (m *Tilemap) Width() int {
	if len(m.Grid) == 0 {
		return 0
	}
	return len(m.Grid[0])
}

// Tilemap Method: Height returns height of grid in units
func (m *Tilemap) Height() int {
	return len(m.Grid)
}

// Tilemap Method: IsWalkable returns tru if coordinate is within bounds and
// not wall
func (m *Tilemap) IsWalkable(x, y int) bool {
	if x < 0 || x >= m.Width() || y < 0 || y >= m.Height() {
		return false	// Out-of-bounds
	}
	tile := m.Grid[y][x]
	return tile != TileWall
}

// Function: return start pixel (both dimension) of a tile inside of the map
// tileset based on what tile type it is
func GetTileCoord(tileType int) (x, y int) {
	var tileGridX, tileGridY int
	if tileType == TileGrass {
		tileGridX = 8
		tileGridY = 1
	} else if tileType == TileWall {
		tileGridX = 2
		tileGridY = 4
	} else {
		log.Fatalf("[ERROR] Unknown tile type.")
	}
	return tileGridX * tileSize, tileGridY * tileSize
}

// Function: return start pixel (both dimension) of a tile insde of the player
// tileset based on what action it is
func GetPlayerCoord(actionType int) (x, y int) {
	var playerGridX, playerGridY int
	if actionType == DirLeft {
		playerGridX = 0
		playerGridY = 1
	} else if actionType == DirRight {
		playerGridX = 1
		playerGridY = 1
	} else if actionType == DirUp {
		playerGridX = 2
		playerGridY = 1
	} else if actionType == DirDown {
		playerGridX = 3
		playerGridY = 1
	} else {
		log.Fatalf("[ERROR] Unknown player action.")
	}
	return playerGridX * tileSize, playerGridY * tileSize
}

// Player structure: tracks player position
type Player struct {
	GridX int
	GridY int
	Dir int
}

// Game data structure
type Game struct {
	Tilemap Tilemap
	Player Player
	inputDelay int
	tileSetImage *ebiten.Image
	playerSetImage *ebiten.Image
}

// Game Method: Update
func (g *Game) Update() error {
	// Input delay is used for frame cooldown
	if g.inputDelay > 0 {
		g.inputDelay--
		return nil
	}

	nextX, nextY := g.Player.GridX, g.Player.GridY
	moved := false

	// Capture intent
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		fmt.Printf("Key Press: Arrow Left\n")
		nextX--
		g.Player.Dir = DirLeft
		moved = true
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		fmt.Printf("Key Press: Arrow Right\n")
		nextX++
		g.Player.Dir = DirRight
		moved = true
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		fmt.Printf("Key Press: Arrow Up\n")
		nextY--
		g.Player.Dir = DirUp
		moved = true
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		fmt.Printf("Key Press: Arrow Down\n")
		nextY++
		g.Player.Dir = DirDown
		moved = true
	}

	// Collision check: if walkable, then update player position
	if moved {
		if g.Tilemap.IsWalkable(nextX, nextY) {
			g.Player.GridX = nextX
			g.Player.GridY = nextY
		}
		g.inputDelay = 5 // Frame cooldown
	}

	return nil
}

// Game Method: Draw background (called in Draw method)
func (g *Game) DrawBackground(screen *ebiten.Image) {
	// Keep player focused in screen center
	playerPixelX := g.Player.GridX * tileSize
	playerPixelY := g.Player.GridY * tileSize
	// Camera position
	camX := float64((screenWidth / 2) - playerPixelX - (tileSize / 2))
	camY := float64((screenHeight / 2) - playerPixelY - (tileSize / 2))
	// Draw each tile in the background
	for y:= 0; y < g.Tilemap.Height(); y++ {
		for x := 0; x < g.Tilemap.Width(); x++ {
			tileType := g.Tilemap.Grid[y][x]
			// Calculate where the tile lives in the tile image
			srcX, srcY := GetTileCoord(tileType)
			// Crop out the precise 32x32 tile rectangle with SubImage
			rect := image.Rect(srcX, srcY, srcX + tileSize, srcY + tileSize)
			tileSprite := g.tileSetImage.SubImage(rect).(*ebiten.Image)
			// Matrix configuration ot translate the image location
			op := &ebiten.DrawImageOptions{}
			renderX := float64(x * tileSize) + camX
			renderY := float64(y * tileSize) + camY
			op.GeoM.Translate(renderX, renderY)
			// Draw call
			screen.DrawImage(tileSprite, op)
		}
	}
}

func (g *Game) DrawPlayer(screen *ebiten.Image) {
	// Keep player focused in screen center
	playerPixelX := g.Player.GridX * tileSize
	playerPixelY := g.Player.GridY * tileSize
	// Camera position
	camX := float64((screenWidth / 2) - playerPixelX - (tileSize / 2))
	camY := float64((screenHeight / 2) - playerPixelY - (tileSize / 2))
	// Get player start coordinate inside of the player tileset
	pSrcX, pSrcY := GetPlayerCoord(g.Player.Dir)
	// Get the player rectangle image coordinate
	//pRect := image.Rect(pSrcX, pSrcY, pSrcX + tileSize, pSrcY + tileSize)
	pRect := image.Rect(pSrcX + tileSize, pSrcY + tileSize, pSrcX, pSrcY)
	// Extract the subimage from the player tileset
	playerSprite := g.playerSetImage.SubImage(pRect).(*ebiten.Image)
	// Set up the draw options (start point of the draw call)
	popts := &ebiten.DrawImageOptions{}
	// If player is facing left, flip the player image
	if g.Player.Dir == DirLeft {
		// Flip horizontally
		popts.GeoM.Scale(-1, 1)
		// Push it back to bounding square
		popts.GeoM.Translate(float64(tileSize), 0)
	}
	pRenderX := float64(playerPixelX) + camX
	pRenderY := float64(playerPixelY) + camY
	popts.GeoM.Translate(pRenderX, pRenderY)
	// Draw call
	screen.DrawImage(playerSprite, popts)
}

// Game Method: Draw
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear out canvas first
	screen.Fill(color.RGBA{0, 0, 0, 255})

	// Draw background
	g.DrawBackground(screen)

	// Draw Player
	g.DrawPlayer(screen)

	// Display Tick per second
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))

}

// Game Method: Layout
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// Main
func main() {
	// Set window properties
	ebiten.SetWindowSize(screenWidth * 3, screenHeight * 3)
	ebiten.SetWindowTitle("Tiles (Ebitengine Demo)")

	// Define layout blueprint array
	// 10 x 10 grid
	initialMap := Tilemap {
		Grid: [][]int {
			{TileWall , TileWall , TileWall , TileWall , TileWall , TileWall , TileWall , TileWall , TileWall , TileWall },
			{TileWall , TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileWall },
			{TileWall , TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileWall },
			{TileWall , TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileWall },
			{TileWall , TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileWall },
			{TileWall , TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileWall },
			{TileWall , TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileWall },
			{TileWall , TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileWall },
			{TileWall , TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileGrass, TileWall },
			{TileWall , TileWall , TileWall , TileWall , TileWall , TileWall , TileWall , TileWall , TileWall , TileWall },
		},
	}

	// Instantiate game state
	g := &Game {
		Tilemap: initialMap,
		tileSetImage: tileSetImage,
		playerSetImage: playerSetImage,
		Player: Player {
			GridX: 5,
			GridY: 5,
			Dir: DirDown,
		},
	}

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
