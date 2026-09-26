package main

import (
	//"cmp"
	//"slices"
	//"math"
	_ "embed"
	"fmt"
	//"image"
	"image/color"
	_ "image/png"
	//"log"

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

// Game structure
type FlappyGame struct {
	BgImage *ebiten.Image
	WidthPixels float64
	HeightPixels float64
	Obstacle RenderItem
	CurrentState int
	Score int
	TouchButtons []TouchButton
	IsActive bool
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

// Function: mini-game update
func (fg *FlappyGame) Update() error {
	if fg.IsActive {
		switch fg.CurrentState {
		case StateTitle: // Title scene
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
			isPressed, buttonType := fg.GetButtonJustPressed()
			if isPressed { // Make character jump?
				if buttonType == ButtonA {
					fg.CurrentState = StateOver
					fg.IsActive = true
					fmt.Printf("Keep Play\n")
				} else if buttonType == ButtonD { // Exit
					fg.CurrentState = StateExit
					fg.IsActive = true
					fmt.Printf("Exit\n")
				}
			}
		case StateOver: // Game over scene
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
		stateStr = "MINI FLAPPY TITLE SCENE"
	case StatePlay:
		stateStr = "MINI FLAPPY PLAY SCENE"
	case StateOver:
		stateStr = "MINI FLAPPY OVER SCENE"
	}
	
	// Draw virtual buttons
	fg.DrawTouchButtons(screen)
	// Display Tick per second
	ebitenutil.DebugPrintAt(screen, stateStr, 30, screenHeight / 2)
}
