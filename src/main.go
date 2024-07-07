package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

// to store the width and height of the terminal
var ROWS, COLS int

// to store the scoll
var OFFSET_X, OFFSET_Y int

var text_buffer = [][]rune{
	{'h', 'e', 'l', 'l', 'o'},
	{'w', 'o', 'r', 'l', 'd'},
}

func print_message(row, col int, foreground, background termbox.Attribute, msg string) {
	for _, ch := range msg {
		termbox.SetCell(row, col, ch, foreground, background)
		row += runewidth.RuneWidth(ch)
	}
}

func display_text_buffer() {
	var row, col int
	for row = 0; row < ROWS; row++ {
		text_buffer_row := row + OFFSET_Y
		for col = 0; col < COLS; col++ {
			text_buffer_col := col + OFFSET_X
			if text_buffer_row >= 0 && text_buffer_row < len(text_buffer) && text_buffer_col < len(text_buffer[text_buffer_row]) {
				if text_buffer[text_buffer_row][text_buffer_col] != '\t' {
					termbox.SetChar(col, row, text_buffer[text_buffer_row][text_buffer_col])
				} else {
					termbox.SetCell(col, row, rune(' '), termbox.ColorDefault, termbox.ColorGreen)
				}
			} else if row+OFFSET_Y > len(text_buffer) {
				termbox.SetCell(0, row, rune('*'), termbox.ColorBlue, termbox.ColorDefault)
				termbox.SetChar(col, row, rune('\n'))
			}
		}
	}
}

func run_editor() {
	err := termbox.Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for {
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

		COLS, ROWS = termbox.Size()
		ROWS--
		if COLS < 78 {
			COLS = 78
		}
		// termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		// display_text_buffer()
		termbox.Flush()
		event := termbox.PollEvent()
		if event.Type == termbox.EventKey && event.Key == termbox.KeyEsc {
			termbox.Close()
			break
		}
	}
}

func main() {
	run_editor()
}
