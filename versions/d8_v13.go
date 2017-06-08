package main

import (
"strings"
"math/rand"
"os"
"time"
"fmt"
"bufio"
"math"
"strconv"
)


func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	bot := Bot{}
	game := Game{}
	game.initGame()
	game.run(bot)
}

/* Inputs:
settings timebank 10000
settings time_per_move 500
settings player_names player0,player1
settings your_bot player0
settings your_botid 0
settings field_width 16
settings field_height 16

update game round 0
update game field .,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,0,.,.,.,.,.,.,.,.,1,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.
action move 10000
*/
/* Inputs:
settings timebank 10000
settings time_per_move 500
settings player_names player0,player1
settings your_bot player0
settings your_botid 0
settings field_width 16
settings field_height 16

update game round 0
update game field .,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,x,x,x,.,.,.,.,.,.,.,.,.,x,x,x,x,x,.,x,.,.,.,.,.,.,.,.,.,x,x,x,x,x,.,x,.,.,.,.,.,.,.,.,.,.,.,x,x,x,.,x,.,.,.,.,.,.,.,.,.,.,x,x,x,x,.,x,x,x,x,x,.,.,.,.,.,.,x,1,x,x,.,x,x,x,x,x,.,.,.,.,.,.,.,x,x,x,x,x,x,x,.,x,x,.,.,.,.,.,.,.,x,x,x,x,x,x,.,x,x,.,.,.,.,.,x,x,x,x,x,x,x,.,.,.,.,.,.,.,.,.,x,x,x,x,x,x,x,x,x,.,.,.,.,.,.,.,.,.,.,.,x,x,x,x,x,x,.,.,.,.,.,.,.,.,.,.,x,x,0,x,x,x,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.,.
action move 8753
*/

type Node struct {
	field     Field
	positions [2]int
	player    int
	//children map[int]*Node
}

func coordToIndex(r, c int) int {
	return r*boardWidth + c
}
func (n Node) isFinal() bool {
	result := len(n.field.getEmptyNeighbors(n.positions[0])) == 0
	result = result || len(n.field.getEmptyNeighbors(n.positions[1])) == 0
	return result
}
func (n Node) evaluate(d int) float64 {
	open0 := n.field.getEmptyNeighbors(n.positions[n.player])
	open1 := n.field.getEmptyNeighbors(n.positions[1-n.player])
	if len(open0) == 0 && len(open1) > 0 {
		return -10000.0 - float64(d) // losing as tardively as possible
	}
	if len(open0) > 0 && len(open1) == 0 {
		return 10000.0 + float64(d) // winning as fast as possible
	}
	v0, v1 := voronoi(open0, make(map[int]int), open1, make(map[int]int), n.field)
	score := float64(v0) - float64(v1)
	return score
}
func (n Node) next(playerMoving int, dest int) Node {
	// copy
	result := n.copy()
	// new position
	result.positions[playerMoving] = dest
	// new field
	result.field.cells[dest] = blocked
	return result
}
func (n Node) copy() Node {
	result := Node{player: n.player}
	copy(result.positions[:], n.positions[:])
	result.field = n.field.copy()
	return result
}
func alphabeta(n Node, depth int, alpha, beta float64, maximizingPlayer int) (float64, int) {
	if depth == 0 || n.isFinal() { //is final node
		return n.evaluate(depth), -1
	}
	moves := n.field.getEmptyNeighbors(n.positions[maximizingPlayer]) //self.possible_moves(board, curr)
	orderToReturn := -1
	for destination, order := range moves {
		nextState := n.next(maximizingPlayer, destination)
		score, _ := alphabeta(nextState, depth-1, alpha, beta, 1-maximizingPlayer)
		if maximizingPlayer == n.player {
			if score > alpha {
				alpha = score
				orderToReturn = order
			}
			if alpha >= beta {
				break
			}
		} else {
			if score < beta {
				beta = score
			}
			if alpha >= beta {
				break
			}
		}
	}
	if maximizingPlayer == n.player {
		return alpha, orderToReturn
	} else {
		return beta, -1
	}
}


// Field class
type Field struct {
	cells [nCells]int
}

