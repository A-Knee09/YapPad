/*
NOTE:
Startup splash animation. A "decrypt" style reveal of the YapPad logo
(scrambled glyphs resolving column by column), followed by a shine sweep
across a theme-coloured gradient and a typed-out tagline.
Any key (or mouse click) skips it.
*/
package main

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	splashFPS        = 30
	splashFrameTime  = time.Second / splashFPS
	splashShineWidth = 4.0
	splashShineSpeed = 2.2 // diagonal units per frame
	splashHoldFrames = 18
	splashTagline    = "a terminal note-taking app"
)

var splashLetters = map[rune][]string{
	'Y': {
		"██╗   ██╗",
		"╚██╗ ██╔╝",
		" ╚████╔╝ ",
		"  ╚██╔╝  ",
		"   ██║   ",
		"   ╚═╝   ",
	},
	'A': {
		" █████╗ ",
		"██╔══██╗",
		"███████║",
		"██╔══██║",
		"██║  ██║",
		"╚═╝  ╚═╝",
	},
	'P': {
		"██████╗ ",
		"██╔══██╗",
		"██████╔╝",
		"██╔═══╝ ",
		"██║     ",
		"╚═╝     ",
	},
	'D': {
		"██████╗ ",
		"██╔══██╗",
		"██║  ██║",
		"██║  ██║",
		"██████╔╝",
		"╚═════╝ ",
	},
}

var splashScramble = []rune("░▒▓█▀▄▌▐■□▪◆◇○●╳╱╲#$%&*+=<>?/\\|~^01")

type splashTickMsg struct{}

type splashCell struct {
	ch       rune
	appearAt int // frame the scrambled glyph first shows up
	revealAt int // frame it resolves into the real glyph
	seed     int
}

type splash struct {
	active     bool
	frame      int
	logo       [][]splashCell
	logoWidth  int
	revealDone int // frame by which every cell is resolved
	shineEnd   int
	endFrame   int
	from, to   rgb
	text       rgb
	muted      lipgloss.Color

	// Latest terminal size seen while the splash was up; replayed to the
	// main UI on dismiss so previews/images don't render under the splash.
	width, height int
	sized         bool
}

func newSplash(t Theme) splash {
	rows := make([]string, 6)
	for _, r := range "YAPPAD" {
		for i, line := range splashLetters[r] {
			rows[i] += line
		}
	}

	s := splash{
		active: true,
		from:   colorToRGB(t.Primary),
		to:     colorToRGB(t.Accent),
		text:   rgb{255, 255, 255},
		muted:  t.Muted,
	}

	for y, line := range rows {
		runes := []rune(line)
		s.logoWidth = max(s.logoWidth, len(runes))
		row := make([]splashCell, len(runes))
		for x, ch := range runes {
			if ch == ' ' {
				row[x] = splashCell{ch: ' '}
				continue
			}
			appear := rand.IntN(14) + x/6
			reveal := 10 + x*2/3 + y + rand.IntN(18)
			row[x] = splashCell{ch: ch, appearAt: appear, revealAt: max(reveal, appear+4), seed: rand.IntN(1 << 16)}
			s.revealDone = max(s.revealDone, row[x].revealAt)
		}
		s.logo = append(s.logo, row)
	}

	diag := float64(s.logoWidth + 2*len(rows))
	s.shineEnd = s.revealDone + 4 + int((diag+2*splashShineWidth)/splashShineSpeed)
	s.endFrame = s.shineEnd + splashHoldFrames
	return s
}

func splashTick() tea.Cmd {
	return tea.Tick(splashFrameTime, func(time.Time) tea.Msg { return splashTickMsg{} })
}

// update advances the splash. done reports that the splash just finished.
func (s *splash) update(msg tea.Msg) (cmd tea.Cmd, done bool) {
	switch msg := msg.(type) {
	case splashTickMsg:
		s.frame++
		if s.frame >= s.endFrame {
			s.active = false
			return nil, true
		}
		return splashTick(), false
	case tea.WindowSizeMsg:
		s.width, s.height, s.sized = msg.Width, msg.Height, true
	case tea.KeyMsg:
		s.active = false
		return nil, true
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			s.active = false
			return nil, true
		}
	}
	return nil, false
}

