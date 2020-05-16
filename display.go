package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten"
)

const (
	width  = 8
	height = 32
)

type Display struct {
	data   [width][height]byte
	update bool
}

func NewDisplay() *Display {
	return &Display{}
}

func (d *Display) Set(x, y uint8, data byte) error {
	if x >= width {
		return fmt.Errorf("invalid x: %d", x)
	}

	if y >= height {
		return fmt.Errorf("invalid y: %d", d)
	}

	d.data[x][y] = data
	return nil
}

func (d *Display) Get(x, y uint8) (byte, error) {
	if x >= width {
		return 0, fmt.Errorf("invalid x: %d", x)
	}

	if y >= height {
		return 0, fmt.Errorf("invalid y: %d", d)
	}

	return d.data[x][y], nil
}

func (d *Display) SetFlag() {
	d.update = true
}

func (d *Display) Update(screen *ebiten.Image) error {
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			d := d.data[x][y]
			for i := 7; i >= 0; i-- {
				c := color.White
				if (d>>i)&1 == 1 {
					c = color.Black
				}
				screen.Set((x+1)*8-i-1, y, c)
			}
		}
	}
	return nil
}

func (d *Display) Draw(screen *ebiten.Image) {
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			d := d.data[x][y]
			for i := 7; i >= 0; i-- {
				c := color.White
				if (d>>i)&1 == 1 {
					c = color.Black
				}
				screen.Set((x+1)*8-i-1, y, c)
			}
		}
	}
}

func (d *Display) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 64, 32
}

func (d *Display) Run() error {
	ebiten.SetWindowSize(64, 32)
	ebiten.SetWindowTitle("chip8")
	if err := ebiten.RunGame(d); err != nil {
		return err
	}
	return nil
}
