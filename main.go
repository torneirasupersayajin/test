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
	ans := make([]byte, 0)
	buf := make([]byte, 6)
	isVisible := p.StartsVisible
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return nil, err
		}
		if slices.Equal(buf, []byte{8}) || slices.Equal(buf, []byte{27, 91, 51, 126}) {
			if len(ans) > 0 {
				fmt.Print("\b \b")
				ans = ans[:len(ans)-1]
			}
			continue
		}
		if n > 1 {
			continue
		}
		b := buf[0]
		if b == 13 {
			fmt.Print("\r\n")
			return ans, nil
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
				fmt.Print(strings.Repeat("\b \b", len(ans)))
				if isVisible {
					fmt.Print(strings.Repeat(string(p.Mask), len(ans)))
				} else {
					fmt.Print(string(ans))
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
		ans = append(ans, b)
	}
	panic("unreachable")
}
