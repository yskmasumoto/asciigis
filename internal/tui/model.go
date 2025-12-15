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

type model struct {
	geoPath   string
	geometry  geo.TuiGeometry
	width     int
	height    int
	mapWidth  int
	mapHeight int
	ready     bool
	loading   bool
	err       error
}

// NewModel creates a Bubble Tea model configured with a GeoJSON path.
func NewModel(geoPath string) model {
	return model{geoPath: geoPath}
}

// Run launches the TUI.
func Run(geoPath string) error {
	_, err := tea.NewProgram(NewModel(geoPath), tea.WithAltScreen()).Run()
	return err
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.mapWidth = maxInt(msg.Width-8, minMapWidth)
		m.mapHeight = maxInt(msg.Height-8, minMapHeight)
		m.ready = true
		m.loading = true
		return m, loadGeometryCmd(m.geoPath, m.mapWidth, m.mapHeight)

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
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			if m.ready {
				m.loading = true
				return m, loadGeometryCmd(m.geoPath, m.mapWidth, m.mapHeight)
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return "Calculating viewport..."
	}

	if m.err != nil {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render("asciigis viewer"),
			infoStyle.Render(fmt.Sprintf("Failed to load: %v", m.err)),
			helpStyle.Render("q: quit"),
		)
	}

	mapBlock := mapStyle.Width(m.geometry.Width + 2).Render(renderCanvas(m.geometry))

	info := infoStyle.Render(strings.Join([]string{
		fmt.Sprintf("File: %s", m.geoPath),
		fmt.Sprintf("Bounds: lon %.4f .. %.4f | lat %.4f .. %.4f", m.geometry.Bounds.LonMin, m.geometry.Bounds.LonMax, m.geometry.Bounds.LatMin, m.geometry.Bounds.LatMax),
		fmt.Sprintf("Canvas: %dx%d", m.geometry.Width, m.geometry.Height),
		fmt.Sprintf("Polygons: %d", len(m.geometry.Polygons)),
	}, "\n"))

	statusText := "Loaded"
	if m.loading {
		statusText = "Loading..."
	}

	footer := helpStyle.Render(fmt.Sprintf("q: quit | r: reload | %s", statusText))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("asciigis viewer"),
		mapBlock,
		info,
		footer,
	)
}

func renderCanvas(geometry geo.TuiGeometry) string {
	if geometry.Width == 0 || geometry.Height == 0 {
		return "No geometry yet"
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
