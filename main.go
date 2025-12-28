package password

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"golang.org/x/term"
)

var (
	ErrOperationInterrupted = errors.New("Operation was interrupted by the user")
	ErrOperationCanceled    = errors.New("Operation was canceled by the user")
	ErrNotTTY               = errors.New("The input device is not a TTY")
)

type Password struct {
	Message                string
	StartsVisible          bool
	EnableVisibilityToggle bool
	Mask                   rune
	Skippable              bool
	EnableArrowsNavigation bool
}

func (p *Password) SetMask(mask rune) {
	p.Mask = mask
}

func (p Password) Prompt() ([]byte, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil, ErrNotTTY
	}
	state, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	defer term.Restore(fd, state)
	fmt.Print(p.Message)
	answer := make([]byte, 0)
	buf := make([]byte, 6)
	isVisible := p.StartsVisible
	cursorPosition := 0
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return nil, err
		}
		seq := buf[:n]
		if slices.Equal(seq, []byte{8}) || slices.Equal(seq, []byte{27, 91, 51, 126}) {
			if len(answer) > 0 {
				fmt.Print("\b \b")
				answer = answer[:len(answer)-1]
			}
			continue
		}
		if p.EnableArrowsNavigation {
			if slices.Equal(seq, []byte{27, 91, 68}) {
				if cursorPosition > 0 {
					fmt.Print("\x1b[D")
					cursorPosition--
				}
				continue
			}
			if slices.Equal(seq, []byte{27, 91, 67}) {
				if cursorPosition < len(answer) {
					fmt.Print("\x1b[C")
					cursorPosition++
				}
				continue
			}
		}
		if n > 1 {
			continue
		}
		b := seq[0]
		if b == 13 {
			fmt.Print("\r\n")
			return answer, nil
		}
		if b == 3 || b == 27 {
			if p.Skippable {
				return nil, nil
			}
			fmt.Print("\r\n")
			if b == 27 {
				return nil, ErrOperationCanceled
			}
			return nil, ErrOperationInterrupted
		}
		if b == 18 {
			if p.EnableVisibilityToggle {
				fmt.Print(strings.Repeat("\b \b", len(answer)))
				if isVisible {
					fmt.Print(strings.Repeat(string(p.Mask), len(answer)))
				} else {
					fmt.Print(string(answer))
				}
				isVisible = !isVisible
			}
			continue
		}
		if isVisible {
			fmt.Print(string(b))
		} else {
			fmt.Print(string(p.Mask))
		}
		answer = append(answer[:cursorPosition], b)
		cursorPosition++
	}
	panic("unreachable")
}
