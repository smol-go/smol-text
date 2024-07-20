package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

// view mode and edit mode
var mode int

// to store the width and height of the terminal
var ROWS, COLS int

// to store the scoll
var OFFSET_COL, OFFSET_ROW int

// current cursor position
var CURRENT_COL, CURRENT_ROW int

var text_buffer = [][]rune{}
var undo_buffer = [][]rune{}
var copy_buffer = []rune{}

var source_file string

var modified bool

func print_message(row, col int, foreground, background termbox.Attribute, msg string) {
	for _, ch := range msg {
		termbox.SetCell(row, col, ch, foreground, background)
		row += runewidth.RuneWidth(ch)
	}
}

func read_file(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		// if the file name given by the user exits, open the file
		// if not, create a new file with that file name
		source_file = filename
		text_buffer = append(text_buffer, []rune{})
		return
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	line_number := 0

	for scanner.Scan() {
		line := scanner.Text()
		text_buffer = append(text_buffer, []rune{})

		for i := 0; i < len(line); i++ {
			text_buffer[line_number] = append(text_buffer[line_number], rune(line[i]))
		}
		line_number++
	}
	if line_number == 0 {
		text_buffer = append(text_buffer, []rune{})
	}
}

func insert_rune(event termbox.Event) {
	insert_rune := make([]rune, len(text_buffer[CURRENT_ROW])+1)
	copy(insert_rune[:CURRENT_COL], text_buffer[CURRENT_ROW][:CURRENT_COL])

	if event.Key == termbox.KeySpace {
		insert_rune[CURRENT_COL] = rune(' ')
	} else if event.Key == termbox.KeyTab {
		insert_rune[CURRENT_COL] = rune(' ')
	} else {
		insert_rune[CURRENT_COL] = rune(event.Ch)
	}

	copy(insert_rune[CURRENT_COL+1:], text_buffer[CURRENT_ROW][CURRENT_COL:])
	text_buffer[CURRENT_ROW] = insert_rune
	CURRENT_COL++
}

func delete_rune() {
	if CURRENT_COL > 0 {
		CURRENT_COL--
		delete_line := make([]rune, len(text_buffer[CURRENT_ROW])-1)

		copy(delete_line[:CURRENT_COL], text_buffer[CURRENT_ROW][:CURRENT_COL])
		copy(delete_line[CURRENT_COL:], text_buffer[CURRENT_ROW][CURRENT_COL+1:])

		text_buffer[CURRENT_ROW] = delete_line
	} else if CURRENT_ROW > 0 {
		append_line := make([]rune, len(text_buffer[CURRENT_ROW]))

		copy(append_line, text_buffer[CURRENT_ROW][CURRENT_COL:])

		new_text_buffer := make([][]rune, len(text_buffer)-1)

		copy(new_text_buffer[:CURRENT_ROW], text_buffer[:CURRENT_ROW])
		copy(new_text_buffer[CURRENT_ROW:], text_buffer[CURRENT_ROW+1:])

		text_buffer = new_text_buffer
		CURRENT_ROW--
		CURRENT_COL = len(text_buffer[CURRENT_ROW])
		insert_line := make([]rune, len(text_buffer[CURRENT_ROW])+len(append_line))

		copy(insert_line[:len(text_buffer[CURRENT_ROW])], text_buffer[CURRENT_ROW])
		copy(insert_line[len(text_buffer[CURRENT_ROW]):], append_line)

		text_buffer[CURRENT_ROW] = insert_line
	}
}

func insert_line() {
	right_line := make([]rune, len(text_buffer[CURRENT_ROW][CURRENT_COL:]))

	copy(right_line, text_buffer[CURRENT_ROW][CURRENT_COL:])

	left_line := make([]rune, len(text_buffer[CURRENT_ROW][:CURRENT_COL]))

	copy(left_line, text_buffer[CURRENT_ROW][:CURRENT_COL])

	text_buffer[CURRENT_ROW] = left_line
	CURRENT_ROW++
	CURRENT_COL = 0
	new_text_buffer := make([][]rune, len(text_buffer)+1)

	copy(new_text_buffer, text_buffer[:CURRENT_ROW])

	new_text_buffer[CURRENT_ROW] = right_line

	copy(new_text_buffer[CURRENT_ROW+1:], text_buffer[CURRENT_ROW:])

	text_buffer = new_text_buffer
}

