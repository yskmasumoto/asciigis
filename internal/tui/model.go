package tui

import (
	"fmt"
	"strings"

	"asciigis/internal/geo"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")).
			Background(lipgloss.Color("#3A86FF")).
			Padding(0, 1).
			Bold(true)

	mapStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#94E458")).
			Padding(1, 1)

	infoStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#5BE3FF")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			MarginTop(1)
)

const (
	minMapWidth  = 20
	minMapHeight = 10
)

type geometryLoadedMsg struct {
	geometry geo.TuiGeometry
	err      error
}

// Options configures the TUI behavior.
// MapWidth/MapHeight, when > 0, request a fixed canvas size.
// The final size may be clamped to the current terminal size.
type Options struct {
	MapWidth  int
	MapHeight int
}

type model struct {
	geoPath        string
	inputPath      string
	editing        bool
	geometry       geo.TuiGeometry
	width          int
	height         int
	mapWidth       int
	mapHeight      int
	maxMapWidth    int
	maxMapHeight   int
	fixedMapWidth  int
	fixedMapHeight int
	ready          bool
	loading        bool
	err            error
}

// NewModel creates a Bubble Tea model configured with a GeoJSON path.
func NewModel(geoPath string, opts Options) model {
	m := model{geoPath: geoPath, fixedMapWidth: opts.MapWidth, fixedMapHeight: opts.MapHeight}
	if strings.TrimSpace(geoPath) == "" {
		m.editing = true
		m.inputPath = ""
	} else {
		m.inputPath = geoPath
	}
	return m
}

// Run launches the TUI.
func Run(geoPath string) error {
	return RunWithOptions(geoPath, Options{})
}

// RunWithOptions launches the TUI with additional configuration.
func RunWithOptions(geoPath string, opts Options) error {
	_, err := tea.NewProgram(NewModel(geoPath, opts), tea.WithAltScreen()).Run()
	return err
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		prevW, prevH := m.mapWidth, m.mapHeight
		m.width = msg.Width
		m.height = msg.Height
		computedMapWidth := maxInt(msg.Width-8, minMapWidth)
		computedMapHeight := maxInt(msg.Height-8, minMapHeight)
		m.maxMapWidth = computedMapWidth
		m.maxMapHeight = computedMapHeight

		desiredW := computedMapWidth
		desiredH := computedMapHeight

		if m.fixedMapWidth > 0 {
			desiredW = maxInt(minInt(m.fixedMapWidth, computedMapWidth), minMapWidth)
		}
		if m.fixedMapHeight > 0 {
			desiredH = maxInt(minInt(m.fixedMapHeight, computedMapHeight), minMapHeight)
		}
		m.mapWidth = desiredW
		m.mapHeight = desiredH
		m.ready = true
		if strings.TrimSpace(m.geoPath) != "" && (prevW != m.mapWidth || prevH != m.mapHeight || m.geometry.Width == 0 || m.geometry.Height == 0) {
			m.loading = true
			return m, loadGeometryCmd(m.geoPath, m.mapWidth, m.mapHeight)
		}
		m.loading = false
		return m, nil

	case geometryLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.geometry = geo.TuiGeometry{}
			return m, nil
		}
		m.geometry = msg.geometry
		m.err = nil
		return m, nil

	case tea.KeyMsg:
		// Path input mode.
		if m.editing {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				// If we already have a loaded path, allow canceling back to view mode.
				if strings.TrimSpace(m.geoPath) != "" {
					m.editing = false
					m.inputPath = m.geoPath
					return m, nil
				}
				// Otherwise stay in editing mode.
				return m, nil
			case "enter":
				p := strings.TrimSpace(m.inputPath)
				if p == "" {
					m.err = fmt.Errorf("path is empty")
					return m, nil
				}
				m.geoPath = p
				m.inputPath = p
				m.editing = false
				m.loading = true
				m.err = nil
				m.geometry = geo.TuiGeometry{}
				if m.ready {
					return m, loadGeometryCmd(m.geoPath, m.mapWidth, m.mapHeight)
				}
				return m, nil
			case "backspace", "ctrl+h":
				m.inputPath = dropLastRune(m.inputPath)
				return m, nil
			case "ctrl+u":
				m.inputPath = ""
				return m, nil
			}

			if msg.Type == tea.KeyRunes {
				m.inputPath += string(msg.Runes)
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			if m.ready && strings.TrimSpace(m.geoPath) != "" {
				m.loading = true
				return m, loadGeometryCmd(m.geoPath, m.mapWidth, m.mapHeight)
			}
		case "a":
			return m.resizeCanvas(-1, 0)
		case "d":
			return m.resizeCanvas(1, 0)
		case "w":
			return m.resizeCanvas(0, 1)
		case "s":
			return m.resizeCanvas(0, -1)
		case "/", "p":
			m.editing = true
			if strings.TrimSpace(m.geoPath) != "" {
				m.inputPath = m.geoPath
			}
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return "Calculating viewport..."
	}

	mapBlock := mapStyle.Width(m.mapWidth).Render(renderCanvas(m.geometry, m.loading, m.err))

	pathPanel := ""
	if m.editing {
		input := m.inputPath
		// simple caret
		input = input + "_"
		pathPanel = infoStyle.Render(strings.Join([]string{
			"Enter GeoJSON path:",
			input,
			"Enter: load | Esc: cancel | Ctrl+U: clear",
		}, "\n"))
	}

	infoLines := []string{fmt.Sprintf("File: %s", emptyWhen(strings.TrimSpace(m.geoPath), "(none)"))}
	if m.geometry.Width > 0 && m.geometry.Height > 0 && m.err == nil {
		infoLines = append(infoLines,
			fmt.Sprintf("Bounds: lon %.4f .. %.4f | lat %.4f .. %.4f", m.geometry.Bounds.LonMin, m.geometry.Bounds.LonMax, m.geometry.Bounds.LatMin, m.geometry.Bounds.LatMax),
			fmt.Sprintf("Canvas: %dx%d", m.geometry.Width, m.geometry.Height),
			fmt.Sprintf("Polygons: %d", len(m.geometry.Polygons)),
		)
	}
	if m.err != nil {
		infoLines = append(infoLines, fmt.Sprintf("Error: %v", m.err))
	}
	info := infoStyle.Render(strings.Join(infoLines, "\n"))

	statusText := "Loaded"
	if m.loading {
		statusText = "Loading..."
	}

	footerText := fmt.Sprintf("q: quit | r: reload | a/d: width -/+ | w/s: height +/- | / or p: set path | %s", statusText)
	if m.editing {
		footerText = "q: quit | typing..."
	}
	footer := helpStyle.Render(footerText)

	parts := []string{
		titleStyle.Render("asciigis viewer"),
		mapBlock,
	}
	if m.editing {
		parts = append(parts, pathPanel)
	}
	parts = append(parts, info, footer)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		parts...,
	)
}

