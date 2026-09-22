package main

import (
	"cmp"
	"slices"
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
const screenWidth = 432
const screenHeight = 240

// Tile size:
const tileSize = 16

// Tile constant to map grid IDs to semantic meanings
const (
	Walkable = iota // 0
	Block // 1
)

// Scenes
const (
	SceneParkingLot = iota // 0
	SceneCourtYard // 1
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

// Renderable Item ID
const (
	PlayerID = iota // 0
	TreeID // 1
)

// Declare the embedded compile-time asset bytes
//go:embed assets/exterior-sprites.png
var exteriorByteData []byte
//go:embed assets/player-sprites.png
var playerByteData []byte
//go:embed assets/parking-lot.png
var parkingLotByteData []byte
//go:embed assets/parking-lot-collision.csv
var parkingLotCsvStr string
//go:embed assets/court-yard.png
var courtYardByteData []byte
//go:embed assets/court-yard-collision.csv
var courtYardCsvStr string

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

// RenderItem holds everything needed for a single draw call for an object
// This is needed so that the objects can be drawn with depth wrt player
type RenderItem struct {
	ID int
	BaseY int // Y coordinate of the base to the object, i.e. foot
	ScreenDstX, ScreenDstY int // Coordinate of object on the drawn screen
	SpriteImg *ebiten.Image
}

// Scene data structure
type Scene struct {
	ID int
	BgImage *ebiten.Image
	Collision [][]int
	WidthPixels float64
	HeightPixels float64
	RenderItems []RenderItem
}

// Game data structure
type Game struct {
	Scenes map[int]*Scene
	CurrentScene int
	Player Player
	playerSetImage *ebiten.Image
	TouchButtons []TouchButton
	// Queue of render items, reset with slice[:0] to clear length but preserve
	// underlying array capacity
	RenderQueue []RenderItem
}

// Game Method: IsWalkable returns tru if coordinate is within bounds and
// not wall
func (g *Game) IsWalkable(x, y int) bool {
	currentScene := g.Scenes[g.CurrentScene]
	//fmt.Printf("Scene: %d  X: %d  Y: %d  Collision: %d\n",
	//	currentScene.ID, x, y, currentScene.Collision[y][x])
	if y <= 0 || y >= len(currentScene.Collision) || x <= 0 || x >= len(currentScene.Collision[y]) {
		//fmt.Printf("Out of bounds\n")
		return false	// Out-of-bounds
	} else {
		return currentScene.Collision[y][x] != Block
	}
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
		if g.IsWalkable(nextX, nextY) {
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
	// Grab current scene
	currentScene := g.Scenes[g.CurrentScene]
	// Keep player focused in screen center
	playerPixelX := g.Player.GridX * tileSize
	playerPixelY := g.Player.GridY * tileSize
	// Camera position
	camX := float64((screenWidth / 2) - playerPixelX - (tileSize / 2))
	camY := float64((screenHeight / 2) - playerPixelY - (tileSize / 2))
	// Draw call
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(camX, camY)
	screen.DrawImage(currentScene.BgImage, op)
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

// Game Method: Draw Y-sorted Layer (called in Draw method)
func (g *Game) DrawYSortedLayer(screen *ebiten.Image) {
	for _, item := range g.RenderQueue {
		if item.ID == PlayerID { // If item is player, call DrawPlayer
			g.DrawPlayer(screen)
		} else {
			// Keep player focused in screen center
			playerPixelX := g.Player.GridX * tileSize
			playerPixelY := g.Player.GridY * tileSize
			// Camera position
			camX := float64((screenWidth / 2) - playerPixelX - (tileSize / 2))
			camY := float64((screenHeight / 2) - playerPixelY - (tileSize / 2))
			// Get player start coordinate inside of the player tileset
			pSrcX, pSrcY, pDstX, pDstY := GetExteriorItemCoord(item.ID)
			// Get the player rectangle image coordinate
			pRect := image.Rect(pSrcX, pSrcY, pDstX, pDstY)
			// Extract the subimage from the player tileset
			itemSprite := item.SpriteImg.SubImage(pRect).(*ebiten.Image)
			// Set up the draw options (start point of the draw call)
			popts := &ebiten.DrawImageOptions{}
			pRenderX := float64(item.ScreenDstX) + camX
			pRenderY := float64(item.ScreenDstY) + camY
			popts.GeoM.Translate(pRenderX, pRenderY)
			// Draw call
			screen.DrawImage(itemSprite, popts)
		}
	}
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
	// Set up the Y-sorted layer queue
	// Clear it first
	g.RenderQueue = g.RenderQueue[:0]
	// Push playe to queue
	g.RenderQueue = append(g.RenderQueue, RenderItem {
		ID: PlayerID,
		BaseY: (g.Player.GridY + 1) * tileSize, // player's feet
		ScreenDstX: g.Player.GridX * tileSize,
		ScreenDstY: g.Player.GridY * tileSize,
		SpriteImg: g.playerSetImage,
	})
	// Push all the entity in scene
	for _, item := range g.Scenes[g.CurrentScene].RenderItems {
		g.RenderQueue = append(g.RenderQueue, item)
	}
	// Sort the layer by Y value
	slices.SortFunc(g.RenderQueue, func(a, b RenderItem) int {
		return cmp.Compare(a.BaseY, b.BaseY)
	})	
	// Draw Y-sorted layer
	g.DrawYSortedLayer(screen)
	// Draw virtual buttons
	g.DrawTouchButtons(screen)
	// Display Tick per second
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
}

// Game Method: Layout
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// Helper functions
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
		playerGridY = 2
	} else if actionType == ActionWalk {
		playerGridY = 4
	} else {
		log.Fatalf("[ERROR] Unknown player action.")
	}
	return playerGridX * tileSize, playerGridY * tileSize
}

// Function to load embedded image
func loadEmbeddedImage(byteData []byte) *ebiten.Image {
	// Decode and image from the image file's byte slice.
	img, _, err := image.Decode(bytes.NewReader(byteData))
	if err != nil {
		log.Fatal(err)
	}
	return ebiten.NewImageFromImage(img)
}

// Function to load collision csv
func loadCollisionCsv(csvStr string) [][]int {
	// Load scene walkable csv
	reader := csv.NewReader(strings.NewReader(csvStr))
	// Loop through each line
	var collisionCsv [][]int
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
		collisionCsv = append(collisionCsv, intRow)
	}
	return collisionCsv
}

// Global declaration of RenderItems
var exteriorImage *ebiten.Image

func GetExteriorItemCoord(itemID int) (int, int, int, int) {
	var GridSrcX, GridSrcY, GridDstX, GridDstY int
	switch itemID {
	case TreeID:
		GridSrcX = 33
		GridSrcY = 10
		GridDstX = GridSrcX + 1
		GridDstY = GridSrcY + 2
	default:
		GridSrcX = 0
		GridDstX = 1
		GridSrcY = 0
		GridDstY = 1
	}
	return GridSrcX * tileSize, GridSrcY * tileSize, GridDstX * tileSize, GridDstY * tileSize
}

// Main
func main() {
	// Set window properties
	ebiten.SetWindowSize(screenWidth * 3, screenHeight * 3)
	ebiten.SetWindowTitle("Tiles (Ebitengine Demo)")
	ebiten.SetTPS(30)

	// Read in assets
	exteriorImage := loadEmbeddedImage(exteriorByteData)
	parkingLotBg := loadEmbeddedImage(parkingLotByteData)
	courtYardBg := loadEmbeddedImage(courtYardByteData)
	playerSetImage := loadEmbeddedImage(playerByteData)

	// Build scene registry index container map
	gameScenes := make(map[int]*Scene)
	// Scene: Parking Lot
	gameScenes[SceneParkingLot] = &Scene {
		ID: SceneParkingLot,
		BgImage: parkingLotBg,
		Collision: loadCollisionCsv(parkingLotCsvStr),
		WidthPixels: float64(parkingLotBg.Bounds().Dx()),
		HeightPixels: float64(parkingLotBg.Bounds().Dy()),
		RenderItems: []RenderItem {
			{
				ID: TreeID,
				BaseY: 23 * tileSize, // base of tree
				ScreenDstX: 2 * tileSize,
				ScreenDstY: 22 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 23 * tileSize, // base of tree
				ScreenDstX: 10 * tileSize,
				ScreenDstY: 22 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 21 * tileSize, // base of tree
				ScreenDstX: 17 * tileSize,
				ScreenDstY: 20 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 13 * tileSize, // base of tree
				ScreenDstX: 6 * tileSize,
				ScreenDstY: 12 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 13 * tileSize, // base of tree
				ScreenDstX: 14 * tileSize,
				ScreenDstY: 12 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 11 * tileSize, // base of tree
				ScreenDstX: 17 * tileSize,
				ScreenDstY: 10 * tileSize,
				SpriteImg: exteriorImage,
			},
		},
	}

	// Scene: Court Yard
	gameScenes[SceneCourtYard] = &Scene {
		ID: SceneCourtYard,
		BgImage: courtYardBg,
		Collision: loadCollisionCsv(courtYardCsvStr),
		WidthPixels: float64(courtYardBg.Bounds().Dx()),
		HeightPixels: float64(courtYardBg.Bounds().Dy()),
		RenderItems: []RenderItem {
			{
				ID: TreeID,
				BaseY: 19 * tileSize, // base of tree
				ScreenDstX: 2 * tileSize,
				ScreenDstY: 18 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 15 * tileSize, // base of tree
				ScreenDstX: 2 * tileSize,
				ScreenDstY: 14 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 7 * tileSize, // base of tree
				ScreenDstX: 3 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 6 * tileSize, // base of tree
				ScreenDstX: 5 * tileSize,
				ScreenDstY: 5 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 7 * tileSize, // base of tree
				ScreenDstX: 7 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 6 * tileSize, // base of tree
				ScreenDstX: 9 * tileSize,
				ScreenDstY: 5 * tileSize,
				SpriteImg: exteriorImage,
			},
		},
	}

	// Initialize player start postion on scene in grid units
	playerStartX := 3
	playerStartY := 3

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
		Scenes: gameScenes,
		CurrentScene: SceneParkingLot,
		playerSetImage: playerSetImage,
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
