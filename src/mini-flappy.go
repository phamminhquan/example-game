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
	FlappyActionIdle = iota // 0
	FlappyActionRun // 1
	FlappyActionJump // 2
)

// Game structure
type FlappyGame struct {
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
func (fg *FlappyGame) GetButtonJustPressed() (bool, int) {
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
			for _, b := range fg.TouchButtons {
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
func (fg *FlappyGame) GetPlayerCoord(actionType, dir, frame int) (x, y int) {
	// From the way the player tile set is set up row 2 contains all the
	// walking animation tyles. The first 6 tiles of row 2 is moving right.
	// Next 6 is moving up, next 6 is moving left, and next 6 is moving down
	// Row 1 of tileset is idle
	var playerGridX, playerGridY int
	playerGridX = (dir * 6) + frame
	if actionType == FlappyActionIdle {
		playerGridY = 2
	} else if actionType == FlappyActionRun || actionType == FlappyActionJump {
		playerGridY = 4
	} else {
		log.Fatalf("[ERROR] Unknown player action.")
	}
	return playerGridX * tileSize, playerGridY * tileSize
}

// Function: mini-game update
func (fg *FlappyGame) Update() error {
	if fg.IsActive {
		switch fg.CurrentState {
		case StateTitle: // Title scene
			fg.Player.Action = FlappyActionIdle
			fg.Player.PixelX = CenterX
			fg.Player.PixelY = GroundY - 2 * tileSize
			fg.Obstacles = []RenderItem {
				{
					ID: TreeID,
					BaseY: 19 * tileSize, // base of tree
					ScreenDstX: CenterX + 13 * tileSize,
					ScreenDstY: GroundY - 2 * tileSize,
					SpriteImg: exteriorImage,
				},
			}
			fg.Score = 0
			fg.NumObstaclesDeleted = 0
			isPressed, buttonType := fg.GetButtonJustPressed()
			if isPressed {
				if buttonType == ButtonA { // Start Play
					fg.CurrentState = StatePlay
					fg.IsActive = true
					fmt.Printf("Start Play\n")
				} else if buttonType == ButtonD { // Exit
					fg.CurrentState = StateExit
					fg.IsActive = true
					fmt.Printf("Exit\n")
				}
			}
		case StatePlay: // Play scene
			// Player start running immediately after entering this scene
			fg.Player.Action = FlappyActionRun
			// Collsion check
			for _, obs := range fg.Obstacles {
				if (fg.Player.PixelX + tileSize < obs.ScreenDstX ||
				obs.ScreenDstX + tileSize < fg.Player.PixelX ||
				fg.Player.PixelY + 2 * tileSize < obs.ScreenDstY || 
				obs.ScreenDstY + 2 * tileSize < fg.Player.PixelY) == false {
					fmt.Printf("Collision\n")
					fg.CurrentState = StateOver
				}
			}
			// Independent animation layer
			// Accumulate fractional time completely separate from moveSpeed
			fg.Player.AnimProgress += playerAnimSpeed
			if fg.Player.AnimProgress >= 1.0 {
				fg.Player.AnimProgress = 0.0
				fg.Player.AnimFrame = (fg.Player.AnimFrame + 1) % 6 // 6 frames per movement
			}
			// Move the obstacles
			var numPassedObstacles int = 0;
			for i, _ := range fg.Obstacles {
				fg.Obstacles[i].ScreenDstX -= ObstacleSpeed
				// Increment score for each passed obstacle
				if float64(fg.Obstacles[i].ScreenDstX) < fg.Player.PixelX - tileSize {
					numPassedObstacles++
				}
			}
			// If obstacle reach left handside then remove
			if len(fg.Obstacles) != 0 {
				if fg.Obstacles[0].ScreenDstX < 100 - tileSize - screenWidth / 2 {
					fg.Obstacles = slices.Delete(fg.Obstacles, 0, 1)
					fg.NumObstaclesDeleted++
				}
			}
			// Add another obstacle if the last one has reached a certain point
			if fg.Obstacles[len(fg.Obstacles)-1].ScreenDstX < 100 + 5 * tileSize {
				fg.Obstacles = append(fg.Obstacles, RenderItem {
					ID: TreeID,
					ScreenDstX: CenterX + 13 * tileSize,
					ScreenDstY: GroundY - 2 * tileSize,
					SpriteImg: exteriorImage,
				})
			}
			fg.Score = fg.NumObstaclesDeleted + numPassedObstacles
			// Register key presses
			isPressed, buttonType := fg.GetButtonJustPressed()
			if isPressed { // Make character jump?
				if buttonType == ButtonA {
					fg.CurrentState = StatePlay
					fg.IsActive = true
					fmt.Printf("Keep Play\n")
					fmt.Printf("Distance from ground: %f\n", math.Abs(fg.Player.PixelY + 2 * tileSize - GroundY))
					// Make character jump by descreasing velocity in Y direction
					fg.PlayerVelocityY = -12.0
					fmt.Printf("Number of obstacles in queue: %d\n", len(fg.Obstacles))
				} else if buttonType == ButtonD { // Exit
					fg.CurrentState = StateExit
					fg.IsActive = true
					fmt.Printf("Exit\n")
				}
			}
			// Can't fall through the ground
			if fg.Player.PixelY + 2 * tileSize == GroundY {
				if fg.PlayerVelocityY > 0 {
					fg.PlayerVelocityY = 0
				}
			} else if fg.Player.PixelY + 2 * tileSize < GroundY {
				// Gravity effect
				fg.PlayerVelocityY += Gravity
			} else if fg.Player.PixelY + 2 * tileSize > GroundY {
				fg.PlayerVelocityY = -1
			}
			fg.Player.PixelY += fg.PlayerVelocityY

		case StateOver: // Game over scene
			fg.Player.Action = FlappyActionJump
			isPressed, buttonType := fg.GetButtonJustPressed()
			if isPressed {
				if buttonType == ButtonA { // Start Play again (go back to title scene)
					fg.CurrentState = StateTitle
					fg.IsActive = true
					fmt.Printf("Play again\n")
				} else if buttonType == ButtonD { // Exit
					fg.CurrentState = StateExit
					fg.IsActive = true
					fmt.Printf("Exit\n")
				}
			}
		case StateExit: // Exit game
			fg.CurrentState = StateTitle // return to title scene after exit
			fg.IsActive = false
			fmt.Printf("StateExit\n")
		}
	}
	return nil
}

// Mini-game Method:  Draw touch buttons for mobile
func (fg *FlappyGame) DrawTouchButtons(screen *ebiten.Image) {
	// Draw each button
	for _, b := range fg.TouchButtons {
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
func (fg *FlappyGame) DrawPlayer(screen *ebiten.Image) {
	// Camera position
	fixedCamX := 100.0
	fixedCamY := 100.0
	var camMatrix ebiten.GeoM
	camMatrix.Translate(-fixedCamX + screenWidth / 2, -fixedCamY + screenHeight / 2)
	// Get player start coordinate inside of the player tileset
	pSrcX, pSrcY := fg.GetPlayerCoord(fg.Player.Action, fg.Player.Dir, fg.Player.AnimFrame)
	// Get the player rectangle image coordinate
	pRect := image.Rect(pSrcX, pSrcY, pSrcX + tileSize, pSrcY + 2 * tileSize)
	// Extract the subimage from the player tileset
	playerSprite := fg.PlayerSetImage.SubImage(pRect).(*ebiten.Image)
	// Set up the draw options (start point of the draw call)
	popts := &ebiten.DrawImageOptions{}
	popts.GeoM.Translate(fg.Player.PixelX, fg.Player.PixelY)
	popts.GeoM.Concat(camMatrix)
	// Draw call
	screen.DrawImage(playerSprite, popts)
}

// Game Method: Draw Player (called in Draw method)
func (fg *FlappyGame) DrawObstacle(screen *ebiten.Image) {
	// Camera position
	fixedCamX := 100.0
	fixedCamY := 100.0
	var camMatrix ebiten.GeoM
	camMatrix.Translate(-fixedCamX + screenWidth / 2, -fixedCamY + screenHeight / 2)
	for _, obstacle := range fg.Obstacles {
		// Get player start coordinate inside of the player tileset
		pSrcX, pSrcY, pDstX, pDstY := GetExteriorItemCoord(obstacle.ID)
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

// Function: mini-game draw call
func (fg *FlappyGame) Draw(screen *ebiten.Image) {
	// Blackout background for now
	screen.Fill(color.NRGBA{0, 0, 0, 255})

	// Draw text on screen for now
	var stateStr string
	switch fg.CurrentState {
	case StateTitle:
		stateStr = "MINI FLAPPY\nPress A to Play"
	case StatePlay:
		stateStr = "MINI FLAPPY\nPlaying..."

	case StateOver:
		stateStr = "MINI FLAPPY\nGame Over"
	}
	
	// Draw Obstacle
	fg.DrawObstacle(screen)
	// Draw Player
	fg.DrawPlayer(screen)
	// Draw virtual buttons
	fg.DrawTouchButtons(screen)
	// Display Score
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", fg.Score), 320, 30)
	// Display Tick per second
	ebitenutil.DebugPrintAt(screen, stateStr, 30, 30)
}
