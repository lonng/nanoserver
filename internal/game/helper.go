package game

import (
	"runtime"
	"strings"

	"github.com/lonng/nanoserver/pkg/errutil"
	"github.com/lonng/nanoserver/protocol"

	"github.com/lonng/nano/session"
)

const (
	ModeTrios = 3 // 三人模式
	ModeFours = 4 // 四人模式
)

func verifyOptions(opts *protocol.DeskOptions) bool {
	if opts == nil {
		return false
	}

	if opts.Mode != ModeTrios && opts.Mode != 4 {
		return false
	}

	if opts.MaxRound != 1 && opts.MaxRound != 4 && opts.MaxRound != 8 && opts.MaxRound != 16 {
		return false
	}

	return true
}

func requireCardCount(round int) int {
	if c, ok := consume[round]; ok {
		return c
	}

	c := 2
	switch round {
	case 8:
		c = 3
	case 16:
		c = 4
	}

	return c
}

func playerWithSession(s *session.Session) (*Player, error) {
	p, ok := s.Value(kCurPlayer).(*Player)
	if !ok {
		return nil, errutil.ErrPlayerNotFound
	}
	return p, nil
}

func stack() string {
	buf := make([]byte, 10000)
	n := runtime.Stack(buf, false)
	buf = buf[:n]

	s := string(buf)

	// skip nano frames lines
	const skip = 7
	count := 0
	index := strings.IndexFunc(s, func(c rune) bool {
		if c != '\n' {
			return false
		}
		count++
		return count == skip
	})
	return s[index+1:]
}
