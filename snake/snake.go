package snake

import (
	"fmt"
	"time"
	"os"
	"encoding/json"
	"bytes"
	"math/rand"
	"os/exec"
)

type Coord struct {
	X int
	Y int
}

func (c Coord) String() string {
	return fmt.Sprintf("%#v", c)
}

type Field struct {
	XSize   int
	YSize   int
	Members map[string]string
}

var appleChar = "@"

func (field Field) PlaceApple(coord Coord) {
	field.Members[coord.String()] = appleChar
}

func (field Field) FindRandomEmptySpace() Coord {
	numTries := 10
	for i := 0; i <= numTries; i++ {
		xCoord := rand.Intn(field.XSize)
		yCoord := rand.Intn(field.YSize)
		coord := Coord{xCoord, yCoord}
		if field.Members[coord.String()] == "" {
			return coord
		}
	}
	return Coord{-1, -1}
}

type Snake struct {
	Body []Coord
}

var snakeBodyChar = "*"
var snakeHeadChar = "O"

func (snake Snake) Head() *Coord {
	return &snake.Body[0]
}

func (snake *Snake) Tail() *Coord {
	return &snake.Body[len(snake.Body) - 1]
}

func (snake *Snake) Move(dir Direction, field *Field) {
	oldTail := snake.Body[len(snake.Body) - 1].String()
	for index := len(snake.Body) - 1; index >= 1; index-- {
		// Special case for growing
		field.Members[snake.Body[index - 1].String()] = snakeBodyChar
		snake.Body[index] = snake.Body[index - 1]
	}
	head := snake.Head()
	switch dir {
	case UP:
		head.Y -= 1
	case DOWN:
		head.Y += 1
	case LEFT:
		head.X -= 1
	case RIGHT:
		head.X += 1
	}
	oldChar := field.Members[head.String()]
	if oldChar == appleChar {
		snake.Grow()
	}
	field.Members[head.String()] = snakeHeadChar
	delete(field.Members, oldTail)
}

func (snake *Snake) Grow() {
	var tail Coord = *snake.Tail()
	snake.Body = append(snake.Body, tail)
}

type Direction int

const (
	UP Direction = iota
	DOWN
	LEFT
	RIGHT
)

func fromString(s string, initialDirection Direction) Direction {
	switch s {
	case "a": return LEFT
	case "s": return DOWN
	case "w": return UP
	case "d": return RIGHT
	default:
		return initialDirection
	}
}

type Game struct {
	Snake  Snake
	Field  Field
	Score  int
	Apples []Coord
}

func StartGame() {
	s := Snake{
		Body:
		[]Coord{
			{2, 2},
			{1, 2},
			{0, 2},
		},
	}
	game := Game{
		s,
		Field{XSize: 30, YSize: 30, Members: make(map[string]string) },
		0,
		[]Coord{
			{5, 5},
			{7, 7},
			{9, 9},
		},
	}

	// disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()

	game.Run()
}

func (game Game) Json() []byte {
	result, err := json.Marshal(game)
	if err != nil {
		fmt.Println("error: ", err)
	}
	return result
}

func readCharNonBlocking(byteChannel chan <- byte) {
	var b []byte = make([]byte, 1)
	_, err := os.Stdin.Read(b)
	if err != nil {
		fmt.Println("error: ", err)
	}
	byteChannel <- b[0]
}

var tickSpeed = 200 * time.Millisecond
var inputTimeout = 50 * time.Millisecond

func (game Game) Run() {
	for _, coord := range game.Snake.Body {
		game.Field.Members[coord.String()] = snakeBodyChar
	}

	for _, coord := range game.Apples {
		game.Field.PlaceApple(coord)
	}

	previousDirection := RIGHT
	byteChannel := make(chan byte, 1)
	for {
		go readCharNonBlocking(byteChannel)
		movementDirection := previousDirection
		select {
		case b := <-byteChannel:
			movementDirection = fromString(string(b), previousDirection)
			game.Snake.Move(movementDirection, &game.Field)
		case <-time.After(inputTimeout):
			game.Snake.Move(movementDirection, &game.Field)
		}
		previousDirection = movementDirection
		game.Print()
		time.Sleep(tickSpeed)
	}
}

func (game Game) Print() {
	fmt.Println(game.Text())
	os.Stdout.Write(game.Json())
}

func (game Game) Text() string {
	var buffer bytes.Buffer
	for y := 0; y < game.Field.YSize; y++ {
		for x := 0; x < game.Field.XSize; x++ {
			coord := Coord{x, y}
			char := game.Field.Members[coord.String()]
			if len(char) == 0 {
				char = "."
			}
			buffer.WriteString(char)
		}
		buffer.WriteString("\n")
	}
	return buffer.String()
}

