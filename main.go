package main

import (
	"github.com/lnunno/snake/snake"
)

func startGame() {
	s := snake.Snake{
		Body:
		[]snake.Coord {
			{2, 2},
			{1, 2},
			{0, 2},
		},
	}
	game := snake.Game{
		s,
		snake.Field{XSize: 30, YSize: 30, Members: make(map[string]string) },
		0,
		[]snake.Coord {
			{2,2},
			{5,5},
			{7,7},
			{9,9},
		},
	}
	game.Run()
}

func main() {
	startGame()
}
