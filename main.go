package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	c := NewChip8()
	if len(os.Args) != 2 {
		panic("args should be 2")
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println(err)
		return
	}
	c.LoadProgram(data)
	c.Run()
}
