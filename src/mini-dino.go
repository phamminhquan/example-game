package main

import (
	//"cmp"
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

// Game state
const (
	StateTitle = iota // 0
	StatePlay // 1
	StateOver // 2
	StateExit // 3
)

// Center X
const CenterX = 100

// Ground level Y
const GroundY = 164

// Gravity
const Gravity = 1

// Obstacle speed
const ObstacleSpeed = 2.2

// Player action
const (
	DinoActionIdle = iota // 0
	DinoActionRun // 1
	DinoActionJump // 2
)

// Game structure
type DinoGame struct {
	BgImage *ebiten.Image
	PlayerSetImage *ebiten.Image
	Player Player
	WidthPixels float64
	HeightPixels float64
	Obstacles []RenderItem
	CurrentState int
	Score int
	TouchButtons []TouchButton
	IsActive bool
	PlayerVelocityY float64
	NumObstaclesDeleted int
}

// Function to get a key press (first press)
// First return argument is true if there is a touch, otherwise false
// Second return argument is touch button type
func (dg *DinoGame) GetButtonJustPressed() (bool, int) {
	// Capture interaction key presses
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		fmt.Printf("Button Pressed: A\n")
		return true, ButtonA
	} else if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		fmt.Printf("Button Pressed: D\n")
		return true, ButtonD
	} else {
		touchIDs := inpututil.JustPressedTouchIDs()
		for _, id := range touchIDs {
			tx, ty := ebiten.TouchPosition(id) // Grab touch position
			// Loop through our TouchButtons
			for _, b := range dg.TouchButtons {
				if tx >= b.boundX && tx <= b.boundX + b.boundWidth &&
				ty >= b.boundY && ty <= b.boundY + b.boundHeight {
					return true, b.ButtonType
				}
			}
		}
	}
	return false, ButtonDown
}

// Helper functions
// Function: return start pixel (both dimension) of a tile insde of the player
// tileset based on what action it is
func (dg *DinoGame) GetPlayerCoord(actionType, dir, frame int) (x, y int) {
	// From the way the player tile set is set up row 2 contains all the
	// walking animation tyles. The first 6 tiles of row 2 is moving right.
	// Next 6 is moving up, next 6 is moving left, and next 6 is moving down
	// Row 1 of tileset is idle
	var playerGridX, playerGridY int
	playerGridX = (dir * 6) + frame
	if actionType == DinoActionIdle {
		playerGridY = 2
	} else if actionType == DinoActionRun || actionType == DinoActionJump {
		playerGridY = 4
	} else {
		log.Fatalf("[ERROR] Unknown player action.")
	}
	return playerGridX * tileSize, playerGridY * tileSize
}

