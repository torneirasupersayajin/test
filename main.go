package password

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"
)

var (
	ErrOperationCanceled    = errors.New("Operation was canceled by the user")
	ErrOperationInterrupted = errors.New("Operation was interrupted by the user")
	ErrNotTTY               = errors.New("The input device is not a TTY")
)

type RenderConfig struct {
}

type Password struct {
	Message   string
	Mask      rune
	Skippable bool
	RenderCfg any
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
	buf := make([]byte, 1)
	for {
		_, err := os.Stdin.Read(buf)
		if err != nil {
			return nil, err
		}
		b := buf[0]
		if b == 13 {
			fmt.Print("\r\n")
			return ans, nil
		}
		if b == 8 || b == 127 {
			if len(ans) > 0 {
				ans = ans[:len(ans)-1]
				fmt.Print("\b \b")
			}
			continue
		}
		if b == 3 || b == 27 {
			fmt.Print("\033[0;31mskipped")
			if p.Skippable {
				return nil, nil
			}
			if b == 27 {
				return nil, ErrOperationCanceled
			}
			return nil, ErrOperationInterrupted
		}
		if p.Mask != 0 {
			fmt.Print(string(p.Mask))
		}
		ans = append(ans, b)
	}
	panic("unreachable")
}
