package tui

import "github.com/charmbracelet/lipgloss"

// Color palette — a warm terminal aesthetic
var (
	colorAccent   = lipgloss.Color("#7D56F4")   // purple
	colorUser     = lipgloss.Color("#5FB3D9")   // blue
	colorAssistant = lipgloss.Color("#7D56F4")  // purple
	colorTool     = lipgloss.Color("#E5C07B")   // gold
	colorResult   = lipgloss.Color("#98C379")   // green
	colorSystem   = lipgloss.Color("#5C6370")   // gray
	colorError    = lipgloss.Color("#E06C75")   // red
	colorDim      = lipgloss.Color("#3E4451")   // dark gray
	colorSuccess  = lipgloss.Color("#98C379")   // green
)

var (
	// headerStyle is the top bar showing project + session info
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(colorAccent).
			Padding(0, 1)

	// dimStyle for less important text
	dimStyle = lipgloss.NewStyle().
			Foreground(colorSystem)

	// labelStyle for message role labels
	labelStyle = lipgloss.NewStyle().Bold(true)

	userLabelStyle = labelStyle.Foreground(colorUser)
	assistantLabelStyle = labelStyle.Foreground(colorAssistant)
	toolLabelStyle = labelStyle.Foreground(colorTool)
	resultLabelStyle = labelStyle.Foreground(colorResult)
	systemLabelStyle = labelStyle.Foreground(colorSystem)
	errorLabelStyle = labelStyle.Foreground(colorError)

	// contentStyle for message bodies
	userContentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#D7DAE0"))
	assistantContentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#D7DAE0"))
	toolContentStyle = lipgloss.NewStyle().Foreground(colorTool).Faint(true)
	resultContentStyle = lipgloss.NewStyle().Foreground(colorResult)
	systemContentStyle = lipgloss.NewStyle().Foreground(colorSystem).Faint(true)
	errorContentStyle = lipgloss.NewStyle().Foreground(colorError)

	// composerStyle for the input prompt
	composerStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	// hintStyle for the bottom status bar
	hintStyle = lipgloss.NewStyle().
			Foreground(colorSystem).
			Faint(true)

	// successStyle for success messages
	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	// boxStyle for bordered panels (help, sessions)
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2)
)
