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

type SnakeState int

const (
	MOVING SnakeState = iota
	GROWING
	DYING
)

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

func (snake *Snake) Move(dir Direction, field *Field) SnakeState {
	state := MOVING
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
		state = GROWING
	} else if oldChar == snakeBodyChar || oldChar == snakeHeadChar {
		state = DYING
	}
	if head.X < 0 || head.Y < 0 {
		state = DYING
	} else if head.X >= field.XSize || head.Y >= field.YSize {
		state = DYING
	}
	field.Members[head.String()] = snakeHeadChar
	delete(field.Members, oldTail)
	return state
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
	Snake          Snake
	Field          Field
	Score          int
	Apples         []Coord
	NumApplesEaten int
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
		Field{XSize: 70, YSize: 45, Members: make(map[string]string) },
		0,
		[]Coord{},
		0,
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

var tickSpeed = 160 * time.Millisecond
var speedStep = 10 * time.Millisecond
var fastestSpeed = 70 * time.Millisecond
var inputTimeout = 1 * time.Millisecond
var iterationsPerNewApple = 15
var numApplesEatenSpeedup = 7

func (game *Game) Run() {
	for _, coord := range game.Snake.Body {
		game.Field.Members[coord.String()] = snakeBodyChar
	}

	numStartingApples := game.Field.XSize * game.Field.YSize / 70
	if numStartingApples <= 0 {
		numStartingApples = 5
	}
	for i := 0; i < numStartingApples; i++{
		c := game.Field.FindRandomEmptySpace()
		if c.X >= 0 {
			game.Field.PlaceApple(c)
		}
	}
	for _, coord := range game.Apples {
		game.Field.PlaceApple(coord)
	}

	previousDirection := RIGHT
	byteChannel := make(chan byte, 1)
	iterations := 0
	for {
		if (iterations % iterationsPerNewApple) == 0 {
			c := game.Field.FindRandomEmptySpace()
			if c.X >= 0 {
				game.Field.PlaceApple(c)
			}
		}
		snakeState := MOVING
		go readCharNonBlocking(byteChannel)
		movementDirection := previousDirection
		select {
		case b := <-byteChannel:
			movementDirection = fromString(string(b), previousDirection)
			snakeState = game.Snake.Move(movementDirection, &game.Field)
		case <-time.After(inputTimeout):
			snakeState = game.Snake.Move(movementDirection, &game.Field)
		}
		switch snakeState {
		case GROWING:
			game.NumApplesEaten += 1
			if game.NumApplesEaten % numApplesEatenSpeedup == 0 && (tickSpeed - speedStep) > fastestSpeed {
				tickSpeed -= speedStep
			}
		case DYING:
			game.Print()
			fmt.Println("You're dead! Game over.")
			return
		}
		previousDirection = movementDirection
		game.Score += 5 + (10 * game.NumApplesEaten)
		game.Print()
		time.Sleep(tickSpeed)
		iterations += 1
	}
}

func (game Game) Print() {
	fmt.Println(game.Text())
	//os.Stdout.Write(game.Json())
}

func (game Game) Text() string {
	var buffer bytes.Buffer
	for x := 0; x < game.Field.XSize+2; x++ {
		buffer.WriteString("=")
	}
	buffer.WriteString("\n")
	for y := 0; y < game.Field.YSize; y++ {
		for x := 0; x < game.Field.XSize; x++ {
			if x == 0 {
				buffer.WriteString("|")
			}
			coord := Coord{x, y}
			char := game.Field.Members[coord.String()]
			if len(char) == 0 {
				char = "."
			}
			buffer.WriteString(char)
			if x == game.Field.XSize - 1 {
				buffer.WriteString("|")
			}
		}
		buffer.WriteString("\n")
	}
	for x := 0; x < game.Field.XSize+2; x++ {
		buffer.WriteString("=")
	}
	s := fmt.Sprintf("\nScore: %d\t\tNum Apples Eaten: %d", game.Score, game.NumApplesEaten)
	buffer.WriteString(s)
	return buffer.String()
}

