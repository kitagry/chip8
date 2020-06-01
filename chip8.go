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
	i     uint16
	sp    uint
	stack [16]uint16

	dt uint8
	st uint8
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
			err := c.cycle()
			if err != nil {
				panic(err)
			}
			time.Sleep(300 * time.Microsecond)
		}
	}()
	c.display.Run()
}

type pcOrder interface {
	newPC(cur uint16) uint16
}

type Next struct {
}

func (n *Next) newPC(cur uint16) uint16 {
	return cur + 2
}

var next = &Next{}

type Skip struct {
}

func (n *Skip) newPC(cur uint16) uint16 {
	return cur + 4
}

var skip = &Skip{}

type Jump struct {
	to uint16
}

func NewJump(to uint16) pcOrder {
	return &Jump{to: to}
}

func (j *Jump) newPC(cur uint16) uint16 {
	return j.to
}

func (c *Chip8) instruct(opcode uint16) (pcOrder, error) {
	x := (opcode & 0x0F00) >> 8
	y := (opcode & 0x00F0) >> 4
	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode {
		case 0x00E0:
			c.display.Clear()
			return next, nil
		case 0x00EE:
			c.sp--
			newPc := c.stack[c.sp]
			return NewJump(newPc), nil
		default:
			return next, fmt.Errorf("undefined opcode: %x", opcode)
		}
	case 0X1000:
		return NewJump(opcode & 0x0FFF), nil
	case 0x2000:
		c.stack[c.sp] = c.pc + 2
		c.sp++
		return NewJump(opcode & 0x0FFF), nil
	case 0x3000:
		if c.v[x] == uint8(opcode&0x00FF) {
			return skip, nil
		}
		return next, nil
	case 0x4000:
		if c.v[x] != uint8(opcode&0x00FF) {
			return skip, nil
		}
		return next, nil
	case 0x5000:
		if c.v[x] == c.v[y] {
			return skip, nil
		}
		return next, nil
	case 0x6000:
		c.v[x] = uint8(opcode & 0x00FF)
		return next, nil
	case 0x7000:
		c.v[x] += uint8(opcode & 0x00FF)
		return next, nil
	case 0x8000:
		switch opcode & 0x000F {
		case 0x0000:
			c.v[x] = c.v[y]
		case 0x0001:
			c.v[x] |= c.v[y]
		case 0x0002:
			c.v[x] &= c.v[y]
		case 0x0003:
			c.v[x] ^= c.v[y]
		case 0x0004:
			if c.v[y] > (0xFF - c.v[x]) {
				c.v[0xF] = 1
			} else {
				c.v[0xF] = 0
			}
			c.v[x] += c.v[y]
		case 0x0005:
			if c.v[x] > c.v[y] {
				c.v[0xF] = 1
			} else {
				c.v[0xF] = 0
			}
			c.v[x] -= c.v[y]
		case 0x0006:
			c.v[0xF] = c.v[x] & 0x1
			c.v[x] >>= 1
		case 0x0007:
			if c.v[x] < c.v[y] {
				c.v[0xF] = 1
			} else {
				c.v[0xF] = 0
			}
			c.v[x] = c.v[y] - c.v[x]
		case 0x000E:
			if c.v[x]&0x80 > 0 {
				c.v[0xF] = 1
			} else {
				c.v[0xF] = 0
			}
			c.v[x] <<= 1
		default:
			return next, fmt.Errorf("unknown opcode: %x", opcode)
		}
		return next, nil
	case 0x9000:
		if c.v[x] != c.v[y] {
			return skip, nil
		}
		return next, nil
	case 0XA000:
		c.i = opcode & 0x0FFF
		return next, nil
	case 0XB000:
		return NewJump(uint16(c.v[0]) + opcode&0x0FFF), nil
	case 0xC000:
		c.v[x] = uint8(rand.Uint32() & uint32(opcode) & 0x00FF)
		return next, nil
	case 0xD000:
		x := c.v[x]
		y := c.v[y]
		height := opcode & 0x000F
		c.v[0xF] = 0
		for yLine := 0; yLine < int(height); yLine++ {
			pixel, err := c.mem.Fetch(int(c.i) + yLine)
			if err != nil {
				return next, xerrors.Errorf("failed to fetch data from memory: %w", err)
			}

			for xLine := 0; xLine < 8; xLine++ {
				if (pixel & 0x80) > 0 {
					isSet, err := c.display.Set(x+uint8(xLine), y+uint8(yLine))
					if err != nil {
						return next, xerrors.Errorf("failed to set display buffer(%d, %d): %w", x+uint8(xLine), y+uint8(yLine), err)
					}

					if isSet {
						c.v[0xF] = 1
					}
				}
				pixel <<= 1
			}
		}
		c.display.SetFlag()
		return next, nil
	case 0xE000:
		switch opcode & 0x00FF {
		case 0x009E:
			if c.v[x] < 16 && c.display.keys[c.v[x]] {
				return skip, nil
			}
			return next, nil
		case 0x00A1:
			if c.v[x] < 16 && !c.display.keys[c.v[x]] {
				return skip, nil
			}
			return next, nil
		default:
			return next, fmt.Errorf("unknown opcode: %x", opcode)
		}
	case 0xF000:
		switch opcode & 0x00FF {
		case 0x0007:
			c.v[x] = c.dt
		case 0x000A:
			for i := 0; i < len(c.display.keys); i++ {
				if c.display.keys[i] {
					c.v[x] = uint8(i)
					return next, nil
				}
			}
			return NewJump(c.pc), nil
		case 0x0015:
			c.dt = c.v[x]
		case 0x0018:
			c.st = c.v[x]
		case 0x001E:
			c.i += uint16(c.v[x])
		case 0x0029:
			c.i = uint16(c.v[x]) * 5
		case 0x0033:
			num := c.v[x] % 10
			for i := 3; i > 0; i-- {
				err := c.mem.Set(int(c.i)+i-1, uint8(num%10))
				if err != nil {
					fmt.Println(err)
				}
				num /= 10
			}
		case 0x0055:
			for i := 0; i <= int(x); i++ {
				c.mem.Set(int(c.i)+i, c.v[i])
			}
		case 0x0065:
			var err error
			for i := 0; i <= int(x); i++ {
				c.v[i], err = c.mem.Fetch(int(c.i) + i)
				if err != nil {
					fmt.Println(err)
					err = nil
				}
			}
		default:
			fmt.Printf("unknown opcode: %x\n", opcode)
		}
		return next, nil
	default:
		return next, fmt.Errorf("unknown opcode: %x", opcode)
	}
}

func (c *Chip8) cycle() error {
	opcode, err := c.mem.Fetch16(int(c.pc))
	if err != nil {
		return xerrors.Errorf("failed to fetch memory: %w", err)
	}

	pcord, err := c.instruct(opcode)
	if err != nil {
		return xerrors.Errorf("failed to order: %w", err)
	}
	c.pc = pcord.newPC(c.pc)

	if c.dt > 0 {
		c.dt--
	}

	if c.st > 0 {
		if c.st == 1 {
			fmt.Println("beep!")
			c.st--
		}
	}

	return nil
}
