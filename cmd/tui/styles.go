package tui

import "github.com/charmbracelet/lipgloss"

// ── Colors from TUI_DESIGN.md ──
const (
	colorOrange  = lipgloss.Color("#F97316")
	colorEmerald = lipgloss.Color("#10B981")
	colorAmber   = lipgloss.Color("#F59E0B")
	colorRed     = lipgloss.Color("#EF4444")
	colorMuted   = lipgloss.Color("#6B7280")
	colorText    = lipgloss.Color("#9CA3AF")
	colorBright  = lipgloss.Color("#F9FAFB")
	colorSurface = lipgloss.Color("#111827")
	colorBorder  = lipgloss.Color("#374151")
)

// ── Panel borders ──
var panelStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorBorder).
	Padding(0, 1)

var activePanelStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorOrange).
	Padding(0, 1)

// ── Header bar ──
func headerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(colorOrange).
		Foreground(colorSurface).
		Bold(true).
		Padding(0, 2).
		Width(width)
}

// ── Menu items ──
var selectedItemStyle = lipgloss.NewStyle().
	BorderLeft(true).
	BorderStyle(lipgloss.ThickBorder()).
	BorderForeground(colorOrange).
	PaddingLeft(1).
	Bold(true).
	Foreground(colorBright)

var normalItemStyle = lipgloss.NewStyle().
	PaddingLeft(3).
	Foreground(colorText)

var menuDescStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

// ── Badges ──
var successBadge = lipgloss.NewStyle().
	Background(colorEmerald).
	Foreground(colorBright).
	Bold(true).
	Padding(0, 1)

var errorBadge = lipgloss.NewStyle().
	Background(colorRed).
	Foreground(colorBright).
	Bold(true).
	Padding(0, 1)

// ── Key hints ──
var keyStyle = lipgloss.NewStyle().
	Foreground(colorOrange).
	Bold(true)

var descStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

// ── Logo ──
var logoStyle = lipgloss.NewStyle().
	Foreground(colorOrange).
	Bold(true)

var taglineStyle = lipgloss.NewStyle().
	Foreground(colorMuted).
	Italic(true)

// ── Result preview ──
var statLabelStyle = lipgloss.NewStyle().
	Foreground(colorMuted)

var statValueStyle = lipgloss.NewStyle().
	Foreground(colorEmerald).
	Bold(true)

// ── History ──
var historySelectedStyle = lipgloss.NewStyle().
	BorderLeft(true).
	BorderStyle(lipgloss.ThickBorder()).
	BorderForeground(colorOrange).
	PaddingLeft(1).
	Foreground(colorBright)

var historyNormalStyle = lipgloss.NewStyle().
	PaddingLeft(3).
	Foreground(colorText)

var historyMetaStyle = lipgloss.NewStyle().
	PaddingLeft(3).
	Foreground(colorMuted)

// ── Helpers ──
func keyHint(key, desc string) string {
	return keyStyle.Render("["+key+"]") + " " + descStyle.Render(desc)
}
