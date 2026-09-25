package main

import (
	"cmp"
	"slices"
	"math"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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
	ActA // 4
	ActB // 5
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
	ActionInteract // 2
)

// Player interaction keys
const (
	InteractA = iota // 0
	InteractD // 1
)

// Interaction types
const (
	InteractionTypeConversation = iota // 0
	InteractionTypeMiniGame0 // 1
)

	// Renderable Item ID
const (
	PlayerID = iota // 0
	TreeID // 1
)

// Type of button
const (
	ButtonRight = iota // 0
	ButtonUp // 1
	ButtonLeft // 2
	ButtonDown // 3
	ButtonInteractA // 4
	ButtonInteractD // 5
)

// Touch Button defines a simple interactive screen bounding box area
type TouchButton struct {
	ButtonType int // movement or interaction A or interaction D
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

// WarpTrigger defines a specific tile on the current scene where there is a
// transition
type WarpTrigger struct {
	DespawnX, DespawnY int // Coordinate of tile in current scene in grid unit
	SpawnX, SpawnY int // Coordinate of tile in next scene in grid unit
	TargetScene int // ID of next scene
}

// InteractionTrigger defines a specific tile one the current scene where
// there is an interaction
type InteractionTrigger struct {
	GridX, GridY int // Coordinate of tile in current scene in grid unit
	PlayerDir int // Direction that player is facing
	InteractionID int // ID of interaction (referenced to interaction structure)
}

// Interaction structure
type Interaction struct {
	InteractionID int // ID of interaction (referenced from the interaction trigger)
	InteractionType int // type of interaction (conversation, minigame, etc)
	InteractionStates int // number of states of interaction
	InteractionText []string // scripts for texbox type interaction
}

// Scene data structure
type Scene struct {
	ID int
	BgImage *ebiten.Image
	Collision [][]int
	WidthPixels float64
	HeightPixels float64
	RenderItems []RenderItem
	WarpTriggers []WarpTrigger
	InteractionTriggers []InteractionTrigger
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
	Interactions []Interaction
	CurrentInteraction int
	CurrentInteractionState int
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
	// If player is in interaction
	if g.Player.Action == ActionInteract {
		// Capture interaction key presses
		if inpututil.IsKeyJustPressed(ebiten.KeyA) {
			fmt.Printf("Key Press: A\n")
			g.CurrentInteractionState++
			fmt.Printf("Interaction State: %d\n", g.CurrentInteractionState)
		} else if inpututil.IsKeyJustPressed(ebiten.KeyD) {
			fmt.Printf("Key Press: D\n")
			g.CurrentInteractionState = g.Interactions[g.CurrentInteraction].InteractionStates
		} else {
			touchIDs := ebiten.TouchIDs()
			for _, id := range touchIDs {
				tx, ty := ebiten.TouchPosition(id) // Grab touch position
				// Loop through our TouchButtons
				for _, b := range g.TouchButtons {
					if tx >= b.boundX && tx <= b.boundX + b.boundWidth &&
					ty >= b.boundY && ty <= b.boundY + b.boundHeight {
						if b.ButtonType == ButtonInteractA {
							g.CurrentInteractionState++
						} else if b.ButtonType == ButtonInteractD {
							g.CurrentInteractionState = g.Interactions[g.CurrentInteraction].InteractionStates
						}
						// Stop checking other buttons when this one is a hit
						break
					}
				}
			}
		}

		// Reset to idle if done with interaction
		if g.CurrentInteractionState == g.Interactions[g.CurrentInteraction].InteractionStates {
			g.Player.Action = ActionIdle
		}
		return nil
	}

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
	interacted := false
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		fmt.Printf("Key Press: A\n")
		interacted = true
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
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
	} else { // Check for mobile virtual button inputs// Grab all active finger touch IDs pressed on browser
		touchIDs := ebiten.TouchIDs()
		for _, id := range touchIDs {
			tx, ty := ebiten.TouchPosition(id) // Grab touch position
			// Loop through our TouchButtons
			for _, b := range g.TouchButtons {
				if tx >= b.boundX && tx <= b.boundX + b.boundWidth &&
				ty >= b.boundY && ty <= b.boundY + b.boundHeight {
					if b.ButtonType == ButtonLeft {
						// Finger touch is within bounding box of button
						activeDir = b.Dir
						moved = true
						nextX--
					} else if b.ButtonType == ButtonRight {
						activeDir = b.Dir
						moved = true
						nextX++
					} else if b.ButtonType == ButtonUp {
						activeDir = b.Dir
						moved = true
						nextY--
					} else if b.ButtonType == ButtonDown {
						activeDir = b.Dir
						moved = true
						nextY++
					} else if b.ButtonType == ButtonInteractA {
						interacted = true
					}
					// Stop checking other buttons when this one is a hit
					break
				}
			}
		}
	}

	// Checking for first interaction trigger
	// Interaction is higher priority than movement
	if interacted {
		fmt.Printf("Interacted : %t\n", interacted)
		// Check if tile and direction allows for an interaction
		for _, trigger := range g.Scenes[g.CurrentScene].InteractionTriggers {
			if g.Player.GridX == trigger.GridX && g.Player.GridY == trigger.GridY &&
			g.Player.Dir == trigger.PlayerDir {
				g.Player.Action = ActionInteract
				g.CurrentInteraction = trigger.InteractionID
				g.CurrentInteractionState = 0 // start interaction in start state
			}
		}
	} else if moved {
		// Collision check: if walkable, then update player position
		g.Player.Dir = activeDir
		// Check if next tile is a WarpTrigger
		for _, trigger := range g.Scenes[g.CurrentScene].WarpTriggers {
			if nextX == trigger.DespawnX && nextY == trigger.DespawnY {
				fmt.Printf("Transition triggered: nextX: %d\tnextY: %d\n", nextX, nextY)
				g.CurrentScene = trigger.TargetScene
				g.Player.GridX = trigger.SpawnX
				g.Player.GridY = trigger.SpawnY
				g.Player.Action = ActionIdle
				fmt.Printf("Transition triggered: GridX: %d\tGridY: %d\n", g.Player.GridX, g.Player.GridY)
			} else if g.IsWalkable(nextX, nextY) {
				g.Player.GridX = nextX
				g.Player.GridY = nextY
				g.Player.Action = ActionWalk
			}
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
	// Check if we're currently in an interaction
	if g.Player.Action == ActionInteract {
		// Check if it's a conversation interaction
		interaction := g.Interactions[g.CurrentInteraction]
		if interaction.InteractionType == InteractionTypeConversation {
			// Draw textbox
			// Set textbox dimensions
			var boxX float32 = 16
			var boxY float32 = 160
			var boxWidth float32 = 400
			var boxHeight float32 = 64
			// Draw black background box
			boxColor := color.NRGBA{0, 0, 0, 200} // white box
			vector.DrawFilledRect(screen, boxX, boxY, boxWidth, boxHeight, boxColor, false)
			// Draw white border
			borderColor := color.NRGBA{255, 255, 255, 255}
			vector.StrokeRect(screen, boxX, boxY, boxWidth, boxHeight, 1.5, borderColor, false)
			// Draw text (offset slightly from top left edge)
			ebitenutil.DebugPrintAt(screen, interaction.InteractionText[g.CurrentInteractionState],
				int(boxX) + 12, int(boxY) + 12)
		}
	}
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
	ebiten.SetWindowSize(screenWidth * 3, screenHeight * 3)
	ebiten.SetWindowTitle("Tiles (Ebitengine Demo)")
	ebiten.SetTPS(30)

	// Initialize player start postion on scene in grid units
	playerStartX := 3
	playerStartY := 3

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
		Interactions: gameInteractions,
	}

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
