package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 1200
	screenHeight = 800
	groundHeight = 100
)

type Vector2 struct {
	X, Y float64
}

func (v Vector2) Add(other Vector2) Vector2 {
	return Vector2{v.X + other.X, v.Y + other.Y}
}

func (v Vector2) Scale(s float64) Vector2 {
	return Vector2{v.X * s, v.Y * s}
}

func (v Vector2) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

type Ball struct {
	Position     Vector2
	Velocity     Vector2
	InitialPos   Vector2
	InitialVel   Vector2
	Time         float64
	Launched     bool
	Trail        []Vector2
	MaxTrailLen  int
	Color        color.RGBA
}

type Game struct {
	ball          Ball
	cannon        Vector2
	aimAngle      float64
	aimPower      float64
	showTrail     bool
	showVectors   bool
	paused        bool
	gravity       float64
	scale         float64
	timeScale     float64
	targets       []Vector2
	score         int
	attempts      int
}

// Consts
var (
	defaultGravity   = 9.8   // m/s²
	defaultScale     = 50.0  // pixels per meter
	defaultTimeScale = 1.0   // time multiplier
)

func NewGame() *Game {
	game := &Game{
		cannon:      Vector2{100, float64(screenHeight - groundHeight)},
		aimAngle:    45.0,
		aimPower:    20.0,
		showTrail:   true,
		showVectors: true,
		gravity:     defaultGravity,
		scale:       defaultScale,
		timeScale:   defaultTimeScale,
	}
	
	game.ball = Ball{
		Position:    game.cannon,
		MaxTrailLen: 200,
		Color:       color.RGBA{255, 100, 100, 255},
	}
	
	// Targets
	game.targets = []Vector2{
		{800, float64(screenHeight - groundHeight - 50)},
		{600, float64(screenHeight - groundHeight - 100)},
		{1000, float64(screenHeight - groundHeight - 30)},
	}
	
	return game
}

func (b *Ball) Update(dt float64) {
	if !b.Launched {
		return
	}
	
	b.Time += dt
	
	// Physics projectile motion equations
	b.Position.X = b.InitialPos.X + b.InitialVel.X*b.Time
	b.Position.Y = b.InitialPos.Y - (b.InitialVel.Y*b.Time - 0.5*9.8*b.Time*b.Time)
	
	// Add to trail
	if len(b.Trail) > 0 {
		lastPos := b.Trail[len(b.Trail)-1]
		distance := math.Sqrt((b.Position.X-lastPos.X)*(b.Position.X-lastPos.X) + 
							 (b.Position.Y-lastPos.Y)*(b.Position.Y-lastPos.Y))
		if distance > 5 {
			b.Trail = append(b.Trail, b.Position)
		}
	} else {
		b.Trail = append(b.Trail, b.Position)
	}
	
	// Limit trail length
	if len(b.Trail) > b.MaxTrailLen {
		b.Trail = b.Trail[1:]
	}
}

func (b *Ball) Launch(angle, power float64, startPos Vector2) {
	b.Launched = true
	b.Time = 0
	b.InitialPos = startPos
	b.Position = startPos
	b.Trail = []Vector2{startPos}
	
	// Convert to rads
	angleRad := angle * math.Pi / 180.0
	b.InitialVel = Vector2{
		X: power * math.Cos(angleRad),
		Y: power * math.Sin(angleRad),
	}
	b.Velocity = b.InitialVel
}

func (b *Ball) Reset() {
	b.Launched = false
	b.Time = 0
	b.Trail = []Vector2{}
}

func (b *Ball) IsGrounded() bool {
	return b.Position.Y >= float64(screenHeight-groundHeight-10)
}