func (f Field) copy() Field {
	result := Field{}
	copy(result.cells[:], f.cells[:])
	return result
}
func (f *Field) initField() {
	for i := 0; i < nCells; i++ {
		f.cells[i] = empty
	}
}
func stringToInt(s string) int {
	switch s {
	case ".":
		return empty
	case "0":
		return player1
	case "1":
		return player2
	default:
		return blocked
	}
}
func (f *Field) parse(text string) {
	values := strings.Split(text, ",")
	for i := 0; i < nCells; i++ {
		f.cells[i] = stringToInt(values[i])
	}
}
func (f Field) getEmptyNeighbors(cell int) map[int]int { // map [order]destination
	up := cell - boardWidth
	down := cell + boardWidth
	right := cell + 1
	left := cell - 1

	result := make(map[int]int)
	if up >= 0 && f.cells[up] == empty {
		result[up] = UP
	}
	if down < nCells && f.cells[down] == empty {
		result[down] = DOWN
	}
	if right%boardWidth != 0 && right < nCells && f.cells[right] == empty {
		result[right] = RIGHT
	}
	if left >= 0 && left%boardWidth != boardWidth-1 && f.cells[left] == empty {
		result[left] = LEFT
	}
	return result
}
func (f Field) print() {
	for row := 0; row < boardHeight; row++ {
		for col := 0; col < boardWidth; col++ {
			switch f.cells[coordToIndex(row, col)] {
			case player1:
				fmt.Fprint(os.Stderr, "0")
			case player2:
				fmt.Fprint(os.Stderr, "1")
			case empty:
				fmt.Fprint(os.Stderr, ".")
			default:
				fmt.Fprint(os.Stderr, "X")
			}
		}
		fmt.Fprintln(os.Stderr, "")
	}
}

func voronoi(open0, close0, open1, close1 map[int]int, field Field) (int, int) {
	//fmt.Fprintln(os.Stderr, len(close0), len(close1))
	if len(open0) == 0 && len(open1) == 0 {
		return len(close0), len(close1)
	}
	nextOpen0 := make(map[int]int)
	// o = open destination
	for o, _ := range open0 {
		_, inClose1 := close1[o] // closer to the opponent
		_, inOpen1 := open1[o]   // same distance as the opponent
		_, inClose0 := close0[o] // already tested
		if !(inClose1 || inOpen1 || inClose0) {
			close0[o] = 1
			for dest, _ := range field.getEmptyNeighbors(o) {
				nextOpen0[dest] = 1
			}
		}
	}
	nextOpen1 := make(map[int]int)
	for o, _ := range open1 {
		_, inClose0 := close0[o] // closer to the opponent
		_, inOpen0 := open0[o]   // same distance as the opponent
		_, inClose1 := close1[o] // already tested
		if !(inClose0 || inOpen0 || inClose1) {
			close1[o] = 1
			for dest, _ := range field.getEmptyNeighbors(o) {
				nextOpen1[dest] = 1
			}
		}
	}
	return voronoi(nextOpen0, close0, nextOpen1, close1, field)
}


// Bot here goes Hal
type Bot struct {
	row, col int
}

func (b Bot) play(g Game) {
	eRow := -1
	eCol := -1
	for row := 0; row < g.fieldHeight; row++ {
		for col := 0; col < g.fieldWidth; col++ {
			if g.field.cells[row*boardWidth+col] == g.myBotID {
				b.row = row
				b.col = col
			}
			if g.field.cells[row*boardWidth+col] == 1-g.myBotID {
				eRow = row
				eCol = col
			}
		}
	}
	//g.field.print()
	// create root
	root := Node{}
	root.player = g.myBotID
	root.positions[g.myBotID] = coordToIndex(b.row, b.col)
	root.positions[1-g.myBotID] = coordToIndex(eRow, eCol)
	root.field = g.field.copy()

	//fmt.Fprintln(os.Stderr, "eval root:", root.evaluate(0))

	// expand
	//startTree := time.Now()
	depth := 8
	_, intOrder := alphabeta(root, depth, math.Inf(-1), math.Inf(1), root.player)
	//elapsedTime := time.Since(startTree).Nanoseconds()
	if intOrder == -1 { // enemy dies immediately
		fmt.Fprintln(os.Stderr, "Enemy is dying, now :-)")
		open := g.field.getEmptyNeighbors(coordToIndex(b.row, b.col))
		for _, order := range open {
			intOrder = order
		}
	}
	// log
	//fmt.Fprintln(os.Stderr, "Depth:", depth, ", time elapsed:", (elapsedTime / 1000000), "ms, score:", score)

	// play
	fmt.Println(intToStringOrder(intOrder))

	/*
		fmt.Fprintln(os.Stderr, "bot   position: row = ", b.row, ", col = ", b.col)
		fmt.Fprintln(os.Stderr, "enemy position: row = ", eRow, ", col = ", eCol)
		//
		open0 := g.field.getEmptyNeighbors(b.row*boardWidth + b.col)
		open1 := g.field.getEmptyNeighbors(eRow*boardWidth + eCol)
		v0, v1 := voronoi(open0, make(map[int]int), open1, make(map[int]int), g.field)
		fmt.Fprintln(os.Stderr, "open0:", open0)
		fmt.Fprintln(os.Stderr, "open1:", open1)
		fmt.Fprintln(os.Stderr, "voronoi:", v0, v1)*/
}

