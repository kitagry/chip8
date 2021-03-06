package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten"
)

const (
	width  = 64
	height = 32
)

type Display struct {
	data   [width][height]byte
	keys   [16]bool
	update bool
}

func NewDisplay() *Display {
	return &Display{}
}

func (d *Display) Set(x, y uint8) (bool, error) {
	if x >= width {
		x -= width
	}

	if y >= height {
		y -= height
	}

	d.data[x][y] = 1 ^ d.data[x][y]
	return d.data[x][y] == 1, nil
}

func (d *Display) Get(x, y uint8) (byte, error) {
	if x >= width {
		return 0, fmt.Errorf("invalid x: %d", x)
	}

	if y >= height {
		return 0, fmt.Errorf("invalid y: %d", y)
	}

	return d.data[x][y], nil
}

func (d *Display) SetFlag() {
	d.update = true
}

func (d *Display) Clear() {
	for i := 0; i < len(d.data); i++ {
		for j := 0; j < len(d.data[0]); j++ {
			d.data[i][j] = 0
		}
	}
}

func (d *Display) Update(screen *ebiten.Image) error {
	for i := 0; i < len(d.keys); i++ {
		d.keys[i] = false
	}
	switch {
	case ebiten.IsKeyPressed(ebiten.Key1):
		d.keys[0] = true
	case ebiten.IsKeyPressed(ebiten.Key2):
		d.keys[1] = true
	case ebiten.IsKeyPressed(ebiten.Key3):
		d.keys[2] = true
	case ebiten.IsKeyPressed(ebiten.Key4):
		d.keys[3] = true
	case ebiten.IsKeyPressed(ebiten.KeyQ):
		d.keys[4] = true
	case ebiten.IsKeyPressed(ebiten.KeyW):
		d.keys[5] = true
	case ebiten.IsKeyPressed(ebiten.KeyE):
		d.keys[6] = true
	case ebiten.IsKeyPressed(ebiten.KeyR):
		d.keys[7] = true
	case ebiten.IsKeyPressed(ebiten.KeyA):
		d.keys[8] = true
	case ebiten.IsKeyPressed(ebiten.KeyS):
		d.keys[9] = true
	case ebiten.IsKeyPressed(ebiten.KeyD):
		d.keys[10] = true
	case ebiten.IsKeyPressed(ebiten.KeyF):
		d.keys[11] = true
	case ebiten.IsKeyPressed(ebiten.KeyZ):
		d.keys[12] = true
	case ebiten.IsKeyPressed(ebiten.KeyX):
		d.keys[13] = true
	case ebiten.IsKeyPressed(ebiten.KeyC):
		d.keys[14] = true
	case ebiten.IsKeyPressed(ebiten.KeyV):
		d.keys[15] = true
	}
	return nil
}

func (d *Display) Draw(screen *ebiten.Image) {
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			d := d.data[x][y]
			c := color.White
			if d&1 == 1 {
				c = color.Black
			}
			screen.Set(2*x, 2*y, c)
			screen.Set(2*x+1, 2*y, c)
			screen.Set(2*x, 2*y+1, c)
			screen.Set(2*x+1, 2*y+1, c)
		}
	}
}

func (d *Display) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 128, 64
}

func (d *Display) Run() error {
	ebiten.SetWindowSize(128, 64)
	ebiten.SetWindowTitle("chip8")
	if err := ebiten.RunGame(d); err != nil {
		return err
	}
	return nil
}
