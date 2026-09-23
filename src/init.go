package main

import (
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
	exteriorImage = loadEmbeddedImage(exteriorByteData)
	parkingLotBg = loadEmbeddedImage(parkingLotByteData)
	courtYardBg = loadEmbeddedImage(courtYardByteData)
	playerSetImage = loadEmbeddedImage(playerByteData)

	// Build scene registry index container map
	//gameScenes := make(map[int]*Scene)
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
				BaseY: 20 * tileSize, // base of tree
				ScreenDstX: 3 * tileSize,
				ScreenDstY: 19 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 16 * tileSize, // base of tree
				ScreenDstX: 3 * tileSize,
				ScreenDstY: 15 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 8 * tileSize, // base of tree
				ScreenDstX: 4 * tileSize,
				ScreenDstY: 7 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 7 * tileSize, // base of tree
				ScreenDstX: 6 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 8 * tileSize, // base of tree
				ScreenDstX: 8 * tileSize,
				ScreenDstY: 7 * tileSize,
				SpriteImg: exteriorImage,
			},
			{
				ID: TreeID,
				BaseY: 7 * tileSize, // base of tree
				ScreenDstX: 10 * tileSize,
				ScreenDstY: 6 * tileSize,
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
	padX := 40 // Bottom left cluster placement
	padY := 180 / 2
	mobileButtons = []TouchButton {
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
}
