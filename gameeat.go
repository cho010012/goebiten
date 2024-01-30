package main

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

// 螢幕大小
const (
	screenWidth  = 640
	screenHeight = 480
	cellSize     = 10
)

// 遊戲設置相關參數
type Game struct {
	snake       *Snake
	food        *Food
	lastDir     Direction
	gameOver    bool
	gameOverMsg string
	restart     bool
	gameState   GameState // Add gameState field
}
type GameState int

// 遊戲按鈕定義暫停結束
const (
	StatePaused GameState = iota
	StateRunning
	StateGameOver
)

type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
)

type Snake struct {
	segments []Segment
}

type Segment struct {
	x, y int
}

type Food struct {
	x, y int
}

func (s *Snake) move(dir Direction) {
	head := s.segments[0]
	var newHead Segment

	switch dir {
	case DirUp:
		newHead = Segment{head.x, head.y - 1}
	case DirDown:
		newHead = Segment{head.x, head.y + 1}
	case DirLeft:
		newHead = Segment{head.x - 1, head.y}
	case DirRight:
		newHead = Segment{head.x + 1, head.y}
	}

	s.segments = append([]Segment{newHead}, s.segments...)
	s.segments = s.segments[:len(s.segments)-1]
}

func (s *Snake) draw(screen *ebiten.Image) {
	for _, segment := range s.segments {
		ebitenutil.DrawRect(screen, float64(segment.x*cellSize), float64(segment.y*cellSize), float64(cellSize), float64(cellSize), color.RGBA{0, 255, 0, 255})
	}
}

func (f *Food) draw(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, float64(f.x*cellSize), float64(f.y*cellSize), float64(cellSize), float64(cellSize), color.RGBA{255, 0, 0, 255})
}

// Game 結構應實作 ebiten.Game 介面的 Draw 方法
func (g *Game) Draw(screen *ebiten.Image) {
	switch g.gameState {
	case StatePaused:
		ebitenutil.DebugPrint(screen, "Press Space to start the game.")
	case StateRunning:
		g.snake.draw(screen)
		g.food.draw(screen)
	case StateGameOver:
		ebitenutil.DebugPrint(screen, g.gameOverMsg+"\nPress 'R' to restart.")
	}
}

func (f *Food) randomizePosition() {
	f.x = rand.Intn(screenWidth / cellSize)
	f.y = rand.Intn(screenHeight / cellSize)
}

// 在 Update 函數中檢查 restart 標誌並執行相應的邏輯
func (g *Game) Update(screen *ebiten.Image) error {
	switch {
	case g.gameState == StatePaused:
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			log.Println("Game started")
			g.gameState = StateRunning
		}
	case g.gameState == StateRunning:
		if g.restart {
			log.Println("Game restarted")
			g.restartGame()
			g.restart = false
		}

		if g.gameOver {
			g.gameState = StateGameOver
		} else {
			if err := g.handleInput(); err != nil {
				return err
			}
			g.snake.move(g.lastDir)
			g.checkCollision()
			g.generateFood()
		}
	case g.gameState == StateGameOver:
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			log.Println("Game restarted")
			g.restart = true
			g.gameState = StatePaused
		}
	}

	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// handleInput 函數中添加重啟按鈕的檢測
func (g *Game) handleInput() error {
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyUp):
		g.lastDir = DirUp
	case ebiten.IsKeyPressed(ebiten.KeyDown):
		g.lastDir = DirDown
	case ebiten.IsKeyPressed(ebiten.KeyLeft):
		g.lastDir = DirLeft
	case ebiten.IsKeyPressed(ebiten.KeyRight):
		g.lastDir = DirRight
	case ebiten.IsKeyPressed(ebiten.KeyR):
		// Only set restart to true if it's currently false to avoid repeated setting
		if !g.restart {
			g.restart = true
			fmt.Println("R key pressed")
		}
	}

	return nil
}

// 在 restartGame 函数中修改 Snake 的初始位置
func (g *Game) restartGame() {
	g.snake = &Snake{
		segments: []Segment{{4, 4}, {3, 4}, {2, 4}}, // 修改为新的起始位置
	}
	g.food = &Food{}
	g.food.randomizePosition()
	g.gameOver = false
	g.gameOverMsg = ""
	g.restart = false
}

func (g *Game) checkCollision() {
	head := g.snake.segments[0]

	// 撞墙处理
	if head.x < 0 || head.x >= screenWidth/cellSize || head.y < 0 || head.y >= screenHeight/cellSize {
		g.gameOver = true
		g.gameOverMsg = "Game Over - Collided with the wall!"
		return
	}

	// 检查是否与身体碰撞
	for i := 1; i < len(g.snake.segments); i++ {
		if head.x == g.snake.segments[i].x && head.y == g.snake.segments[i].y {
			g.gameOver = true
			g.gameOverMsg = "Game Over - Collided with yourself!"
			return
		}
	}
}
func (g *Game) generateFood() {
	head := g.snake.segments[0]
	if head.x == g.food.x && head.y == g.food.y {
		g.snake.segments = append(g.snake.segments, Segment{})
		g.food.randomizePosition()
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Snake Game in Golang")

	// 将错误信息输出到 "error.log"
	errorLogFile, err := os.Create("error.log")
	if err != nil {
		log.Fatal(err)
	}
	defer errorLogFile.Close()
	log.SetOutput(errorLogFile)

	// 记录程序启动信息到 "game.log"
	gameLogFile, err := os.Create("game.log")
	if err != nil {
		log.Fatal(err)
	}
	defer gameLogFile.Close()

	// 在这里将程序启动信息同时输出到标准输出和 "game.log"
	multiLogFile := io.MultiWriter(os.Stdout, gameLogFile)
	log.SetOutput(multiLogFile)

	// 记录程序启动信息
	log.Println("Game started")
	rand.Seed(time.Now().UnixNano())

	game := &Game{
		snake: &Snake{
			segments: []Segment{{2, 0}, {1, 0}, {0, 0}},
		},
		food: &Food{},
	}

	game.food.randomizePosition()
	// 其他初始化代码...
	rand.Seed(time.Now().UnixNano())
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Snake Game in Golang")

	fmt.Println("Press Space to start the game and R to restart.")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal("Error:", err)

	}
	log.Println("Game finished")
}
