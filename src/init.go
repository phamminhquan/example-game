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
//go:embed assets/interior-sprites.png
var interiorByteData []byte
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
//go:embed assets/mini-dino.png
var miniDinoByteData []byte
//go:embed assets/reception.png
var receptionByteData []byte
//go:embed assets/reception-collision.csv
var receptionCsvStr string
//go:embed assets/cubicle.png
var cubicleByteData []byte
//go:embed assets/cubicle-collision.csv
var cubicleCsvStr string
//go:embed assets/adam.png
var npc1ByteData []byte
//go:embed assets/alex.png
var npc2ByteData []byte
// Global variable storing all the scenes
var gameScenes = make(map[int]*Scene)

// Global variable storing the images of the scenes
var exteriorImage *ebiten.Image
var interiorImage *ebiten.Image
var parkingLotBg *ebiten.Image
var courtYardBg *ebiten.Image
var playerSetImage *ebiten.Image
var miniDinoBg *ebiten.Image
var receptionBg *ebiten.Image
var cubicleBg *ebiten.Image
var npc1SetImage *ebiten.Image
var npc2SetImage *ebiten.Image

// Global variable storing the virtual touch buttons info
var mobileButtons []TouchButton

// Global variables storing the collision csv and concurrency wait group
var Time time.Time
var parkingLotCollision [][]int
var courtYardCollision [][]int
var receptionCollision [][]int
var cubicleCollision [][]int
var wg sync.WaitGroup

// Global variable storing interactions
var gameInteractions []Interaction