func scroll_text_buffer() {
	if CURRENT_ROW < OFFSET_ROW {
		OFFSET_ROW = CURRENT_ROW
	}
	if CURRENT_COL < OFFSET_COL {
		OFFSET_COL = CURRENT_COL
	}
	if CURRENT_ROW >= OFFSET_ROW+ROWS {
		OFFSET_ROW = CURRENT_ROW - ROWS + 1
	}
	if CURRENT_COL >= OFFSET_COL+COLS {
		OFFSET_COL = CURRENT_COL - COLS + 1
	}
}

func display_text_buffer() {
	var row, col int
	for row = 0; row < ROWS; row++ {
		text_buffer_row := row + OFFSET_ROW
		for col = 0; col < COLS; col++ {
			text_buffer_col := col + OFFSET_COL
			if text_buffer_row >= 0 && text_buffer_row < len(text_buffer) && text_buffer_col < len(text_buffer[text_buffer_row]) {
				if text_buffer[text_buffer_row][text_buffer_col] != '\t' {
					termbox.SetChar(col, row, text_buffer[text_buffer_row][text_buffer_col])
				} else {
					termbox.SetCell(col, row, rune(' '), termbox.ColorDefault, termbox.ColorGreen)
				}
			} else if row+OFFSET_ROW > len(text_buffer)-1 {
				termbox.SetCell(0, row, rune('*'), termbox.ColorBlue, termbox.ColorDefault)
			}
			termbox.SetChar(col, row, rune('\n'))
		}
	}
}

func display_status_bar() {
	var mode_status string
	var file_status string
	var copy_status string
	var undo_status string
	var cursor_status string

	if mode > 0 {
		mode_status = " EDIT: "
	} else {
		mode_status = " VIEW: "
	}

	filename_length := len(source_file)
	if filename_length > 8 {
		filename_length = 8
	}

	file_status = source_file[:filename_length] + " - " + strconv.Itoa(len(text_buffer)) + " lines"

	if modified {
		file_status += " modified"
	} else {
		file_status += " saved"
	}

	cursor_status = " Row " + strconv.Itoa(CURRENT_ROW+1) + ", Col " + strconv.Itoa(CURRENT_COL+1) + " "

	if len(copy_buffer) > 0 {
		copy_status = " [Copy]"
	}
	if len(undo_buffer) > 0 {
		undo_status = " [Undo]"
	}

	used_space := len(mode_status) + len(file_status) + len(cursor_status) + len(copy_status) + len(undo_status)
	spaces := strings.Repeat(" ", COLS-used_space)

	message := mode_status + file_status + copy_status + undo_status + spaces + cursor_status
	print_message(0, ROWS, termbox.ColorBlack, termbox.ColorWhite, message)
}

func get_key() termbox.Event {
	var key_event termbox.Event

	switch event := termbox.PollEvent(); event.Type {
	case termbox.EventKey:
		key_event = event
	case termbox.EventError:
		panic(event.Err)
	}

	return key_event
}

