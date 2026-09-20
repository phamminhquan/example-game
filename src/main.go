package main

import (
	"encoding/csv"
	"io"
	"strconv"
	"os"
	_ "embed"
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Game window resolution
const screenWidth = 256
const screenHeight = 256

// Tile size:
const tileSize = 16

// Tile constant to map grid IDs to semantic meanings
const (
	Walkable = iota 	// 0
	Block					// 1
)

const (
	DirDown = iota	// 0
	DirLeft					// 1
	DirRight				// 2
	DirUp						// 3
)

// Variable: tile image that contains all the available tiles
// Will be initialized by init() function
var sceneBgImage []*ebiten.Image
var playerSetImage *ebiten.Image

// Walkable CSV
var walkableCsv [][]int

// Function: init
// special, predefined function that executes automatically when a package is
// initialized. Takes no parameters, returns no vale
// This function is used to read in Tiles.png and convert it to an image data
// structure
func init() {
	// Decode and image from the image file's byte slice.
	imgBytes, err := os.ReadFile("tiled-project/parking-lot.png")
	if err != nil {
		log.Fatal(err)
	}
	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	//img, _, err := image.Decode(bytes.NewReader(images.Tiles_png))
	if err != nil {
		log.Fatal(err)
	}
	sceneBgImage = append(sceneBgImage, ebiten.NewImageFromImage(img))

	// Decode player image
	imgBytes, err = os.ReadFile("tiled-project/character.png")
	if err != nil {
		log.Fatal(err)
	}
	img, _, err = image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		log.Fatal(err)
	}
	playerSetImage = ebiten.NewImageFromImage(img)

	// Load scene walkable csv
	file, err := os.Open("tiled-project/parking-lot.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	// Loop through each line
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		// Convert each string of '0' or '1' to an integer
		var intRow []int
		for _, val := range record {
			num, err := strconv.Atoi(val)
			if err != nil {
				log.Fatal(err)
			}
			intRow = append(intRow, num)
		}
		walkableCsv = append(walkableCsv, intRow)
	}
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
	return tile != Block
}

// Function: return start pixel (both dimension) of a tile insde of the player
// tileset based on what action it is
func GetPlayerCoord(actionType int) (x, y int) {
	var playerGridX, playerGridY int
	if actionType == DirLeft {
		playerGridX = 2
		playerGridY = 0
	} else if actionType == DirRight {
		playerGridX = 0
		playerGridY = 0
	} else if actionType == DirUp {
		playerGridX = 1
		playerGridY = 0
	} else if actionType == DirDown {
		playerGridX = 3
		playerGridY = 0
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
	Scene int
	sceneBgImage []*ebiten.Image
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
	// Draw call
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(camX, camY)
	screen.DrawImage(g.sceneBgImage[g.Scene], op)
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
	pRect := image.Rect(pSrcX, pSrcY, pSrcX + tileSize, pSrcY + 2 * tileSize)
	// Extract the subimage from the player tileset
	playerSprite := g.playerSetImage.SubImage(pRect).(*ebiten.Image)
	// Set up the draw options (start point of the draw call)
	popts := &ebiten.DrawImageOptions{}
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

	// Instantiate game state
	g := &Game {
		Tilemap: Tilemap {
			Grid: walkableCsv,
		},
		sceneBgImage: sceneBgImage,
		playerSetImage: playerSetImage,
		Scene: 0,
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