// Function: mini-game update
func (dg *DinoGame) Update() error {
	if dg.IsActive {
		switch dg.CurrentState {
		case StateTitle: // Title scene
			dg.Player.Action = DinoActionIdle
			dg.Player.PixelX = CenterX
			dg.Player.PixelY = GroundY - 2 * tileSize
			dg.Obstacles = []RenderItem {
				{
					ID: TreeID,
					BaseY: 19 * tileSize, // base of tree
					ScreenDstX: CenterX + 13 * tileSize,
					ScreenDstY: GroundY - 2 * tileSize,
					SpriteImg: exteriorImage,
					SpriteGridSrcX: 33,
					SpriteGridSrcY: 10,
					SpriteGridDstX: 34,
					SpriteGridDstY: 12,
				},
			}
			dg.Score = 0
			dg.NumObstaclesDeleted = 0
			isPressed, buttonType := dg.GetButtonJustPressed()
			if isPressed {
				if buttonType == ButtonA { // Start Play
					dg.CurrentState = StatePlay
					dg.IsActive = true
					fmt.Printf("Start Play\n")
				} else if buttonType == ButtonD { // Exit
					dg.CurrentState = StateExit
					dg.IsActive = true
					fmt.Printf("Exit\n")
				}
			}
		case StatePlay: // Play scene
			// Player start running immediately after entering this scene
			dg.Player.Action = DinoActionRun
			// Collsion check
			for _, obs := range dg.Obstacles {
				if (dg.Player.PixelX + tileSize < obs.ScreenDstX ||
				obs.ScreenDstX + tileSize < dg.Player.PixelX ||
				dg.Player.PixelY + 2 * tileSize < obs.ScreenDstY || 
				obs.ScreenDstY + 2 * tileSize < dg.Player.PixelY) == false {
					fmt.Printf("Collision\n")
					dg.CurrentState = StateOver
				}
			}
			// Independent animation layer
			// Accumulate fractional time completely separate from moveSpeed
			dg.Player.AnimProgress += playerAnimSpeed
			if dg.Player.AnimProgress >= 1.0 {
				dg.Player.AnimProgress = 0.0
				dg.Player.AnimFrame = (dg.Player.AnimFrame + 1) % 6 // 6 frames per movement
			}
			// Move the obstacles
			var numPassedObstacles int = 0;
			for i, _ := range dg.Obstacles {
				dg.Obstacles[i].ScreenDstX -= ObstacleSpeed
				// Increment score for each passed obstacle
				if float64(dg.Obstacles[i].ScreenDstX) < dg.Player.PixelX - tileSize {
					numPassedObstacles++
				}
			}
			// If obstacle reach left handside then remove
			if len(dg.Obstacles) != 0 {
				if dg.Obstacles[0].ScreenDstX < 100 - tileSize - screenWidth / 2 {
					dg.Obstacles = slices.Delete(dg.Obstacles, 0, 1)
					dg.NumObstaclesDeleted++
				}
			}
			// Add another obstacle if the last one has reached a certain point
			if dg.Obstacles[len(dg.Obstacles)-1].ScreenDstX < 100 + 5 * tileSize {
				dg.Obstacles = append(dg.Obstacles, RenderItem {
					ID: TreeID,
					ScreenDstX: CenterX + 13 * tileSize,
					ScreenDstY: GroundY - 2 * tileSize,
					SpriteImg: exteriorImage,
					SpriteGridSrcX: 33,
					SpriteGridSrcY: 10,
					SpriteGridDstX: 34,
					SpriteGridDstY: 12,
				})
			}
			dg.Score = dg.NumObstaclesDeleted + numPassedObstacles
			// Register key presses
			isPressed, buttonType := dg.GetButtonJustPressed()
			if isPressed { // Make character jump?
				if buttonType == ButtonA {
					dg.CurrentState = StatePlay
					dg.IsActive = true
					fmt.Printf("Keep Play\n")
					fmt.Printf("Distance from ground: %f\n", math.Abs(dg.Player.PixelY + 2 * tileSize - GroundY))
					// Make character jump by descreasing velocity in Y direction
					dg.PlayerVelocityY = -12.0
					fmt.Printf("Number of obstacles in queue: %d\n", len(dg.Obstacles))
				} else if buttonType == ButtonD { // Exit
					dg.CurrentState = StateExit
					dg.IsActive = true
					fmt.Printf("Exit\n")
				}
			}
			// Can't fall through the ground
			if dg.Player.PixelY + 2 * tileSize == GroundY {
				if dg.PlayerVelocityY > 0 {
					dg.PlayerVelocityY = 0
				}
			} else if dg.Player.PixelY + 2 * tileSize < GroundY {
				// Gravity effect
				dg.PlayerVelocityY += Gravity
			} else if dg.Player.PixelY + 2 * tileSize > GroundY {
				dg.PlayerVelocityY = -1
			}
			dg.Player.PixelY += dg.PlayerVelocityY

		case StateOver: // Game over scene
			dg.Player.Action = DinoActionJump
			isPressed, buttonType := dg.GetButtonJustPressed()
			if isPressed {
				if buttonType == ButtonA { // Start Play again (go back to title scene)
					dg.CurrentState = StateTitle
					dg.IsActive = true
					fmt.Printf("Play again\n")
				} else if buttonType == ButtonD { // Exit
					dg.CurrentState = StateExit
					dg.IsActive = true
					fmt.Printf("Exit\n")
				}
			}
		case StateExit: // Exit game
			dg.CurrentState = StateTitle // return to title scene after exit
			dg.IsActive = false
			fmt.Printf("StateExit\n")
		}
	}
	return nil
}

