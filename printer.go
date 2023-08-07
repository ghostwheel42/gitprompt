package gitprompt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"unicode"
)

const (
	tAttribute rune = '@'
	tColor     rune = '#'
	tReset     rune = '_'
	tLeak      rune = '>'
	tData      rune = '%'
	tGroupOp   rune = '['
	tGroupCl   rune = ']'
	tEsc       rune = '\\'
)

var attrs = map[rune]uint8{
	'b': 1, // bold
	'f': 2, // faint
	'i': 3, // italic
}

var resetAttrs = map[rune]uint8{
	'B': 1, // bold
	'F': 2, // faint
	'I': 3, // italic
}

var colors = map[rune]uint8{
	'k': 30, // black
	'r': 31, // red
	'g': 32, // green
	'y': 33, // yellow
	'b': 34, // blue
	'm': 35, // magenta
	'c': 36, // cyan
	'w': 37, // white
	'K': 90, // highlight black
	'R': 91, // highlight red
	'G': 92, // highlight green
	'Y': 93, // highlight yellow
	'B': 94, // highlight blue
	'M': 95, // highlight magenta
	'C': 96, // highlight cyan
	'W': 97, // highlight white
}

const (
	head      rune = 'h'
	headcolon rune = 'H'
	untracked rune = 'u'
	modified  rune = 'm'
	staged    rune = 's'
	conflicts rune = 'c'
	ahead     rune = 'a'
	behind    rune = 'b'
	stashed   rune = 'S'
	upstream  rune = 'U'
	// enablers without data
	clean    rune = 'C'
	dirty    rune = 'D'
	outdated rune = 'O'
	latest   rune = 'L'
	local    rune = 'l'
	if_else  rune = 'e'
)

type group struct {
	buf bytes.Buffer

	parent *group
	format formatter

	hasData    bool
	hasValue   bool
	hasEnabler bool
	wasEnabled bool
	leakColor  bool
	leakAttr   bool
	width      int
}

// Print prints the status according to the format.
func Print(s *GitStatus, format string, zsh bool) string {

	in := make(chan rune)
	go func() {
		r := bufio.NewReader(strings.NewReader(format))
		for {
			ch, _, err := r.ReadRune()
			if err != nil {
				close(in)
				return
			}
			in <- ch
		}
	}()

	return buildOutput(s, in, zsh)

}

func buildOutput(s *GitStatus, in chan rune, zsh bool) string {

	root := &group{}
	g := root

	col := false
	att := false
	dat := false
	esc := false
	last := true

	if zsh {
		root.buf.WriteString("%{")
	}

	for ch := range in {
		if esc {
			esc = false
			g.addRune(ch)
			continue
		}

		if col {
			setColor(g, ch)
			col = false
			continue
		}

		if att {
			setAttribute(g, ch)
			att = false
			continue
		}

		if dat {
			setData(g, s, last, ch)
			dat = false
			continue
		}

		switch ch {
		case tEsc:
			esc = true
		case tColor:
			col = true
		case tAttribute:
			att = true
		case tData:
			dat = true
		case tGroupOp:
			g = &group{
				parent: g,
				format: g.format,
			}
			g.format.clearAttributes()
			g.format.clearColor()
		case tGroupCl:
			if g.parent == nil {
				// invalid group close - just print as if escaped
				g.addRune(ch)
				continue
			}
			if g.hasEnabler {
				g.hasData = true
				g.hasValue = g.wasEnabled
			}
			last = g.writeTo(&g.parent.buf)
			if last {
				g.parent.format = g.format
				if !g.leakColor {
					g.parent.format.setColor(0)
				}
				if !g.leakAttr {
					g.parent.format.clearAttributes()
				}
				g.parent.width += g.width
			}
			if g.hasData {
				g.parent.hasData = true
			}
			if g.hasValue {
				g.parent.hasValue = true
			}
			g = g.parent
		default:
			g.addRune(ch)
		}
	}

	// trailing characters
	if col {
		g.addRune(tColor)
	}
	if att {
		g.addRune(tAttribute)
	}
	if dat {
		g.addRune(tData)
	}

	g.format.clearColor()
	g.format.clearAttributes()
	g.format.printANSI(&g.buf)

	if zsh {
		root.buf.WriteString(fmt.Sprintf("%%%dG%%}", root.width))
	}

	return root.buf.String()

}