func (g *Game) Update() error {
	if !g.paused {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			if !g.ball.Launched {
				g.ball.Launch(g.aimAngle, g.aimPower, g.cannon)
				g.attempts++
			} else {
				g.ball.Reset()
				g.ball.Position = g.cannon
			}
		}
		
		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) && g.aimAngle < 90 {
			g.aimAngle += 1
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowDown) && g.aimAngle > 0 {
			g.aimAngle -= 1
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowRight) && g.aimPower < 50 {
			g.aimPower += 0.5
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) && g.aimPower > 5 {
			g.aimPower -= 0.5
		}
		
		// Update ball
		if g.ball.Launched {
			g.ball.Update(1.0/60.0 * g.timeScale)
			
			// Check if ball hit ground
			if g.ball.IsGrounded() {
				// Check if hit any targets
				for i, target := range g.targets {
					distance := math.Sqrt((g.ball.Position.X-target.X)*(g.ball.Position.X-target.X) + 
										 (g.ball.Position.Y-target.Y)*(g.ball.Position.Y-target.Y))
					if distance < 30 {
						g.score++
						// Remove hit target
						g.targets = append(g.targets[:i], g.targets[i+1:]...)
						break
					}
				}
			}
		}
	}
	
	if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		g.showTrail = !g.showTrail
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyV) {
		g.showVectors = !g.showVectors
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.paused = !g.paused
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		*g = *NewGame() // Reset game
	}
	
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen
	screen.Fill(color.RGBA{135, 206, 235, 255}) // Sky blue
	
	// Draw ground
	vector.DrawFilledRect(screen, 0, float32(screenHeight-groundHeight), 
						 screenWidth, groundHeight, color.RGBA{34, 139, 34, 255}, false)
	
	// Draw cannon
	cannonSize := float32(20)
	vector.DrawFilledCircle(screen, float32(g.cannon.X), float32(g.cannon.Y), 
						   cannonSize, color.RGBA{64, 64, 64, 255}, false)
	
	// Draw aim line
	if !g.ball.Launched {
		angleRad := g.aimAngle * math.Pi / 180.0
		aimLength := g.aimPower * 3
		endX := g.cannon.X + math.Cos(angleRad)*aimLength
		endY := g.cannon.Y - math.Sin(angleRad)*aimLength
		
		vector.StrokeLine(screen, float32(g.cannon.X), float32(g.cannon.Y),
						 float32(endX), float32(endY), 3, color.RGBA{255, 255, 0, 255}, false)
	}
	
	// Draw predicted trajectory
	if !g.ball.Launched && g.showVectors {
		angleRad := g.aimAngle * math.Pi / 180.0
		vx := g.aimPower * math.Cos(angleRad)
		vy := g.aimPower * math.Sin(angleRad)
		
		for t := 0.0; t < 10.0; t += 0.1 {
			x := g.cannon.X + vx*t
			y := g.cannon.Y - (vy*t - 0.5*g.gravity*t*t)
			
			if y >= float64(screenHeight-groundHeight) {
				break
			}
			
			vector.DrawFilledCircle(screen, float32(x), float32(y), 2, 
								   color.RGBA{255, 255, 0, 100}, false)
		}
	}
	
	// Draw ball trail
	if g.showTrail && len(g.ball.Trail) > 1 {
		for i := 1; i < len(g.ball.Trail); i++ {
			alpha := uint8(float64(i) / float64(len(g.ball.Trail)) * 255)
			trailColor := color.RGBA{255, 200, 200, alpha}
			
			vector.StrokeLine(screen, float32(g.ball.Trail[i-1].X), float32(g.ball.Trail[i-1].Y),
							 float32(g.ball.Trail[i].X), float32(g.ball.Trail[i].Y), 
							 2, trailColor, false)
		}
	}
	
	// Draw ball
	ballRadius := float32(8)
	vector.DrawFilledCircle(screen, float32(g.ball.Position.X), float32(g.ball.Position.Y), 
						   ballRadius, g.ball.Color, false)
	
	// Draw targets
	for _, target := range g.targets {
		vector.DrawFilledCircle(screen, float32(target.X), float32(target.Y), 15, 
							   color.RGBA{255, 0, 0, 255}, false)
		vector.DrawFilledCircle(screen, float32(target.X), float32(target.Y), 10, 
							   color.RGBA{255, 255, 255, 255}, false)
		vector.DrawFilledCircle(screen, float32(target.X), float32(target.Y), 5, 
							   color.RGBA{255, 0, 0, 255}, false)
	}
	
	// Draw velocity vector
	if g.showVectors && g.ball.Launched {
		scale := 0.1
		endX := g.ball.Position.X + g.ball.Velocity.X*scale
		endY := g.ball.Position.Y - g.ball.Velocity.Y*scale
		
		vector.StrokeLine(screen, float32(g.ball.Position.X), float32(g.ball.Position.Y),
						 float32(endX), float32(endY), 2, color.RGBA{0, 255, 0, 255}, false)
	}
	
	// Draw UI
	g.drawUI(screen)
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Draw semi-transparent background for UI
	vector.DrawFilledRect(screen, 10, 10, 300, 200, color.RGBA{0, 0, 0, 128}, false)
	
	// Draw text information
	texts := []string{
		fmt.Sprintf("Angle: %.1f°", g.aimAngle),
		fmt.Sprintf("Power: %.1f m/s", g.aimPower),
		fmt.Sprintf("Score: %d", g.score),
		fmt.Sprintf("Attempts: %d", g.attempts),
		"",
		"Controls:",
		"Arrow Keys: Aim & Power",
		"Space: Launch/Reset",
		"T: Toggle Trail",
		"V: Toggle Vectors",
		"P: Pause",
		"R: Reset Game",
	}
	
	for i, text := range texts {
		ebitenutil.DebugPrintAt(screen, text, 20, 20+i*15)
	}
	
	// Draw physics info
	if g.ball.Launched {
		physicsTexts := []string{
			fmt.Sprintf("Time: %.2f s", g.ball.Time),
			fmt.Sprintf("Height: %.1f m", (float64(screenHeight-groundHeight)-g.ball.Position.Y)/g.scale),
			fmt.Sprintf("Distance: %.1f m", (g.ball.Position.X-g.cannon.X)/g.scale),
			fmt.Sprintf("Vx: %.1f m/s", g.ball.Velocity.X),
			fmt.Sprintf("Vy: %.1f m/s", g.ball.Velocity.Y),
		}
		
		for i, text := range physicsTexts {
			ebitenutil.DebugPrintAt(screen, text, screenWidth-200, 20+i*15)
		}
	}
	
	if g.paused {
		ebitenutil.DebugPrintAt(screen, "PAUSED", screenWidth/2-30, screenHeight/2)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := NewGame()
	
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Physics Simulator - Projectile Motion")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