// Global variable storing mini dino game
var miniDinoGame DinoGame

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
		interiorImage = loadEmbeddedImage(interiorByteData)
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
		miniDinoBg = loadEmbeddedImage(miniDinoByteData)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		receptionBg = loadEmbeddedImage(receptionByteData)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		cubicleBg = loadEmbeddedImage(cubicleByteData)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		npc1SetImage = loadEmbeddedImage(npc1ByteData)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		npc2SetImage = loadEmbeddedImage(npc2ByteData)
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
	wg.Add(1)
	go func() {
		defer wg.Done()
		receptionCollision = loadCollisionCsv(receptionCsvStr)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		cubicleCollision = loadCollisionCsv(cubicleCsvStr)
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
				SpriteGridSrcX: 33,
				SpriteGridSrcY: 10,
				SpriteGridDstX: 34,
				SpriteGridDstY: 12,
			},
			{
				ID: TreeID,
				BaseY: 24 * tileSize, // base of tree
				ScreenDstX: 10 * tileSize,
				ScreenDstY: 23 * tileSize,
				SpriteImg: exteriorImage,
				SpriteGridSrcX: 33,
				SpriteGridSrcY: 10,
				SpriteGridDstX: 34,
				SpriteGridDstY: 12,
			},
			{
				ID: TreeID,
				BaseY: 22 * tileSize, // base of tree
				ScreenDstX: 17 * tileSize,
				ScreenDstY: 21 * tileSize,
				SpriteImg: exteriorImage,
				SpriteGridSrcX: 33,
				SpriteGridSrcY: 10,
				SpriteGridDstX: 34,
				SpriteGridDstY: 12,
			},
			{
				ID: TreeID,
				BaseY: 14 * tileSize, // base of tree
				ScreenDstX: 6 * tileSize,
				ScreenDstY: 13 * tileSize,
				SpriteImg: exteriorImage,
				SpriteGridSrcX: 33,
				SpriteGridSrcY: 10,
				SpriteGridDstX: 34,
				SpriteGridDstY: 12,
			},
			{
				ID: TreeID,
				BaseY: 14 * tileSize, // base of tree
				ScreenDstX: 14 * tileSize,
				ScreenDstY: 13 * tileSize,
				SpriteImg: exteriorImage,
				SpriteGridSrcX: 33,
				SpriteGridSrcY: 10,
				SpriteGridDstX: 34,
				SpriteGridDstY: 12,
			},
			{
				ID: TreeID,
				BaseY: 12 * tileSize, // base of tree
				ScreenDstX: 17 * tileSize,
				ScreenDstY: 11 * tileSize,
				SpriteImg: exteriorImage,
				SpriteGridSrcX: 33,
				SpriteGridSrcY: 10,
				SpriteGridDstX: 34,
				SpriteGridDstY: 12,
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
				SpriteGridSrcX: 33,
				SpriteGridSrcY: 10,
				SpriteGridDstX: 34,
				SpriteGridDstY: 12,
			},
			{
				ID: TreeID,
				BaseY: 15 * tileSize, // base of tree
				ScreenDstX: 3 * tileSize,
				ScreenDstY: 14 * tileSize,
				SpriteImg: exteriorImage,
				SpriteGridSrcX: 33,
				SpriteGridSrcY: 10,
				SpriteGridDstX: 34,
				SpriteGridDstY: 12,
			},
			{
				ID: TreeID,
				BaseY: 7 * tileSize, // base of tree
				ScreenDstX: 4 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: exteriorImage,
				SpriteGridSrcX: 33,
				SpriteGridSrcY: 10,
				SpriteGridDstX: 34,
				SpriteGridDstY: 12,
			},
			{
				ID: TreeID,
				BaseY: 6 * tileSize, // base of tree
				ScreenDstX: 6 * tileSize,
				ScreenDstY: 5 * tileSize,
				SpriteImg: exteriorImage,
				SpriteGridSrcX: 33,
				SpriteGridSrcY: 10,
				SpriteGridDstX: 34,
				SpriteGridDstY: 12,
			},
			{
				ID: TreeID,
				BaseY: 7 * tileSize, // base of tree
				ScreenDstX: 8 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: exteriorImage,
				SpriteGridSrcX: 33,
				SpriteGridSrcY: 10,
				SpriteGridDstX: 34,
				SpriteGridDstY: 12,
			},
			{
				ID: TreeID,
				BaseY: 6 * tileSize, // base of tree
				ScreenDstX: 10 * tileSize,
				ScreenDstY: 5 * tileSize,
				SpriteImg: exteriorImage,
				SpriteGridSrcX: 33,
				SpriteGridSrcY: 10,
				SpriteGridDstX: 34,
				SpriteGridDstY: 12,
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
			{
				DespawnX: 1,
				DespawnY: 10,
				SpawnX: 12,
				SpawnY: 20,
				TargetScene: SceneReception,
			},
			{
				DespawnX: 1,
				DespawnY: 11,
				SpawnX: 12,
				SpawnY: 21,
				TargetScene: SceneReception,
			},
		},
	}
	
	// Scene: Reception
	gameScenes[SceneReception] = &Scene {
		ID: SceneReception,
		BgImage: receptionBg,
		Collision: receptionCollision,
		WidthPixels: float64(receptionBg.Bounds().Dx()),
		HeightPixels: float64(receptionBg.Bounds().Dy()),
		RenderItems: []RenderItem {
			{
				ID: GlassDoorID,
				BaseY: 13 * tileSize,
				ScreenDstX: 8 * tileSize,
				ScreenDstY: 12 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 3,
				SpriteGridSrcY: 28,
				SpriteGridDstX: 4,
				SpriteGridDstY: 30,
			},
			{
				ID: GlassDoorID,
				BaseY: 13 * tileSize,
				ScreenDstX: 11 * tileSize,
				ScreenDstY: 12 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 3,
				SpriteGridSrcY: 28,
				SpriteGridDstX: 4,
				SpriteGridDstY: 30,
			},
			{
				ID: ChairID,
				BaseY: 9 * tileSize,
				ScreenDstX: 4 * tileSize,
				ScreenDstY: 9 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 3,
				SpriteGridSrcY: 74,
				SpriteGridDstX: 4,
				SpriteGridDstY: 75,
			},
			{
				ID: ChairID,
				BaseY: 9 * tileSize,
				ScreenDstX: 14 * tileSize,
				ScreenDstY: 9 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 3,
				SpriteGridSrcY: 74,
				SpriteGridDstX: 4,
				SpriteGridDstY: 75,
			},
			{
				ID: ChairID,
				BaseY: 9 * tileSize,
				ScreenDstX: 5 * tileSize,
				ScreenDstY: 7 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 7,
				SpriteGridSrcY: 76,
				SpriteGridDstX: 9,
				SpriteGridDstY: 79,
			},
			{
				ID: ChairID,
				BaseY: 10 * tileSize,
				ScreenDstX: 12 * tileSize,
				ScreenDstY: 8 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 9,
				SpriteGridSrcY: 76,
				SpriteGridDstX: 11,
				SpriteGridDstY: 79,
			},
			{
				ID: ChairID,
				BaseY: 7 * tileSize,
				ScreenDstX: 4 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 1,
				SpriteGridSrcY: 72,
				SpriteGridDstX: 4,
				SpriteGridDstY: 74,
			},
			{
				ID: ChairID,
				BaseY: 8 * tileSize,
				ScreenDstX: 12 * tileSize,
				ScreenDstY: 7 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 1,
				SpriteGridSrcY: 72,
				SpriteGridDstX: 4,
				SpriteGridDstY: 74,
			},
			{
				ID: PlantPotID,
				BaseY: 4 * tileSize,
				ScreenDstX: 4 * tileSize,
				ScreenDstY: 3 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 12,
				SpriteGridSrcY: 45,
				SpriteGridDstX: 13,
				SpriteGridDstY: 47,
			},
			{
				ID: DeskID,
				BaseY: 4 * tileSize,
				ScreenDstX: 5 * tileSize,
				ScreenDstY: 4 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 0,
				SpriteGridSrcY: 8,
				SpriteGridDstX: 2,
				SpriteGridDstY: 9,
			},
			{
				ID: DeskID,
				BaseY: 4 * tileSize,
				ScreenDstX: 8 * tileSize,
				ScreenDstY: 4 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 1,
				SpriteGridSrcY: 8,
				SpriteGridDstX: 3,
				SpriteGridDstY: 9,
			},
			{
				ID: MonitorID,
				BaseY: 4 * tileSize,
				ScreenDstX: 7 * tileSize,
				ScreenDstY: 3 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 3,
				SpriteGridSrcY: 8,
				SpriteGridDstX: 4,
				SpriteGridDstY: 10,
			},
		},
		WarpTriggers: []WarpTrigger {
			{
				DespawnX: 13,
				DespawnY: 20,
				SpawnX: 2,
				SpawnY: 10,
				TargetScene: SceneCourtYard,
			},
			{
				DespawnX: 13,
				DespawnY: 21,
				SpawnX: 2,
				SpawnY: 11,
				TargetScene: SceneCourtYard,
			},
			{
				DespawnX: 1,
				DespawnY: 20,
				SpawnX: 28,
				SpawnY: 7,
				TargetScene: SceneCubicle,
			},
			{
				DespawnX: 1,
				DespawnY: 21,
				SpawnX: 28,
				SpawnY: 8,
				TargetScene: SceneCubicle,
			},
		},
	}

	// Scene: Cubicle
	gameScenes[SceneCubicle] = &Scene {
		ID: SceneCubicle,
		BgImage: cubicleBg,
		Collision: cubicleCollision,
		WidthPixels: float64(cubicleBg.Bounds().Dx()),
		HeightPixels: float64(cubicleBg.Bounds().Dy()),
		RenderItems: []RenderItem {
			{
				ID: DeskID,
				BaseY: 2 * tileSize,
				ScreenDstX: 3 * tileSize,
				ScreenDstY: 2 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 0,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 3,
				SpriteGridDstY: 6,
			},
			{
				ID: DeskID,
				BaseY: 2 * tileSize,
				ScreenDstX: 6 * tileSize,
				ScreenDstY: 2 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 0,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 3,
				SpriteGridDstY: 6,
			},
			{
				ID: DeskID,
				BaseY: 2 * tileSize,
				ScreenDstX: 9 * tileSize,
				ScreenDstY: 2 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 0,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 3,
				SpriteGridDstY: 6,
			},
			{
				ID: DeskID,
				BaseY: 2 * tileSize,
				ScreenDstX: 18 * tileSize,
				ScreenDstY: 2 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 0,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 3,
				SpriteGridDstY: 6,
			},
			{
				ID: DeskID,
				BaseY: 2 * tileSize,
				ScreenDstX: 21 * tileSize,
				ScreenDstY: 2 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 0,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 3,
				SpriteGridDstY: 6,
			},
			{
				ID: DeskID,
				BaseY: 2 * tileSize,
				ScreenDstX: 24 * tileSize,
				ScreenDstY: 2 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 0,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 3,
				SpriteGridDstY: 6,
			},
			{
				ID: DeskID,
				BaseY: 6 * tileSize,
				ScreenDstX: 3 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 0,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 1,
				SpriteGridDstY: 6,
			},
			{
				ID: DeskID,
				BaseY: 6 * tileSize,
				ScreenDstX: 6 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 0,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 1,
				SpriteGridDstY: 6,
			},
			{
				ID: DeskID,
				BaseY: 6 * tileSize,
				ScreenDstX: 9 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 0,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 1,
				SpriteGridDstY: 6,
			},
			{
				ID: DeskID,
				BaseY: 6 * tileSize,
				ScreenDstX: 18 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 0,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 1,
				SpriteGridDstY: 6,
			},
			{
				ID: DeskID,
				BaseY: 6 * tileSize,
				ScreenDstX: 21 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 0,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 1,
				SpriteGridDstY: 6,
			},
			{
				ID: DeskID,
				BaseY: 6 * tileSize,
				ScreenDstX: 24 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 0,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 1,
				SpriteGridDstY: 6,
			},
			{
				ID: DeskID,
				BaseY: 6 * tileSize,
				ScreenDstX: 5 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 2,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 3,
				SpriteGridDstY: 6,
			},
			{
				ID: DeskID,
				BaseY: 6 * tileSize,
				ScreenDstX: 8 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 2,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 3,
				SpriteGridDstY: 6,
			},
			{
				ID: DeskID,
				BaseY: 6 * tileSize,
				ScreenDstX: 11 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 2,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 3,
				SpriteGridDstY: 6,
			},
			{
				ID: DeskID,
				BaseY: 6 * tileSize,
				ScreenDstX: 20 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 2,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 3,
				SpriteGridDstY: 6,
			},
			{
				ID: DeskID,
				BaseY: 6 * tileSize,
				ScreenDstX: 23 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 2,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 3,
				SpriteGridDstY: 6,
			},
			{
				ID: DeskID,
				BaseY: 6 * tileSize,
				ScreenDstX: 26 * tileSize,
				ScreenDstY: 6 * tileSize,
				SpriteImg: interiorImage,
				SpriteGridSrcX: 2,
				SpriteGridSrcY: 5,
				SpriteGridDstX: 3,
				SpriteGridDstY: 6,
			},
			{
				ID: NpcID,
				BaseY: 3 * tileSize,
				ScreenDstX: 22 * tileSize,
				ScreenDstY: 2 * tileSize,
				SpriteImg: npc1SetImage,
				SpriteGridSrcX: 11,
				SpriteGridSrcY: 10,
				SpriteGridDstX: 12,
				SpriteGridDstY: 12,
			},
			{
				ID: NpcID,
				BaseY: 3 * tileSize,
				ScreenDstX: 4 * tileSize,
				ScreenDstY: 2 * tileSize,
				SpriteImg: npc2SetImage,
				SpriteGridSrcX: 11,
				SpriteGridSrcY: 10,
				SpriteGridDstX: 12,
				SpriteGridDstY: 12,
			},
		},
		WarpTriggers: []WarpTrigger {
			{
				DespawnX: 29,
				DespawnY: 7,
				SpawnX: 2,
				SpawnY: 20,
				TargetScene: SceneReception,
			},
			{
				DespawnX: 29,
				DespawnY: 8,
				SpawnX: 2,
				SpawnY: 21,
				TargetScene: SceneReception,
			},
		},
		InteractionTriggers: []InteractionTrigger {
			{
				GridX: 22,
				GridY: 3,
				PlayerDir: DirUp,
				InteractionID: 2,
			},
			{
				GridX: 4,
				GridY: 3,
				PlayerDir: DirUp,
				InteractionID: 1,
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
			ContainMiniGame: false,
			MiniGameState: 0,
		},
		{
			InteractionID: 1,
			InteractionType: InteractionTypeConversation,
			InteractionStates: 3,
			InteractionText: []string {
				"Lunch?",
				"Chipotle or Moe's?",
				"Moe's.",
			},
			ContainMiniGame: false,
			MiniGameState: 0,
		},
		{
			InteractionID: 2,
			InteractionType: InteractionTypeMiniGame0,
			InteractionStates: 3,
			InteractionText: []string {
				"Mini Game?",
				"...",
				"Done!",
			},
			ContainMiniGame: true,
			MiniGameState: 1,
		},
	}

	// Initialize mini dino game
	miniDinoGame = DinoGame {
		BgImage: miniDinoBg,
		PlayerSetImage: playerSetImage,
		Player: Player {
			GridX: 5, // Player start position in scene in grid unit
			GridY: 5,
			PixelX: CenterX, // Player start position in pixel
			PixelY: GroundY - 2 * tileSize,
			Action: DinoActionIdle,
			Dir: DirRight,
		},
		WidthPixels: float64(parkingLotBg.Bounds().Dx()),
		HeightPixels: float64(parkingLotBg.Bounds().Dy()),
		Obstacles: []RenderItem {
			{
				ID: TreeID,
				BaseY: 19 * tileSize, // base of tree
				ScreenDstX: CenterX + 10 * tileSize,
				ScreenDstY: GroundY - 2 * tileSize,
				SpriteImg: exteriorImage,
				SpriteGridSrcX: 33,
				SpriteGridSrcY: 10,
				SpriteGridDstX: 34,
				SpriteGridDstY: 12,
			},
		},
		CurrentState: StateTitle,
		Score: 0,
		TouchButtons: []TouchButton {
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
		},
		IsActive: false,
		PlayerVelocityY: 0,
	}
}
