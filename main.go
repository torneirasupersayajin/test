package password

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"
)

var (
	ErrOperationCanceled = errors.New("Operation was canceled by the user")
	ErrNotTTY            = errors.New("The input device is not a TTY")
)

type Password struct {
	Mask      rune
	MinLen    int
	MaxLen    int
	Skippable bool
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
	ans := make([]byte, 0, p.MinLen)
	buf := make([]byte, 0, 1)
	for {
		_, err := os.Stdin.Read(buf)
		if err != nil {
			return nil, err
		}
		b := buf[0]
		if b == 13 {
			break
		}
		if b == 8 || b == 127 {
			if len(ans) > 0 {
				ans = ans[:len(ans)-1]
				fmt.Print("\b \b")
			}
			continue
		}
		if b == 3 || b == 24 || b == 27 {
			if p.Skippable {
				return nil, nil
			}
			return nil, ErrOperationCanceled
		}
		if p.Mask != 0 {
			fmt.Print(string(p.Mask))
		}
		ans = append(ans, b)
	}
	return ans, nil
}
