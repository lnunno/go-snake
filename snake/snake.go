package snake

import (
	"fmt"
	"time"
	//"os/exec"
	"os"
	"encoding/json"
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

type Snake struct {
	Body []Coord
}

func (snake Snake) Head() *Coord {
	return &snake.Body[0]
}

func (snake *Snake) Tail() *Coord {
	return &snake.Body[len(snake.Body) - 1]
}

func Move(snake *Snake, dir Direction, field *Field) {
	growing := false
	oldTail := snake.Body[len(snake.Body) - 1].String()
	for index := len(snake.Body) - 1; index >= 1; index-- {
		// Special case for growing
		if snake.Body[index] == snake.Body[index - 1] {
			growing = true
		}
		field.Members[snake.Body[index - 1].String()] = "*"
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
	field.Members[head.String()] = "O"
	if !growing {
		delete(field.Members, oldTail)
	}
}

func (snake *Snake) Grow() {
	tail := snake.Tail()
	snake.Body = append(snake.Body, *tail)
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

func (game Game) Json() {
	result, err := json.Marshal(game)
	if err != nil {
		fmt.Println("error: ", err)
	}
	os.Stdout.Write(result)
}

func (game Game) Run() {
	for _, coord := range game.Snake.Body {
		game.Field.Members[coord.String()] = "*"
	}

	// disable input buffering
	//exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	//exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	var b []byte = make([]byte, 1)
	for {
		time.Sleep(30 * time.Millisecond)
		os.Stdin.Read(b)
		Move(&game.Snake, fromString(string(b), RIGHT), &game.Field)
		game.Print()
	}
}

func (game Game) Print() {
	for y := 0; y < game.Field.YSize; y++ {
		for x := 0; x < game.Field.XSize; x++ {
			coord := Coord{x, y}
			char := game.Field.Members[coord.String()]
			if len(char) == 0 {
				char = "."
			}
			fmt.Print(char)
		}
		fmt.Print("\n")
	}
	game.Json()
	//fmt.Printf("\033[0;0H")
	time.Sleep(300 * time.Millisecond)
}