func process_keypress() {
	key_event := get_key()

	if key_event.Key == termbox.KeyEsc {
		mode = 0
	} else if key_event.Ch != 0 {
		if mode == 1 {
			insert_rune(key_event)
			modified = true
		} else {
			switch key_event.Ch {
			case 'q':
				termbox.Close()
				os.Exit(0)
			case 'e':
				mode = 1
			}
		}
	} else {
		switch key_event.Key {
		case termbox.KeyEnter:
			if mode == 1 {
				insert_line()
				modified = true
			}
		case termbox.KeyBackspace:
			delete_rune()
			modified = true
		case termbox.KeyBackspace2:
			delete_rune()
			modified = true
		case termbox.KeyTab:
			if mode == 1 {
				for i := 0; i < 4; i++ {
					insert_rune(key_event)
				}
				modified = true
			}
		case termbox.KeySpace:
			if mode == 1 {
				insert_rune(key_event)
				modified = true
			}
		case termbox.KeyHome:
			CURRENT_COL = 0
		case termbox.KeyEnd:
			CURRENT_COL = len(text_buffer[CURRENT_ROW])
		case termbox.KeyPgup:
			if CURRENT_ROW-int(ROWS/4) > 0 {
				CURRENT_ROW -= int(ROWS / 4)
			}
		case termbox.KeyPgdn:
			if CURRENT_ROW+int(ROWS/4) < len(text_buffer)-1 {
				CURRENT_ROW += int(ROWS / 4)
			}
		case termbox.KeyArrowUp:
			if CURRENT_ROW != 0 {
				CURRENT_ROW--
			}
		case termbox.KeyArrowDown:
			if CURRENT_ROW < len(text_buffer)-1 {
				CURRENT_ROW++
			}
		case termbox.KeyArrowLeft:
			if CURRENT_COL != 0 {
				CURRENT_COL--
			} else if CURRENT_ROW > 0 {
				CURRENT_ROW--
				CURRENT_COL = len(text_buffer[CURRENT_ROW])
			}
		case termbox.KeyArrowRight:
			if CURRENT_COL < len(text_buffer[CURRENT_ROW]) {
				CURRENT_COL++
			} else if CURRENT_ROW < len(text_buffer)-1 {
				CURRENT_ROW++
				CURRENT_COL = 0
			}
		}

		if CURRENT_COL > len(text_buffer[CURRENT_ROW]) {
			CURRENT_COL = len(text_buffer[CURRENT_ROW])
		}
	}
}

func run_editor() {
	err := termbox.Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		source_file = os.Args[1]
	} else {
		source_file = "output.txt"
		text_buffer = append(text_buffer, []rune{})
	}

	for {
		// asciiArt := `
		// 	███████╗███╗   ███╗ ██████╗ ██╗  ████████╗███████╗██╗  ██╗████████╗
		// 	██╔════╝████╗ ████║██╔═══██╗██║  ╚══██╔══╝██╔════╝╚██╗██╔╝╚══██╔══╝
		// 	███████╗██╔████╔██║██║   ██║██║     ██║   █████╗   ╚███╔╝    ██║
		// 	╚════██║██║╚██╔╝██║██║   ██║██║     ██║   ██╔══╝   ██╔██╗    ██║
		// 	███████║██║ ╚═╝ ██║╚██████╔╝███████╗██║   ███████╗██╔╝ ██╗   ██║
		// 	╚══════╝╚═╝     ╚═╝ ╚═════╝ ╚══════╝╚═╝   ╚══════╝╚═╝  ╚═╝   ╚═╝
		// `

		// lines := strings.Split(asciiArt, "\n")

		// maxWidth := 0
		// for _, line := range lines {
		// 	width := runewidth.StringWidth(line)
		// 	if width > maxWidth {
		// 		maxWidth = width
		// 	}
		// }

		// message := "A simple text editor written in golang"

		// // calculating the starting x position for message
		// startX := (maxWidth - runewidth.StringWidth(message)) / 2

		// for i, line := range lines {
		// 	print_message(0, i, termbox.ColorCyan, termbox.ColorDefault, line)
		// }

		// print_message(startX, len(lines), termbox.ColorCyan, termbox.ColorDefault, message)

		COLS, ROWS = termbox.Size()
		ROWS--
		if COLS < 78 {
			COLS = 78
		}
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		scroll_text_buffer()
		display_text_buffer()
		display_status_bar()
		termbox.SetCursor(CURRENT_COL-OFFSET_COL, CURRENT_ROW-OFFSET_ROW)
		termbox.Flush()
		process_keypress()
	}
}

func main() {
	run_editor()
}
