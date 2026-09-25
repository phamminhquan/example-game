package main

import (
	"sync"
	"fmt"
	"time"
	"encoding/csv"
	"io"
	"strconv"
	_ "embed"
	"bytes"
	"image"
	_ "image/png"
	"log"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
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
// Global variable storing all the scenes
var gameScenes = make(map[int]*Scene)

// Global variable storing the images of the scenes
var exteriorImage *ebiten.Image
var parkingLotBg *ebiten.Image
var courtYardBg *ebiten.Image
var playerSetImage *ebiten.Image

// Global variable storing the virtual touch buttons info
var mobileButtons []TouchButton

// Global variables storing the collision csv and concurrency wait group
var Time time.Time
var parkingLotCollision [][]int
var courtYardCollision [][]int
var wg sync.WaitGroup

// Global variable storing interactions
var gameInteractions []Interaction

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

// Init function is executed automatically before main
func init() {
	// Read in assets
	Time = time.Now()
	wg.Add(1)
	go func() {
		defer wg.Done()
		exteriorImage = loadEmbeddedImage(exteriorByteData)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		parkingLotBg = loadEmbeddedImage(parkingLotByteData)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		courtYardBg = loadEmbeddedImage(courtYardByteData)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		playerSetImage = loadEmbeddedImage(playerByteData)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		parkingLotCollision = loadCollisionCsv(parkingLotCsvStr)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		courtYardCollision = loadCollisionCsv(courtYardCsvStr)
	}()
	wg.Wait()
	fmt.Printf("Time elasped: %s\n", time.Since(Time))

	// Build scene registry index container map
	// Scene: Parking Lot
	gameScenes[SceneParkingLot] = &Scene {
		ID: SceneParkingLot,
		BgImage: parkingLotBg,
		Collision: parkingLotCollision,
		WidthPixels: float64(parkingLotBg.Bounds().Dx()),
		HeightPixels: float64(parkingLotBg.Bounds().Dy()),
		RenderItems: []RenderItem {
			{
				ID: TreeID,
				BaseY: 24 * tileSize, // base of tree
				ScreenDstX: 2 * tileSize,
				ScreenDstY: 23 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 24 * tileSize, // base of tree
				ScreenDstX: 10 * tileSize,
				ScreenDstY: 23 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 22 * tileSize, // base of tree
				ScreenDstX: 17 * tileSize,
				ScreenDstY: 21 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 14 * tileSize, // base of tree
				ScreenDstX: 6 * tileSize,
				ScreenDstY: 13 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 14 * tileSize, // base of tree
				ScreenDstX: 14 * tileSize,
				ScreenDstY: 13 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 12 * tileSize, // base of tree
				ScreenDstX: 17 * tileSize,
				ScreenDstY: 11 * tileSize,
				SpriteImg: exteriorImage,
			},
		},
		WarpTriggers: []WarpTrigger {
			{
				DespawnX: 1,
				DespawnY: 1,
				SpawnX: 4,
				SpawnY: 20,
				TargetScene: SceneCourtYard,
			},
			{
				DespawnX: 2,
				DespawnY: 1,
				SpawnX: 5,
				SpawnY: 20,
				TargetScene: SceneCourtYard,
			},
		},
		InteractionTriggers: []InteractionTrigger {
			{
				GridX: 3,
				GridY: 7,
				PlayerDir: DirUp,
				InteractionID: 0,
			},
			{
				GridX: 14,
				GridY: 21,
				PlayerDir: DirRight,
				InteractionID: 1,
			},
			{
				GridX: 14,
				GridY: 22,
				PlayerDir: DirRight,
				InteractionID: 1,
			},
			{
				GridX: 15,
				GridY: 20,
				PlayerDir: DirDown,
				InteractionID: 1,
			},
			{
				GridX: 16,
				GridY: 20,
				PlayerDir: DirDown,
				InteractionID: 1,
			},
		},
	}
	
	// Scene: Court Yard
	gameScenes[SceneCourtYard] = &Scene {
		ID: SceneCourtYard,
		BgImage: courtYardBg,
		Collision: courtYardCollision,
		WidthPixels: float64(courtYardBg.Bounds().Dx()),
		HeightPixels: float64(courtYardBg.Bounds().Dy()),
		RenderItems: []RenderItem {
			{
				ID: TreeID,
				BaseY: 19 * tileSize, // base of tree
				ScreenDstX: 3 * tileSize,
				ScreenDstY: 18 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 15 * tileSize, // base of tree
				ScreenDstX: 3 * tileSize,
				ScreenDstY: 14 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 7 * tileSize, // base of tree
				ScreenDstX: 4 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 6 * tileSize, // base of tree
				ScreenDstX: 6 * tileSize,
				ScreenDstY: 5 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 7 * tileSize, // base of tree
				ScreenDstX: 8 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 6 * tileSize, // base of tree
				ScreenDstX: 10 * tileSize,
				ScreenDstY: 5 * tileSize,
				SpriteImg: exteriorImage,
			},
		},
		WarpTriggers: []WarpTrigger {
			{
				DespawnX: 4,
				DespawnY: 21,
				SpawnX: 1,
				SpawnY: 2,
				TargetScene: SceneParkingLot,
			},
			{
				DespawnX: 5,
				DespawnY: 21,
				SpawnX: 2,
				SpawnY: 2,
				TargetScene: SceneParkingLot,
			},
		},
	}
	
	// Initialize TouchButtons bounding box for mobile
	buttonSize := 32
	movementPadX := 40 // Bottom left cluster placement
	movementPadY := 180 / 2
	interactPadX := screenWidth - 40 // Bottom right placement
	interactPadY := 180 / 2
	mobileButtons = []TouchButton {
		{ // A button
			ButtonType: ButtonA,
			boundX: interactPadX,
			boundY: interactPadY,
			boundWidth: buttonSize,
			boundHeight: buttonSize,
		},
		{ // D button
			ButtonType: ButtonD,
			boundX: interactPadX - 2 * buttonSize,
			boundY: interactPadY,
			boundWidth: buttonSize,
			boundHeight: buttonSize,
		},
		{ // Up button
			ButtonType: ButtonUp,
			boundX: movementPadX,
			boundY: movementPadY - buttonSize,
			boundWidth: buttonSize,
			boundHeight: buttonSize,
			Dir: DirUp,
		},
		{ // Down button
			ButtonType: ButtonDown,
			boundX: movementPadX,
			boundY: movementPadY + buttonSize,
			boundWidth: buttonSize,
			boundHeight: buttonSize,
			Dir: DirDown,
		},
		{ // Left button
			ButtonType: ButtonLeft,
			boundX: movementPadX - buttonSize,
			boundY: movementPadY,
			boundWidth: buttonSize,
			boundHeight: buttonSize,
			Dir: DirLeft,
		},
		{ // Right button
			ButtonType: ButtonRight,
			boundX: movementPadX + buttonSize,
			boundY: movementPadY,
			boundWidth: buttonSize,
			boundHeight: buttonSize,
			Dir: DirRight,
		},
	}

	// Initialize interactions
	gameInteractions = []Interaction {
		{
			InteractionID: 0,
			InteractionType: InteractionTypeConversation,
			InteractionStates: 1,
			InteractionText: []string {
				"Hello World!",
			},
		},
		{
			InteractionID: 1,
			InteractionType: InteractionTypeConversation,
			InteractionStates: 2,
			InteractionText: []string {
				"Whose car is this?",
				"It's a Florida plate.",
			},
		},
	}
}