func setColor(g *group, ch rune) {
	if ch == tReset {
		// Reset color code.
		g.format.clearColor()
		return
	}
	if ch == tLeak {
		// Leak color.
		g.leakColor = true
		return
	}
	code, ok := colors[ch]
	if ok {
		g.format.setColor(code)
		return
	}
	g.addRune(tColor)
	g.addRune(ch)
}

func setAttribute(g *group, ch rune) {
	if ch == tReset {
		// Reset attribute.
		g.format.clearAttributes()
		return
	}
	if ch == tLeak {
		// Leak attribute.
		g.leakAttr = true
		return
	}
	code, ok := attrs[ch]
	if ok {
		g.format.setAttribute(code)
		return
	}
	code, ok = resetAttrs[ch]
	if ok {
		g.format.clearAttribute(code)
		return
	}
	g.addRune(tAttribute)
	g.addRune(ch)
}

func setData(g *group, s *GitStatus, last bool, ch rune) {
	switch ch {
	case head:
		g.hasData = true
		g.hasValue = true
		if s.Branch != "" {
			g.addString(s.Branch)
		} else {
			g.addString(s.Sha[:7])
		}
	case headcolon:
		g.hasData = true
		g.hasValue = true
		if s.Branch != "" {
			g.addString(s.Branch)
		} else {
			g.addString(":")
			g.addString(s.Sha[:7])
		}
	case modified:
		g.addInt(s.Modified)
		g.hasData = true
		if s.Modified > 0 {
			g.hasValue = true
		}
	case untracked:
		g.addInt(s.Untracked)
		g.hasData = true
		if s.Untracked > 0 {
			g.hasValue = true
		}
	case staged:
		g.addInt(s.Staged)
		g.hasData = true
		if s.Staged > 0 {
			g.hasValue = true
		}
	case conflicts:
		g.addInt(s.Conflicts)
		g.hasData = true
		if s.Conflicts > 0 {
			g.hasValue = true
		}
	case ahead:
		g.addInt(s.Ahead)
		g.hasData = true
		if s.Ahead > 0 {
			g.hasValue = true
		}
	case behind:
		g.addInt(s.Behind)
		g.hasData = true
		if s.Behind > 0 {
			g.hasValue = true
		}
	case stashed:
		g.addInt(s.Stashed)
		g.hasData = true
		if s.Stashed > 0 {
			g.hasValue = true
		}
	case upstream:
		g.hasData = true
		if s.Upstream != "" {
			g.hasValue = true
			g.addString(s.Upstream)
		}
	case clean:
		g.hasEnabler = true
		if s.Clean {
			g.wasEnabled = true
		}
	case dirty:
		g.hasEnabler = true
		if !s.Clean {
			g.wasEnabled = true
		}
	case outdated:
		g.hasEnabler = true
		if s.Outdated {
			g.wasEnabled = true
		}
	case latest:
		g.hasEnabler = true
		if !s.Outdated {
			g.wasEnabled = true
		}
	case local:
		g.hasEnabler = true
		if s.Upstream == "" {
			g.wasEnabled = true
		}
	case if_else:
		g.hasEnabler = true
		if !last {
			g.wasEnabled = true
		}
	default:
		g.addRune(tData)
		g.addRune(ch)
	}
}

func (g *group) writeTo(b io.Writer) bool {
	if g.hasData && !g.hasValue {
		return false
	}
	if _, err := g.buf.WriteTo(b); err != nil {
		log.Panic(err)
	}
	return true
}

func (g *group) addRune(r rune) {
	if !unicode.IsSpace(r) {
		g.format.printANSI(&g.buf)
	}
	g.width++
	g.buf.WriteRune(r)
}

func (g *group) addString(s string) {
	g.format.printANSI(&g.buf)
	g.width += len(s)
	g.buf.WriteString(s)
}

func (g *group) addInt(i int) {
	g.addString(strconv.Itoa(i))
}
