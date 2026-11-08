/*
NOTE:
Startup splash animation. One of two logo effects plays, followed by a
shine sweep across a theme-coloured gradient and a typed-out tagline:
  - decrypt: scrambled glyphs resolve column by column
  - shuffle: each logo row starts as a random permutation of its glyphs and
    is sorted back into place with a visible odd-even transposition sort

Any key (or mouse click) skips it.
*/
package main

import (
	"fmt"
	"math/rand/v2"
	"slices"
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

const (
	splashEffectDecrypt = "decrypt"
	splashEffectShuffle = "shuffle"
)

var splashEffects = []string{splashEffectDecrypt, splashEffectShuffle}

// splashNotYet marks a phase whose start frame isn't known yet.
const splashNotYet = 1 << 30

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
	effect     string
	frame      int
	logo       [][]splashCell
	logoWidth  int
	revealDone int // frame by which every cell is resolved
	shineEnd   int
	endFrame   int
	from, to   rgb
	text       rgb
	muted      lipgloss.Color

	// Shuffle effect state, per logo row: perm[y][x] is the target column of
	// the glyph currently shown at column x.
	perm       [][]int
	swapped    [][]bool
	lockedAt   [][]int // frame the glyph at x settled into place, -1 if not
	sortStart  int
	sortParity int
	sortCalm   int // consecutive rounds with no swaps

	// Latest terminal size seen while the splash was up; replayed to the
	// main UI on dismiss so previews/images don't render under the splash.
	width, height int
	sized         bool
}

func newSplash(t Theme, effect string) splash {
	if !slices.Contains(splashEffects, effect) {
		effect = splashEffects[rand.IntN(len(splashEffects))]
	}

	rows := make([]string, 6)
	for _, r := range "YAPPAD" {
		for i, line := range splashLetters[r] {
			rows[i] += line
		}
	}

	s := splash{
		active: true,
		effect: effect,
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

	if effect == splashEffectShuffle {
		s.initShuffle()
	} else {
		s.setRevealDone(s.revealDone)
	}
	return s
}

// setRevealDone schedules the shine, tagline and exit after the logo is complete.
func (s *splash) setRevealDone(frame int) {
	s.revealDone = frame
	diag := float64(s.logoWidth + 2*len(s.logo))
	s.shineEnd = s.revealDone + 4 + int((diag+2*splashShineWidth)/splashShineSpeed)
	s.endFrame = s.shineEnd + splashHoldFrames
}

func (s *splash) initShuffle() {
	s.sortStart = 16
	s.revealDone, s.shineEnd, s.endFrame = splashNotYet, splashNotYet, splashNotYet
	for y, row := range s.logo {
		p := rand.Perm(len(row))
		locked := make([]int, len(row))
		for x := range row {
			// Cells fade in quickly before sorting begins.
			s.logo[y][x].appearAt = rand.IntN(s.sortStart - 4)
			locked[x] = -1
			if p[x] == x {
				locked[x] = 0
			}
		}
		s.perm = append(s.perm, p)
		s.swapped = append(s.swapped, make([]bool, len(row)))
		s.lockedAt = append(s.lockedAt, locked)
	}
}

// sortStep runs one odd-even transposition round on every row.
func (s *splash) sortStep() {
	changed := false
	for y, p := range s.perm {
		clear(s.swapped[y])
		for i := s.sortParity; i+1 < len(p); i += 2 {
			if p[i] > p[i+1] {
				p[i], p[i+1] = p[i+1], p[i]
				s.swapped[y][i], s.swapped[y][i+1] = true, true
				changed = true
			}
		}
		for x, t := range p {
			if t != x {
				s.lockedAt[y][x] = -1
			} else if s.lockedAt[y][x] < 0 {
				s.lockedAt[y][x] = s.frame
			}
		}
	}
	s.sortParity ^= 1
	if changed {
		s.sortCalm = 0
		return
	}
	// One quiet odd round plus one quiet even round means every row is sorted.
	if s.sortCalm++; s.sortCalm >= 2 {
		s.perm = nil
		s.setRevealDone(s.frame)
	}
}

func splashTick() tea.Cmd {
	return tea.Tick(splashFrameTime, func(time.Time) tea.Msg { return splashTickMsg{} })
}

// update advances the splash. done reports that the splash just finished.
func (s *splash) update(msg tea.Msg) (cmd tea.Cmd, done bool) {
	switch msg := msg.(type) {
	case splashTickMsg:
		s.frame++
		if s.perm != nil && s.frame >= s.sortStart {
			s.sortStep()
		}
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
	if s.perm != nil {
		return s.renderShuffle()
	}
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
				b.WriteString(s.settledGlyph(c.ch, x, y, c.revealAt, shining, shinePos))
			}
		}
	}
	return b.String()
}

// settledGlyph renders a glyph that is in its final place: gradient colour,
// a brief white flash right after it locks in, and the shine sweep.
func (s splash) settledGlyph(ch rune, x, y, settledAt int, shining bool, shinePos float64) string {
	col := s.gradientRGB(float64(x) / float64(s.logoWidth))
	if since := s.frame - settledAt; since >= 0 && since < 3 {
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
	return lipgloss.NewStyle().Foreground(col.color()).Bold(true).Render(string(ch))
}

// renderShuffle draws the logo mid-sort. Each glyph carries the gradient
// colour of its destination, so the colours visibly sort too.
func (s splash) renderShuffle() string {
	var b strings.Builder
	for y, row := range s.logo {
		if y > 0 {
			b.WriteByte('\n')
		}
		for x, t := range s.perm[y] {
			ch := row[t].ch
			if ch == ' ' || s.frame < row[x].appearAt {
				b.WriteByte(' ')
				continue
			}
			if s.lockedAt[y][x] >= 0 {
				b.WriteString(s.settledGlyph(ch, x, y, s.lockedAt[y][x], false, 0))
				continue
			}
			col := s.gradientRGB(float64(t)/float64(s.logoWidth)).mix(rgb{40, 40, 40}, 0.45)
			if s.swapped[y][x] {
				col = col.mix(s.text, 0.5)
			}
			b.WriteString(lipgloss.NewStyle().Foreground(col.color()).Render(string(ch)))
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