// Mini-game Method:  Draw touch buttons for mobile
func (dg *DinoGame) DrawTouchButtons(screen *ebiten.Image) {
	// Draw each button
	for _, b := range dg.TouchButtons {
		vector.DrawFilledCircle(
			screen,
			float32(b.boundX + b.boundWidth / 2.0),
			float32(b.boundY + b.boundHeight / 2.0),
			float32(b.boundWidth / 2.0),
			color.NRGBA{255, 255, 255, 128},
			false,
		)
		if b.ButtonType == ButtonA {
			ebitenutil.DebugPrintAt(screen, "A", b.boundX + b.boundWidth / 2.0 - 3,
				b.boundY + b.boundHeight / 2.0 - 8)
		} else if b.ButtonType == ButtonD {
			ebitenutil.DebugPrintAt(screen, "D", b.boundX + b.boundWidth / 2.0 - 3,
				b.boundY + b.boundHeight / 2.0 - 8)
		}
	}
}

// Game Method: Draw Player (called in Draw method)
func (dg *DinoGame) DrawPlayer(screen *ebiten.Image) {
	// Camera position
	fixedCamX := 100.0
	fixedCamY := 100.0
	var camMatrix ebiten.GeoM
	camMatrix.Translate(-fixedCamX + screenWidth / 2, -fixedCamY + screenHeight / 2)
	// Get player start coordinate inside of the player tileset
	pSrcX, pSrcY := dg.GetPlayerCoord(dg.Player.Action, dg.Player.Dir, dg.Player.AnimFrame)
	// Get the player rectangle image coordinate
	pRect := image.Rect(pSrcX, pSrcY, pSrcX + tileSize, pSrcY + 2 * tileSize)
	// Extract the subimage from the player tileset
	playerSprite := dg.PlayerSetImage.SubImage(pRect).(*ebiten.Image)
	// Set up the draw options (start point of the draw call)
	popts := &ebiten.DrawImageOptions{}
	popts.GeoM.Translate(dg.Player.PixelX, dg.Player.PixelY)
	popts.GeoM.Concat(camMatrix)
	// Draw call
	screen.DrawImage(playerSprite, popts)
}

// Game Method: Draw Player (called in Draw method)
func (dg *DinoGame) DrawObstacle(screen *ebiten.Image) {
	// Camera position
	fixedCamX := 100.0
	fixedCamY := 100.0
	var camMatrix ebiten.GeoM
	camMatrix.Translate(-fixedCamX + screenWidth / 2, -fixedCamY + screenHeight / 2)
	for _, obstacle := range dg.Obstacles {
		// Get player start coordinate inside of the player tileset
		//pSrcX, pSrcY, pDstX, pDstY := GetExteriorItemCoord(obstacle.ID)
		pSrcX := obstacle.SpriteGridSrcX * tileSize
		pSrcY := obstacle.SpriteGridSrcY * tileSize
		pDstX := obstacle.SpriteGridDstX * tileSize
		pDstY := obstacle.SpriteGridDstY * tileSize
		// Get the player rectangle image coordinate
		pRect := image.Rect(pSrcX, pSrcY, pDstX, pDstY)
		// Extract the subimage from the player tileset
		obstacleSprite := obstacle.SpriteImg.SubImage(pRect).(*ebiten.Image)
		// Set up the draw options (start point of the draw call)
		popts := &ebiten.DrawImageOptions{}
		popts.GeoM.Translate(float64(obstacle.ScreenDstX), float64(obstacle.ScreenDstY))
		popts.GeoM.Concat(camMatrix)
		// Draw call
		screen.DrawImage(obstacleSprite, popts)
	}
}

// Game Method: Draw background (called in Draw method)
func (dg *DinoGame) DrawBackground(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0.0, 0.0)
	screen.DrawImage(dg.BgImage, op)
}

// Function: mini-game draw call
func (dg *DinoGame) Draw(screen *ebiten.Image) {
	// Draw background
	dg.DrawBackground(screen)

	// Draw text on screen for now
	var stateStr string
	switch dg.CurrentState {
	case StateTitle:
		stateStr = "MINI DINO\nPress A to Play"
	case StatePlay:
		stateStr = "MINI DINO\nPlaying..."
	case StateOver:
		stateStr = "MINI DINO\nGame Over"
	}
	
	// Draw Obstacle
	dg.DrawObstacle(screen)
	// Draw Player
	dg.DrawPlayer(screen)
	// Draw virtual buttons
	dg.DrawTouchButtons(screen)
	// Display Score
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", dg.Score), 320, 30)
	// Display Tick per second
	ebitenutil.DebugPrintAt(screen, stateStr, 30, 30)
}
