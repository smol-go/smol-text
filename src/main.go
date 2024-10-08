package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

// view mode and edit mode
var mode int

// variable for syntax highlighting
var highlight = 1

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

func write_file(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for row, line := range text_buffer {
		new_line := "\n"
		if row == len(text_buffer)-1 {
			new_line = ""
		}
		write_line := string(line) + new_line
		_, err = writer.WriteString(write_line)
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
	writer.Flush()
	modified = false
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

func copy_line() {
	copy_line := make([]rune, len(text_buffer[CURRENT_ROW]))
	copy(copy_line, text_buffer[CURRENT_ROW])
	copy_buffer = copy_line
}

func cut_line() {
	copy_line()
	if CURRENT_ROW >= len(text_buffer) || len(text_buffer) < 2 {
		return
	}

	new_text_buffer := make([][]rune, len(text_buffer)-1)

	copy(new_text_buffer[:CURRENT_ROW], text_buffer[:CURRENT_ROW])
	copy(new_text_buffer[CURRENT_ROW:], text_buffer[CURRENT_ROW+1:])

	text_buffer = new_text_buffer
	if CURRENT_ROW > 0 {
		CURRENT_ROW--
		CURRENT_COL = 0
	}
}

func paste_line() {
	if len(copy_buffer) == 0 {
		CURRENT_ROW++
		CURRENT_COL = 0
	}

	new_text_buffer := make([][]rune, len(text_buffer)+1)
	copy(new_text_buffer[:CURRENT_ROW], text_buffer[:CURRENT_ROW])
	new_text_buffer[CURRENT_ROW] = copy_buffer
	copy(new_text_buffer[CURRENT_ROW+1:], text_buffer[CURRENT_ROW:])
	text_buffer = new_text_buffer
}

func push_buffer() {
	copy_undo_buffer := make([][]rune, len(text_buffer))
	copy(copy_undo_buffer, text_buffer)
	undo_buffer = copy_undo_buffer
}

func pull_buffer() {
	if len(undo_buffer) == 0 {
		return
	}
	text_buffer = undo_buffer
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

func highlight_keyword(keyword string, col, row int) {
	for i := 0; i < len(keyword); i++ {
		ch := text_buffer[row+OFFSET_ROW][col+OFFSET_COL+i]
		termbox.SetCell(col+i, row, ch, termbox.ColorWhite|termbox.AttrBold, termbox.ColorDefault)
	}
}

func highlight_string(col, row int) int {
	i := 0
	for {
		if col+OFFSET_COL+i == len(text_buffer[row+OFFSET_ROW]) {
			return i - 1
		}
		ch := text_buffer[row+OFFSET_ROW][col+OFFSET_COL+i]
		if ch == '"' || ch == '\'' {
			termbox.SetCell(col+i, row, ch, termbox.ColorYellow, termbox.ColorDefault)
			return i
		} else {
			termbox.SetCell(col+i, row, ch, termbox.ColorYellow, termbox.ColorDefault)
			i++
		}
	}
}

func highlight_comment(col, row int) int {
	i := 0
	for {
		if col+OFFSET_COL+i == len(text_buffer[row+OFFSET_ROW]) {
			return i - 1
		}
		ch := text_buffer[row+OFFSET_ROW][col+OFFSET_COL+i]
		termbox.SetCell(col+i, row, ch, termbox.ColorMagenta|termbox.AttrBold, termbox.ColorDefault)
		i++
	}
}

func highlight_syntax(col *int, row, text_buffer_col, text_buffer_row int) {
	ch := text_buffer[text_buffer_row][text_buffer_col]
	next_token := string(text_buffer[text_buffer_row][text_buffer_col:])

	if unicode.IsDigit(ch) {
		termbox.SetCell(*col, row, ch, termbox.ColorYellow|termbox.AttrBold, termbox.ColorDefault)
	} else if ch == '\'' {
		termbox.SetCell(*col, row, ch, termbox.ColorYellow, termbox.ColorDefault)
		*col++
		*col += highlight_string(*col, row)
	} else if ch == '"' {
		termbox.SetCell(*col, row, ch, termbox.ColorYellow, termbox.ColorDefault)
		*col++
		*col += highlight_string(*col, row)
	} else if strings.Contains("+-*><=%&|^!:", string(ch)) {
		termbox.SetCell(*col, row, ch, termbox.ColorMagenta|termbox.AttrBold, termbox.ColorDefault)
	} else if ch == '/' {
		termbox.SetCell(*col, row, ch, termbox.ColorMagenta|termbox.AttrBold, termbox.ColorDefault)
		index := strings.Index(next_token, "//")
		if index == 0 {
			*col += highlight_comment(*col, row)
		}
	} else if ch == '#' {
		termbox.SetCell(*col, row, ch, termbox.ColorMagenta|termbox.AttrBold, termbox.ColorDefault)
		*col += highlight_comment(*col, row)
	} else {
		for _, token := range []string{
			"false", "False", "NaN", "None",
			"bool", "break", "byte",
			"case", "catch", "class", "const", "continue",
			"def", "do", "double", "as",
			"elif", "else", "enum", "eval", "except", "exec", "exit", "export", "extends", "extern",
			"finally", "float", "for", "from", "func", "function",
			"global",
			"if", "import", "in", "int", "is",
			"lambda",
			"nil", "not", "null",
			"pass", "print",
			"raise", "return",
			"self", "short", "signed", "sizeof", "static", "struct", "switch",
			"this", "throw", "throws", "true", "True", "try", "typedef", "typeof",
			"undefined", "union", "unsigned", "until",
			"var", "void",
			"while", "with", "yield",
		} {
			index := strings.Index(next_token, token+" ")
			if index == 0 {
				highlight_keyword(token, *col, row)
				*col += len(token)
				break
			} else {
				termbox.SetCell(*col, row, ch, termbox.ColorDefault, termbox.ColorDefault)
			}
		}
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
					if highlight == 1 {
						highlight_syntax(&col, row, text_buffer_col, text_buffer_row)
					} else {
						termbox.SetCell(col, row, text_buffer[text_buffer_row][text_buffer_col],
							termbox.ColorDefault, termbox.ColorDefault)
					}
				} else {
					termbox.SetCell(col, row, rune(' '), termbox.ColorDefault, termbox.ColorGreen)
				}
			} else if row+OFFSET_ROW > len(text_buffer)-1 {
				termbox.SetCell(0, row, rune('*'), termbox.ColorBlue, termbox.ColorDefault)
			}
		}

		if row == CURRENT_ROW-OFFSET_ROW && highlight == 1 {
			for col = 0; col < COLS; col++ {
				current_cell := termbox.GetCell(col, row)
				termbox.SetCell(col, row, current_cell.Ch, termbox.ColorDefault, termbox.ColorBlue)
			}
		}

		termbox.SetChar(col, row, rune('\n'))
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
			case 'w':
				write_file(source_file)
			case 'c':
				copy_line()
			case 'p':
				paste_line()
			case 'd':
				cut_line()
			case 's':
				push_buffer()
			case 'l':
				pull_buffer()
			case 'h':
				highlight ^= 1
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
