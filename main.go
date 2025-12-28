package password

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

var (
	ErrOperationInterrupted = errors.New("Operation was interrupted by the user")
	ErrOperationCanceled    = errors.New("Operation was canceled by the user")
	ErrNotTTY               = errors.New("The input device is not a TTY")
)

type RenderConfig struct {
	PromptPrefix         rune
	AnsweredPromptPrefix rune
}

type Password struct {
	Msg                    string
	StartsVisible          bool
	EnableVisibilityToggle bool
	Mask                   rune
	AllowSkip              bool
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
	fmt.Print(p.Msg)
	ans := make([]byte, 0)
	buf := make([]byte, 6)
	visible := p.StartsVisible
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return nil, err
		}
		if n != 1 {
			continue
		}
		b := buf[0]
		if b == 13 {
			fmt.Print("\r\n")
			return ans, nil
		}
		if b == 3 || b == 27 {
			if p.AllowSkip {
				return nil, nil
			}
			fmt.Print("\r\n")
			if b == 27 {
				return nil, ErrOperationCanceled
			}
			return nil, ErrOperationInterrupted
		}
		if b == 8 || b == 127 {
			if len(ans) > 0 {
				fmt.Print("\b \b")
				ans = ans[:len(ans)-1]
			}
			continue
		}
		if b == 18 {
			if p.EnableVisibilityToggle {
				fmt.Print(strings.Repeat("\b \b", len(ans)))
				if visible {
					visible = false
					fmt.Print(strings.Repeat(string(p.Mask), len(ans)))
				} else {
					visible = true
					fmt.Print(string(ans))
				}
			}
			continue
		}
		if visible {
			fmt.Print(string(b))
		} else {
			fmt.Print(string(p.Mask))
		}
		ans = append(ans, b)
	}
	panic("unreachable")
}
