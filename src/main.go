package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

func print_message(x, y int, foreground, background termbox.Attribute, msg string) {
	for _, ch := range msg {
		termbox.SetCell(x, y, ch, foreground, background)
		x += runewidth.RuneWidth(ch)
	}
}

func run_editor() {
	err := termbox.Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	asciiArt := `
		███████╗███╗   ███╗ ██████╗ ██╗  ████████╗███████╗██╗  ██╗████████╗
		██╔════╝████╗ ████║██╔═══██╗██║  ╚══██╔══╝██╔════╝╚██╗██╔╝╚══██╔══╝
		███████╗██╔████╔██║██║   ██║██║     ██║   █████╗   ╚███╔╝    ██║   
		╚════██║██║╚██╔╝██║██║   ██║██║     ██║   ██╔══╝   ██╔██╗    ██║   
		███████║██║ ╚═╝ ██║╚██████╔╝███████╗██║   ███████╗██╔╝ ██╗   ██║   
		╚══════╝╚═╝     ╚═╝ ╚═════╝ ╚══════╝╚═╝   ╚══════╝╚═╝  ╚═╝   ╚═╝ 
	`

	lines := strings.Split(asciiArt, "\n")

	maxWidth := 0
	for _, line := range lines {
		width := runewidth.StringWidth(line)
		if width > maxWidth {
			maxWidth = width
		}
	}

	message := "A simple text editor written in golang"

	// calculating the starting x position for message
	startX := (maxWidth - runewidth.StringWidth(message)) / 2

	for i, line := range lines {
		print_message(0, i, termbox.ColorCyan, termbox.ColorDefault, line)
	}

	print_message(startX, len(lines), termbox.ColorCyan, termbox.ColorDefault, message)

	termbox.Flush()
	termbox.PollEvent()
	termbox.Close()
}

func main() {
	run_editor()
}