// Player class, unused yet
type Player struct {
}

const player1 int = 0
const player2 int = 1
const empty int = 2
const blocked int = 3

const nCells int = 256
const boardWidth int = 16
const boardHeight int = 16

const UP int = 0
const DOWN int = 1
const RIGHT int = 2
const LEFT int = 3

func intToStringOrder(a int) string {
	switch a {
	case UP:
		return "up"
	case DOWN:
		return "down"
	case RIGHT:
		return "right"
	case LEFT:
		return "left"
	default:
		return "order FUCK"
	}
}


// Game class, based on the python bot
type Game struct {
	initialTimebank, lastTimebank int
	timePerMove                   int
	playerNames                   [2]string
	myBot                         string
	myBotID, otherBotID           int
	fieldWidth, fieldHeight       int
	field                         Field
	round                         int
	players                       [2]Player
}

func (g *Game) initGame() {
	g.timePerMove = 10
	g.playerNames[0] = "I am not set"
	g.playerNames[1] = "Other ain t set"
	g.myBot = "I am still not set"
	g.myBotID = -1
	g.otherBotID = -1
	g.field = Field{}
	g.players[0] = Player{}
	g.players[1] = Player{}
}
func (g Game) myPlayer() Player {
	return g.players[g.myBotID]
}
func (g Game) otherPlayer() Player {
	return g.players[g.otherBotID]
}
func (g *Game) update(message []string) {
	if strings.Compare(message[0], "settings") == 0 {
		if strings.Compare(message[1], "timebank") == 0 {
			g.lastTimebank, _ = strconv.Atoi(message[2])
		} else if strings.Compare(message[1], "time_per_move") == 0 {
			g.timePerMove, _ = strconv.Atoi(message[2])
		} else if strings.Compare(message[1], "player_names") == 0 {
			names := strings.Split(message[2], ",")
			g.playerNames[0] = names[0]
			g.playerNames[1] = names[1]
		} else if strings.Compare(message[1], "your_bot") == 0 {
			g.myBot = message[2]
		} else if strings.Compare(message[1], "your_botid") == 0 {
			g.myBotID, _ = strconv.Atoi(message[2])
			g.otherBotID = 1 - g.myBotID
		} else if strings.Compare(message[1], "field_width") == 0 {
			g.fieldWidth, _ = strconv.Atoi(message[2])
		} else if strings.Compare(message[1], "field_height") == 0 {
			g.fieldHeight, _ = strconv.Atoi(message[2])
		} else if strings.Compare(message[1], "timebank") == 0 {
			g.lastTimebank, _ = strconv.Atoi(message[2])
		} else {
			fmt.Fprintln(os.Stderr, "Can't read settings:", message, "in game.update.")
		}
	} else if strings.Compare(message[0], "update") == 0 {
		if strings.Compare(message[1], "game") == 0 {
			if strings.Compare(message[2], "round") == 0 {
				g.round, _ = strconv.Atoi(message[3])
			} else if strings.Compare(message[2], "field") == 0 {
				g.field.parse(message[3])
			}
		}
	} else if strings.Compare(message[0], "action") == 0 && strings.Compare(message[1], "move") == 0 {
		g.lastTimebank, _ = strconv.Atoi(message[2])
	} /*else {
		fmt.Fprintln(os.Stderr, "Can't read:", message, "in game.update.")
	}*/
}
func (g *Game) run(bot Bot) {
	g.field = Field{}
	reader := bufio.NewReader(os.Stdin)
	//fieldIsSet := false
	for g.round < 202 {
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		message := strings.Split(text, " ")
		g.update(message)
		/*if !fieldIsSet && g.fieldHeight != 0 && g.fieldWidth != 0 {
			g.field.initField(g.fieldHeight, g.fieldWidth)
			fieldIsSet = true
		}*/
		if strings.Compare(message[0], "action") == 0 && strings.Compare(message[1], "move") == 0 {
			bot.play(*g)
		}
	}
}