func renderCanvas(geometry geo.TuiGeometry, loading bool, loadErr error) string {
	if loading {
		return "Loading..."
	}
	if loadErr != nil {
		return fmt.Sprintf("Failed to load: %v", loadErr)
	}
	if geometry.Width == 0 || geometry.Height == 0 {
		return "No geometry yet (press '/' to set path)"
	}

	canvas := make([][]rune, geometry.Height)
	for y := 0; y < geometry.Height; y++ {
		row := make([]rune, geometry.Width)
		for x := 0; x < geometry.Width; x++ {
			row[x] = ' '
		}
		canvas[y] = row
	}

	for _, polygon := range geometry.Polygons {
		for _, ring := range polygon.Rings {
			for _, coord := range ring {
				x, y := coord[0], coord[1]
				if x < 0 || y < 0 || x >= geometry.Width || y >= geometry.Height {
					continue
				}
				canvas[y][x] = '*'
			}
		}
	}

	lines := make([]string, len(canvas))
	for i, row := range canvas {
		lines[i] = string(row)
	}
	return strings.Join(lines, "\n")
}

func loadGeometryCmd(path string, width, height int) tea.Cmd {
	return func() tea.Msg {
		geometry, err := geo.ConvertTui(path, width, height)
		return geometryLoadedMsg{geometry: geometry, err: err}
	}
}

func dropLastRune(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if len(r) == 0 {
		return ""
	}
	return string(r[:len(r)-1])
}

func emptyWhen(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (m model) resizeCanvas(deltaW, deltaH int) (tea.Model, tea.Cmd) {
	if !m.ready {
		return m, nil
	}
	if strings.TrimSpace(m.geoPath) == "" {
		return m, nil
	}

	maxW := m.maxMapWidth
	maxH := m.maxMapHeight
	if maxW <= 0 {
		maxW = maxInt(m.mapWidth, minMapWidth)
	}
	if maxH <= 0 {
		maxH = maxInt(m.mapHeight, minMapHeight)
	}

	newW := clampInt(m.mapWidth+deltaW, minMapWidth, maxW)
	newH := clampInt(m.mapHeight+deltaH, minMapHeight, maxH)
	if newW == m.mapWidth && newH == m.mapHeight {
		return m, nil
	}

	m.mapWidth = newW
	m.mapHeight = newH
	// Persist user choice across WindowSizeMsg.
	m.fixedMapWidth = newW
	m.fixedMapHeight = newH

	m.loading = true
	m.err = nil
	m.geometry = geo.TuiGeometry{}
	return m, loadGeometryCmd(m.geoPath, m.mapWidth, m.mapHeight)
}
