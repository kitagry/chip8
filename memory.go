package main

import "fmt"

const memSize = 4096

type Memory struct {
	data [memSize]byte
}

func NewMemory() *Memory {
	return &Memory{}
}

func (m *Memory) Fetch(address int) (uint16, error) {
	if address < 0 || address >= memSize-1 {
		return 0, fmt.Errorf("invalid address: %d", address)
	}

	return uint16(m.data[address])<<8 | uint16(m.data[address+1]), nil
}

func (m *Memory) Set(address int, data byte) error {
	if address < 0 || address >= memSize {
		return fmt.Errorf("invalid address: %d", address)
	}
	m.data[address] = data
	return nil
}