func (s splash) view() string {
	if !s.sized {
		return ""
	}

	var logo string
	if s.width >= s.logoWidth+4 && s.height >= len(s.logo)+6 {
		logo = s.renderLogo()
	} else {
		logo = s.renderSmallLogo()
	}

	// Tagline types out once the logo has resolved.
	typed := 0
	if s.frame > s.revealDone {
		typed = min(len(splashTagline), (s.frame-s.revealDone)*2/3)
	}
	tagline := lipgloss.NewStyle().Foreground(s.muted).Render(splashTagline[:typed])
	if typed < len(splashTagline) && s.frame > s.revealDone && (s.frame/6)%2 == 0 {
		tagline += lipgloss.NewStyle().Foreground(s.gradientAt(1)).Render("▌")
	}
	// Pad to the full tagline width so it types in place instead of re-centering.
	tagline = lipgloss.NewStyle().Width(len(splashTagline) + 1).Render(tagline)
	tagline = lipgloss.PlaceHorizontal(lipgloss.Width(logo), lipgloss.Center, tagline)

	version := ""
	if Version != "dev" && typed == len(splashTagline) {
		version = lipgloss.NewStyle().Foreground(s.muted).Faint(true).Render(Version)
	}
	version = lipgloss.PlaceHorizontal(lipgloss.Width(logo), lipgloss.Center, version)

	body := lipgloss.JoinVertical(lipgloss.Center, logo, "", tagline, version)
	hint := lipgloss.NewStyle().Foreground(s.muted).Faint(true).Render("press any key to skip")

	main := lipgloss.Place(s.width, max(1, s.height-1), lipgloss.Center, lipgloss.Center, body)
	return main + "\n" + lipgloss.PlaceHorizontal(s.width, lipgloss.Center, hint)
}

func (s splash) renderLogo() string {
	var b strings.Builder
	shinePos := float64(s.frame-s.revealDone-4)*splashShineSpeed - splashShineWidth
	shining := s.frame > s.revealDone+4 && s.frame <= s.shineEnd

	for y, row := range s.logo {
		if y > 0 {
			b.WriteByte('\n')
		}
		for x, c := range row {
			switch {
			case c.ch == ' ' || s.frame < c.appearAt:
				b.WriteByte(' ')

			case s.frame < c.revealAt:
				// Scrambling: cycle random glyphs in dim gradient tones.
				g := splashScramble[(c.seed+s.frame/2*7)%len(splashScramble)]
				col := s.gradientRGB(float64(x)/float64(s.logoWidth)).mix(rgb{40, 40, 40}, 0.55)
				b.WriteString(lipgloss.NewStyle().Foreground(col.color()).Render(string(g)))

			default:
				col := s.gradientRGB(float64(x) / float64(s.logoWidth))
				since := s.frame - c.revealAt
				if since < 3 {
					// Brief white flash as a glyph locks in.
					col = col.mix(s.text, 0.8-float64(since)*0.25)
				}
				if shining {
					d := float64(x+2*y) - shinePos
					if d < 0 {
						d = -d
					}
					if d < splashShineWidth {
						col = col.mix(s.text, 0.85*(1-d/splashShineWidth))
					}
				}
				b.WriteString(lipgloss.NewStyle().Foreground(col.color()).Bold(true).Render(string(c.ch)))
			}
		}
	}
	return b.String()
}

// renderSmallLogo is used when the terminal is too small for the block logo.
func (s splash) renderSmallLogo() string {
	word := []rune("YapPad")
	var b strings.Builder
	for i, ch := range word {
		at := 4 + i*4
		if s.frame < at {
			g := splashScramble[(i*13+s.frame/2*7)%len(splashScramble)]
			b.WriteString(lipgloss.NewStyle().Foreground(s.muted).Render(string(g)))
			continue
		}
		col := s.gradientRGB(float64(i) / float64(len(word)-1))
		b.WriteString(lipgloss.NewStyle().Foreground(col.color()).Bold(true).Render(string(ch)))
	}
	return b.String()
}

func (s splash) gradientRGB(t float64) rgb { return s.from.mix(s.to, t) }
func (s splash) gradientAt(t float64) lipgloss.Color {
	return s.gradientRGB(t).color()
}

// Colour helpers

type rgb struct{ r, g, b float64 }

func (c rgb) mix(o rgb, t float64) rgb {
	t = min(1, max(0, t))
	return rgb{c.r + (o.r-c.r)*t, c.g + (o.g-c.g)*t, c.b + (o.b-c.b)*t}
}

func (c rgb) color() lipgloss.Color {
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", uint8(c.r+0.5), uint8(c.g+0.5), uint8(c.b+0.5)))
}

var ansi16 = [16]rgb{
	{0, 0, 0}, {205, 0, 0}, {0, 205, 0}, {205, 205, 0},
	{0, 0, 238}, {205, 0, 205}, {0, 205, 205}, {229, 229, 229},
	{127, 127, 127}, {255, 0, 0}, {0, 255, 0}, {255, 255, 0},
	{92, 92, 255}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255},
}

// colorToRGB converts a lipgloss colour (hex or ANSI 256 index) to RGB.
func colorToRGB(c lipgloss.Color) rgb {
	s := string(c)
	if strings.HasPrefix(s, "#") && len(s) == 7 {
		v, err := strconv.ParseUint(s[1:], 16, 32)
		if err == nil {
			return rgb{float64(v >> 16 & 0xff), float64(v >> 8 & 0xff), float64(v & 0xff)}
		}
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 || n > 255 {
		return rgb{200, 200, 200}
	}
	switch {
	case n < 16:
		return ansi16[n]
	case n < 232:
		n -= 16
		level := func(v int) float64 {
			if v == 0 {
				return 0
			}
			return float64(55 + v*40)
		}
		return rgb{level(n / 36), level(n / 6 % 6), level(n % 6)}
	default:
		g := float64(8 + (n-232)*10)
		return rgb{g, g, g}
	}
}
