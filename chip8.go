package main

import (
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/xerrors"
)

var chip8Fontset = []byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

type Chip8 struct {
	mem     *Memory
	display *Display

	pc    uint16
	v     [16]uint8
	l     uint16
	sp    uint
	stack [16]uint16
}

func NewChip8() *Chip8 {
	mem := NewMemory()
	display := NewDisplay()

	var err error
	for i, f := range chip8Fontset {
		err = mem.Set(i, f)
		if err != nil {
			panic("Failed to set mem")
		}
	}

	return &Chip8{
		mem:     mem,
		display: display,
		pc:      0x200,
	}
}

func (c *Chip8) LoadProgram(data []byte) error {
	var err error
	for i, d := range data {
		err = c.mem.Set(i+0x200, d)
		if err != nil {
			return xerrors.Errorf("failed to set memory: %w", err)
		}
	}
	return nil
}

func (c *Chip8) Run() {
	go func() {
		for {
			c.cycle()
			time.Sleep(17 * time.Millisecond)
		}
	}()
	c.display.Run()
}

func (c *Chip8) cycle() error {
	opcode, err := c.mem.Fetch(int(c.pc))
	if err != nil {
		return xerrors.Errorf("failed to fetch memory: %w", err)
	}

	switch opcode & 0xF000 {
	case 0x0000:
		// TODO
	case 0X1000:
		c.pc = opcode & 0x0FFF
	case 0x2000:
		c.stack[c.sp] = c.pc
		c.sp++
		c.pc = opcode & 0x0FFF
	case 0x3000:
		if c.v[(opcode&0x0F00)>>8] == uint8(opcode&0x00FF) {
			c.pc += 4
			return nil
		}
		c.pc += 2
	case 0x4000:
		if c.v[(opcode&0x0F00)>>8] != uint8(opcode&0x00FF) {
			c.pc += 4
			return nil
		}
		c.pc += 2
	case 0x5000:
		if c.v[(opcode&0x0F00)>>8] == c.v[(opcode&0x00F0)>>4] {
			c.pc += 4
			return nil
		}
		c.pc += 2
	case 0x6000:
		c.v[(opcode&0x0F00)>>8] = uint8(opcode & 0x00FF)
		c.pc += 2
	case 0x7000:
		c.v[(opcode&0x0F00)>>8] += uint8(opcode & 0x00FF)
		c.pc += 2
	case 0x8000:
		x := (opcode & 0x0F00) >> 8
		y := (opcode & 0x00F0) >> 4
		switch opcode & 0x000F {
		case 0x0001:
			c.v[x] |= c.v[y]
		case 0x0002:
			c.v[x] &= c.v[y]
		case 0x0003:
			c.v[x] ^= c.v[y]
		case 0x0004:
			c.v[x] += c.v[y]
		case 0x0005:
			c.v[x] -= c.v[y]
		case 0x0006:
			c.v[x] >>= 1
		case 0x0007:
			c.v[x] = c.v[y] - c.v[x]
		case 0x000E:
			c.v[x] <<= 1
		default:
			return fmt.Errorf("unknown opcode: %x", opcode)
		}
		c.pc += 2
	case 0x9000:
		if c.v[(opcode&0x0F00)>>8] != c.v[(opcode&0x00F0)>>4] {
			c.pc += 4
			return nil
		}
		c.pc += 2
	case 0XA000:
		c.l = opcode & 0x0FFF
		c.pc += 2
	case 0XB000:
		c.pc = uint16(c.v[0]) + opcode&0x0FFF
		c.pc += 2
	case 0xC000:
		c.v[(opcode&0x0F00)>>8] = uint8(rand.Uint32() & uint32(opcode) & 0x00FF)
		c.pc += 2
	case 0xD000:
		fmt.Println("pixel")
		c.pc += 2
		x := c.v[(opcode&0x0F00)>>8]
		y := c.v[(opcode&0x00F0)>>4]
		height := opcode & 0x000F
		c.v[0xF] = 0
		for yLine := 0; yLine < int(height); yLine++ {
			pixel, err := c.mem.Fetch(int(c.l) + yLine)
			if err != nil {
				return xerrors.Errorf("failed to fetch data from memory: %w", err)
			}

			for xLine := 0; xLine < 8; xLine++ {
				if (pixel & (0x80 >> xLine)) != 0 {
					gfx, err := c.display.Get(x+uint8(xLine), y+uint8(yLine))
					if err != nil {
						return xerrors.Errorf("failed to get display buffer: %w", err)
					}

					if gfx == 1 {
						c.v[0xF] = 1
						err := c.display.Set(x+uint8(xLine), y+uint8(yLine), gfx^1)
						if err != nil {
							return xerrors.Errorf("failed to set display buffer: %w", err)
						}
					}
				}
			}
		}
		c.display.SetFlag()
	case 0xF000:
		// TODO
	default:
		return fmt.Errorf("unknown opcode: %x", opcode)
	}

	return nil
}
