package main

import (
	"math"
	"encoding/csv"
	"io"
	"strconv"
	_ "embed"
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Game window resolution
const screenWidth = 320
const screenHeight = 180

// Tile size:
const tileSize = 16

// Tile constant to map grid IDs to semantic meanings
const (
	Walkable = iota // 0
	Block // 1
)

// Player facing direction
const (
	DirRight = iota // 0
	DirUp // 1
	DirLeft // 2
	DirDown // 3
)

// Player movespeed
// How many pixels player slides per frame
// Must divide cleanly uinto your tileSize, like 2.0 or 4.0 for 16/32px
const (
	moveSpeed = 4.0 // Frame sliding speed (pixel per tick)
	animSpeed = 0.2 // Animation speed (frame per tick)
)

// Player actions
const (
	ActionIdle = iota // 0
	ActionWalk // 1
)

// Variable: tile image that contains all the available tiles
// Will be initialized by init() function
var sceneBgImage []*ebiten.Image
var playerSetImage *ebiten.Image

// Declare the embedded compile-time asset bytes
//go:embed assets/parking-lot.png
var parkingLotByteData []byte
//go:embed assets/character.png
var characterByteData []byte
//go:embed assets/parking-lot.csv
var walkableCsvStr string

// Walkable CSV
var walkableCsv [][]int

// Function: init
// special, predefined function that executes automatically when a package is
// initialized. Takes no parameters, returns no vale
// This function is used to read in Tiles.png and convert it to an image data
// structure
func init() {
	// Decode and image from the image file's byte slice.
	img, _, err := image.Decode(bytes.NewReader(parkingLotByteData))
	if err != nil {
		log.Fatal(err)
	}
	sceneBgImage = append(sceneBgImage, ebiten.NewImageFromImage(img))

	// Decode player image
	img, _, err = image.Decode(bytes.NewReader(characterByteData))
	if err != nil {
		log.Fatal(err)
	}
	playerSetImage = ebiten.NewImageFromImage(img)

	// Load scene walkable csv
	reader := csv.NewReader(strings.NewReader(walkableCsvStr))
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
func GetPlayerCoord(actionType, dir, frame int) (x, y int) {
	// From the way the player tile set is set up row 2 contains all the
	// walking animation tyles. The first 6 tiles of row 2 is moving right.
	// Next 6 is moving up, next 6 is moving left, and next 6 is moving down
	// Row 1 of tileset is idle
	var playerGridX, playerGridY int
	playerGridX = (dir * 6) + frame
	if actionType == ActionIdle {
		//fmt.Printf("Action is Idle.\n")
		playerGridY = 2
	} else if actionType == ActionWalk {
		//fmt.Printf("Action is Walk.\n")
		playerGridY = 4
	} else {
		log.Fatalf("[ERROR] Unknown player action.")
	}

	return playerGridX * tileSize, playerGridY * tileSize
}

// Touch Button defines a simple interactive screen bounding box area
type TouchButton struct {
	boundX, boundY, boundWidth, boundHeight int // bounding box
	Dir int // maps to DirLeft, DirRight, DirUp, DirDown
}

// Player structure: tracks player position
type Player struct {
	GridX int // Logical coordinates
	GridY int
	PixelX float64 // Visual coordinates
	PixelY float64
	Action int // Player action
	Dir int // Facing direction
	AnimFrame int // 0 = idle, 1 = step left foot, 2 = idle, 3 = step right foot
	AnimProgress float64 // Tracks independent elapsed subframe time vector
}

// Game data structure
type Game struct {
	Tilemap Tilemap
	Player Player
	Scene int
	sceneBgImage []*ebiten.Image
	playerSetImage *ebiten.Image
	TouchButtons []TouchButton
}

// Game Method: Update
func (g *Game) Update() error {
	// If player is moving , handel visual sliding interpolation
	if g.Player.Action == ActionWalk {
		targetPixelX := float64(g.Player.GridX * tileSize)
		targetPixelY := float64(g.Player.GridY * tileSize)

		// Slide horizontally toward target
		if g.Player.PixelX < targetPixelX {
			g.Player.PixelX += moveSpeed
		} else if g.Player.PixelX > targetPixelX {
			g.Player.PixelX -= moveSpeed
		}

		// Slide vertically toward target
		if g.Player.PixelY < targetPixelY {
			g.Player.PixelY += moveSpeed
		} else if g.Player.PixelY > targetPixelY {
			g.Player.PixelY -= moveSpeed
		}

		// Independent animation layer
		// Accumulate fractional time completely separate from moveSpeed
		g.Player.AnimProgress += animSpeed
		if g.Player.AnimProgress >= 1.0 {
			g.Player.AnimProgress = 0.0
			g.Player.AnimFrame = (g.Player.AnimFrame + 1) % 6 // 6 frames per movement
		}

		// Check if we arrived perfectly at our destination
		if math.Abs(g.Player.PixelX - targetPixelX) < 0.1 &&
		math.Abs(g.Player.PixelY - targetPixelY) < 0.1 {
			g.Player.PixelX = targetPixelX
			g.Player.PixelY = targetPixelY
			g.Player.Action = ActionIdle
		} else {
			return nil
		}
	}

	// If standing still, snap visual pixels to the grid and look for new key
	// presses
	g.Player.PixelX = float64(g.Player.GridX * tileSize)
	g.Player.PixelY = float64(g.Player.GridY * tileSize)

	nextX, nextY := g.Player.GridX, g.Player.GridY
	moved := false

	// Capture key presses
	var activeDir int
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		fmt.Printf("Key Press: Arrow Left\n")
		nextX--
		activeDir = DirLeft
		moved = true
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		fmt.Printf("Key Press: Arrow Right\n")
		nextX++
		activeDir = DirRight
		moved = true
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		fmt.Printf("Key Press: Arrow Up\n")
		nextY--
		activeDir = DirUp
		moved = true
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		fmt.Printf("Key Press: Arrow Down\n")
		nextY++
		activeDir = DirDown
		moved = true
	}

	// Check for mobile virtual button inputs
	if !moved {
		// Grab all active finger touch IDs pressed on browser
		touchIDs := ebiten.TouchIDs()
		for _, id := range touchIDs {
			tx, ty := ebiten.TouchPosition(id) // Grab touch position
			// Loop through our TouchButtons
			for _, b := range g.TouchButtons {
				if tx >= b.boundX && tx <= b.boundX + b.boundWidth &&
				ty >= b.boundY && ty <= b.boundY + b.boundHeight {
					// Finger touch is within bounding box of button
					activeDir = b.Dir
					moved = true
					switch activeDir {
						case DirLeft: nextX--
						case DirRight: nextX++
						case DirUp: nextY--
						case DirDown: nextY++
					}
					// Stop checking other buttons when this one is a hit
					break
				}
			}
		}
	}

	// Collision check: if walkable, then update player position
	if moved {
		g.Player.Dir = activeDir
		if g.Tilemap.IsWalkable(nextX, nextY) {
			g.Player.GridX = nextX
			g.Player.GridY = nextY
			g.Player.Action = ActionWalk
		}
	} else {
		// If no key is held then we reset player to idle
		g.Player.AnimFrame = 2
		g.Player.AnimProgress = 0.0
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

// Game Method: Draw Player (called in Draw method)
func (g *Game) DrawPlayer(screen *ebiten.Image) {
	// Keep player focused in screen center
	playerPixelX := g.Player.GridX * tileSize
	playerPixelY := g.Player.GridY * tileSize
	// Camera position
	camX := float64((screenWidth / 2) - playerPixelX - (tileSize / 2))
	camY := float64((screenHeight / 2) - playerPixelY - (tileSize / 2))
	// Get player start coordinate inside of the player tileset
	pSrcX, pSrcY := GetPlayerCoord(g.Player.Action, g.Player.Dir, g.Player.AnimFrame)
	//pSrcX := (g.Player.Dir * 6 * tileSize) + g.Player.AnimFrame * tileSize
	//pSrcY := 4 * tileSize
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

// Game Method:  Draw touch buttons for mobile
func (g *Game) DrawTouchButtons(screen *ebiten.Image) {
	// Draw each button
	for _, b := range g.TouchButtons {
		vector.DrawFilledCircle(
			screen,
			float32(b.boundX + b.boundWidth / 2.0),
			float32(b.boundY + b.boundHeight / 2.0),
			float32(b.boundWidth / 2.0),
			color.NRGBA{255, 255, 255, 128},
			false,
		)
	}
}

// Game Method: Draw
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear out canvas first
	screen.Fill(color.RGBA{0, 0, 0, 255})

	// Draw background
	g.DrawBackground(screen)

	// Draw Player
	g.DrawPlayer(screen)

	// Draw virtual buttons
	g.DrawTouchButtons(screen)

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
	ebiten.SetWindowSize(screenWidth * 6, screenHeight * 4)
	ebiten.SetWindowTitle("Tiles (Ebitengine Demo)")
	ebiten.SetTPS(30)

	// Initialize player start postion on scene in grid units
	playerStartX := 5
	playerStartY := 5

	// Initialize TouchButtons bounding box for mobile
	buttonSize := 32
	padX := 40 // Bottom left cluster placement
	padY := 180 / 2
	mobileButtons := []TouchButton {
		{ // Up button
			boundX: padX,
			boundY: padY - buttonSize,
			boundWidth: buttonSize,
			boundHeight: buttonSize,
			Dir: DirUp,
		},
		{ // Down button
			boundX: padX,
			boundY: padY + buttonSize,
			boundWidth: buttonSize,
			boundHeight: buttonSize,
			Dir: DirDown,
		},
		{ // Left button
			boundX: padX - buttonSize,
			boundY: padY,
			boundWidth: buttonSize,
			boundHeight: buttonSize,
			Dir: DirLeft,
		},
		{ // Right button
			boundX: padX + buttonSize,
			boundY: padY,
			boundWidth: buttonSize,
			boundHeight: buttonSize,
			Dir: DirRight,
		},
	}

	// Instantiate game state
	g := &Game {
		Tilemap: Tilemap {
			Grid: walkableCsv,
		},
		sceneBgImage: sceneBgImage,
		playerSetImage: playerSetImage,
		Scene: 0,
		Player: Player {
			GridX: playerStartX, // Player start position in scene in grid unit
			GridY: playerStartY,
			PixelX: float64(playerStartX * tileSize), // Player start position in pixel
			PixelY: float64(playerStartY * tileSize),
			Action: ActionIdle,
			Dir: DirDown,
		},
		TouchButtons: mobileButtons,
	}

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
