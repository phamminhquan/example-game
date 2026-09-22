package main

import (
	"log"
)

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

// Function: return the coordinate of the exterior renderable item in the
// spritesheet in pixels based on item ID
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
