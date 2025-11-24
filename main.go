package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		filename := os.Args[1]
		fmt.Printf("Opening file: %s\n", filename)
		// TODO: Implement file opening
	} else {
		fmt.Println("smol-text v0.1.0")
		fmt.Println("Usage: smol-text ")
	}
}
