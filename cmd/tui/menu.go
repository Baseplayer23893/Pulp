package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Baseplayer23893/Pulp/internal/config"
)

// ── State machine ──

type state int

const (
	stateHome state = iota
	stateExtractInput
	stateSqueezing
	stateResult
	stateHistory
	stateSettings
)

// ── Menu items ──

type menuItem struct {
	icon string
	name string
	desc string
	cmd  string
}

var menuItems = []menuItem{
	{"🌐", "Extract Web Page", "Convert any URL to clean markdown", "extract"},
	{"▶ ", "YouTube Transcript", "Pull video/shorts transcript", "youtube"},
	{"📸", "Instagram Reel", "Extract audio + caption", "instagram"},
	{"🟠", "Reddit Post", "Post + top comments as markdown", "reddit"},
	{"📄", "PDF Extraction", "Extract text from PDF file", "pdf"},
	{"📦", "Package Skill", "Create skill.zip from content", ""},
	{"📂", "History", "View past squeezes", ""},
	{"⚙ ", "Settings", "Configure defaults", ""},
}

// ── Messages ──

type squeezeStartMsg struct{}
type squeezeDoneMsg struct {
	output string
	err    error
	dur    time.Duration
}
type squeezeProgressMsg float64

// ── Model ──

type Model struct {
	state      state
	cursor     int
	quickInput textinput.Model
	urlInput   textinput.Model
	spinner    spinner.Model
	progress   progress.Model
	viewport   viewport.Model
	width      int
	height     int

	// Extraction state
	selectedSource int
	squeezeOutput  string
	squeezeErr     error
	squeezeDur     time.Duration
	squeezeURL     string

	// History
	history *config.History

	// Settings
	settingsCursor  int
	settingsValues  [2]string // output_dir, format
	settingsEditing bool
	settingsInput   textinput.Model
}

func initialModel() Model {
	// Quick squeeze input (home screen bottom bar)
	qi := textinput.New()
	qi.Placeholder = "paste URL and hit Enter..."
	qi.Prompt = "›  "
	qi.CharLimit = 500
	qi.Width = 60
	qi.PromptStyle = lipgloss.NewStyle().Foreground(colorOrange)
	qi.PlaceholderStyle = lipgloss.NewStyle().Foreground(colorMuted)
	qi.TextStyle = lipgloss.NewStyle().Foreground(colorBright)

	// URL input (extract screen)
	ui := textinput.New()
	ui.Placeholder = "https://..."
	ui.Prompt = "🔗  "
	ui.CharLimit = 500
	ui.Width = 60
	ui.PlaceholderStyle = lipgloss.NewStyle().Foreground(colorMuted)
	ui.TextStyle = lipgloss.NewStyle().Foreground(colorBright)

	// Spinner
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colorOrange)

	// Progress bar
	prog := progress.New(
		progress.WithScaledGradient("#F97316", "#10B981"),
		progress.WithoutPercentage(),
	)

	// History
	hist := config.LoadHistory()

	// Settings
	cfg := config.Load()

	si := textinput.New()
	si.Prompt = "  "
	si.CharLimit = 256
	si.Width = 40
	si.PromptStyle = lipgloss.NewStyle().Foreground(colorOrange)
	si.PlaceholderStyle = lipgloss.NewStyle().Foreground(colorMuted)
	si.TextStyle = lipgloss.NewStyle().Foreground(colorBright)

	return Model{
		state:          stateHome,
		quickInput:     qi,
		urlInput:       ui,
		spinner:        sp,
		progress:       prog,
		history:        hist,
		settingsValues: [2]string{cfg.OutputDir, cfg.DefaultFormat},
		settingsInput:  si,
	}
}

// ── Entry points ──

func ShowMenu() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func ShowDashboard() error {
	m := initialModel()
	m.state = stateHistory
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// ── Bubble Tea interface ──

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 12
		if m.progress.Width < 20 {
			m.progress.Width = 20
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		pm, cmd := m.progress.Update(msg)
		m.progress = pm.(progress.Model)
		return m, cmd

	case squeezeDoneMsg:
		m.squeezeOutput = msg.output
		m.squeezeErr = msg.err
		m.squeezeDur = msg.dur
		m.state = stateResult

		// Save to history
		if msg.err == nil {
			words := len(strings.Fields(msg.output))
			title := ""
			for _, l := range strings.Split(msg.output, "\n") {
				if strings.HasPrefix(l, "title:") {
					title = strings.TrimSpace(strings.TrimPrefix(l, "title:"))
					break
				}
			}
			m.history.Add(config.HistoryEntry{
				URL:       m.squeezeURL,
				Source:    menuItems[m.selectedSource].cmd,
				Title:     title,
				WordCount: words,
			})
		}

		// Set up viewport for result
		content := m.squeezeOutput
		if m.squeezeErr != nil {
			content = m.squeezeErr.Error()
		}
		m.viewport = viewport.New(m.width-10, m.height-14)
		m.viewport.SetContent(content)
		return m, nil
	}

	// Pass through to active inputs
	if m.state == stateHome {
		var cmd tea.Cmd
		m.quickInput, cmd = m.quickInput.Update(msg)
		return m, cmd
	}
	if m.state == stateExtractInput {
		var cmd tea.Cmd
		m.urlInput, cmd = m.urlInput.Update(msg)
		return m, cmd
	}
	if m.state == stateResult {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	if m.state == stateSettings && m.settingsEditing {
		var cmd tea.Cmd
		m.settingsInput, cmd = m.settingsInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global exits
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.state {

	// ════ HOME ════
	case stateHome:
		if m.quickInput.Focused() {
			switch key {
			case "esc":
				m.quickInput.Blur()
				return m, nil
			case "enter":
				url := strings.TrimSpace(m.quickInput.Value())
				if url != "" {
					// Auto-detect source and squeeze
					m.squeezeURL = url
					m.selectedSource = detectSource(url)
					m.state = stateSqueezing
					m.quickInput.SetValue("")
					m.quickInput.Blur()
					return m, tea.Batch(m.spinner.Tick, m.startSqueeze())
				}
				return m, nil
			}
			var cmd tea.Cmd
			m.quickInput, cmd = m.quickInput.Update(msg)
			return m, cmd
		}

		switch key {
		case "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(menuItems)-1 {
				m.cursor++
			}
		case "/":
			m.quickInput.Focus()
			return m, textinput.Blink
		case "h":
			m.history = config.LoadHistory()
			m.state = stateHistory
			m.cursor = 0 // reset so history always opens at top
		case "s":
			m.state = stateSettings
			m.settingsCursor = 0
			m.settingsEditing = false
		case "enter":
			return m.selectMenuItem()
		}

	// ════ EXTRACT INPUT ════
	case stateExtractInput:
		switch key {
		case "esc":
			m.state = stateHome
			m.urlInput.Blur()
			return m, nil
		case "enter":
			url := strings.TrimSpace(m.urlInput.Value())
			if url != "" {
				m.squeezeURL = url
				m.state = stateSqueezing
				return m, tea.Batch(m.spinner.Tick, m.startSqueeze())
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.urlInput, cmd = m.urlInput.Update(msg)
		return m, cmd

	// ════ SQUEEZING ════
	case stateSqueezing:
		if key == "esc" || key == "ctrl+c" {
			m.state = stateHome
			return m, nil
		}

	// ════ RESULT ════
	case stateResult:
		switch key {
		case "esc", "q":
			m.state = stateHome
		case "r":
			// Re-squeeze
			m.state = stateSqueezing
			return m, tea.Batch(m.spinner.Tick, m.startSqueeze())
		case "s":
			// Save output to a file in the configured output dir
			if m.squeezeOutput != "" {
				cfg := config.Load()
				outDir := cfg.OutputDir
				if outDir == "" || outDir == "." {
					outDir, _ = os.Getwd()
				}
				// Derive a filename from the URL slug
				slug := filepath.Base(strings.TrimRight(m.squeezeURL, "/"))
				if slug == "." || slug == "" {
					slug = "output"
				}
				out := filepath.Join(outDir, slug+".md")
				os.MkdirAll(outDir, 0755)
				os.WriteFile(out, []byte(m.squeezeOutput), 0644)
			}
		case "c":
			// Copy output to system clipboard
			if m.squeezeOutput != "" {
				clipboard.WriteAll(m.squeezeOutput)
			}
		}
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	// ════ HISTORY ════
	case stateHistory:
		// maxShow mirrors the view's calculation so cursor never goes invisible
		maxVisible := func() int {
			entries := m.history.Recent(30)
			n := m.height - 10
			if n > len(entries) {
				n = len(entries)
			}
			if n < 1 {
				n = 1
			}
			return n
		}
		switch key {
		case "esc", "q":
			m.state = stateHome
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < maxVisible()-1 {
				m.cursor++
			}
		case "d":
			if mv := maxVisible(); mv > 0 && m.cursor < mv {
				// Delete persists immediately to history.json
				m.history.Delete(m.cursor)
				// Clamp cursor so it stays in bounds after removal
				newMax := maxVisible()
				if m.cursor > 0 && m.cursor >= newMax {
					m.cursor--
				}
			}
		case "r":
			// Re-squeeze the selected history entry
			entries := m.history.Recent(30)
			if len(entries) > 0 && m.cursor < len(entries) {
				e := entries[m.cursor]
				m.squeezeURL = e.URL
				// Map source name back to menu index
				for i, item := range menuItems {
					if item.cmd == e.Source {
						m.selectedSource = i
						break
					}
				}
				m.state = stateSqueezing
				return m, tea.Batch(m.spinner.Tick, m.startSqueeze())
			}
		}

	// ════ SETTINGS ════
	case stateSettings:
		if m.settingsEditing {
			// Inside an active text-input field — only intercept Enter/Esc;
			// everything else must be forwarded directly to the input model.
			switch key {
			case "enter":
				v := strings.TrimSpace(m.settingsInput.Value())
				if v != "" {
					m.settingsValues[m.settingsCursor] = v
				}
				m.settingsEditing = false
				m.settingsInput.Blur()
				return m, nil
			case "esc":
				m.settingsEditing = false
				m.settingsInput.Blur()
				return m, nil
			default:
				// Forward the raw key event to the text input so the user can type.
				var cmd tea.Cmd
				m.settingsInput, cmd = m.settingsInput.Update(msg)
				return m, cmd
			}
		}

		// Browse mode (not editing)
		switch key {
		case "esc":
			m.state = stateHome
		case "s":
			// Save current in-memory values and exit
			cfg := config.Load()
			cfg.OutputDir = m.settingsValues[0]
			cfg.DefaultFormat = m.settingsValues[1]
			cfg.Save() // writes to ~/.config/pulp/config.json
			m.state = stateHome
		case "enter":
			// Open the current field for editing
			m.settingsInput.SetValue(m.settingsValues[m.settingsCursor])
			m.settingsInput.CursorEnd()
			m.settingsEditing = true
			m.settingsInput.Focus()
			return m, textinput.Blink
		case "left", "h", "right", "l":
			// Cycle DefaultFormat (field index 1) with arrow keys
			if m.settingsCursor == 1 {
				formats := []string{"md", "skillzip", "single"}
				cur := 0
				for i, f := range formats {
					if f == m.settingsValues[1] {
						cur = i
						break
					}
				}
				if key == "left" || key == "h" {
					cur = (cur - 1 + len(formats)) % len(formats)
				} else {
					cur = (cur + 1) % len(formats)
				}
				m.settingsValues[1] = formats[cur]
			}
		case "up", "k":
			if m.settingsCursor > 0 {
				m.settingsCursor--
			}
		case "down", "j", "tab":
			if m.settingsCursor < 1 {
				m.settingsCursor++
			}
		}
	}

	return m, nil
}

func (m Model) selectMenuItem() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case 5: // Package — just go back for now
		return m, nil
	case 6: // History
		m.history = config.LoadHistory()
		m.state = stateHistory
		m.cursor = 0
		return m, nil
	case 7: // Settings
		m.state = stateSettings
		m.settingsCursor = 0
		m.settingsEditing = false
		return m, nil
	default: // Sources (0-4)
		m.selectedSource = m.cursor
		m.state = stateExtractInput
		m.urlInput.SetValue("")
		m.urlInput.Focus()
		return m, textinput.Blink
	}
}

// ── View ──

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	switch m.state {
	case stateHome:
		return m.viewHome()
	case stateExtractInput:
		return m.viewExtractInput()
	case stateSqueezing:
		return m.viewSqueezing()
	case stateResult:
		return m.viewResult()
	case stateHistory:
		return m.viewHistory()
	case stateSettings:
		return m.viewSettings()
	}
	return ""
}

// ════ HOME VIEW ════

func (m Model) viewHome() string {
	w := m.width
	if w > 76 {
		w = 76
	}

	var s strings.Builder

	// Logo
	logo := `
    ██████╗ ██╗   ██╗██╗     ██████╗
    ██╔══██╗██║   ██║██║     ██╔══██╗
    ██████╔╝██║   ██║██║     ██████╔╝
    ██╔═══╝ ██║   ██║██║     ██╔═══╝
    ██║     ╚██████╔╝███████╗██║
    ╚═╝      ╚═════╝ ╚══════╝╚═╝`
	s.WriteString(logoStyle.Render(logo))
	s.WriteString("\n")
	s.WriteString(lipgloss.PlaceHorizontal(w, lipgloss.Center,
		taglineStyle.Render("squeeze the web into clean markdown for AI")))
	s.WriteString("\n\n")

	// Menu
	var menuContent strings.Builder
	for i, item := range menuItems {
		icon := item.icon
		name := item.name
		desc := item.desc

		if i == m.cursor {
			line := fmt.Sprintf("%s  %-24s %s", icon, name, menuDescStyle.Render(desc))
			menuContent.WriteString(selectedItemStyle.Render(line))
		} else {
			line := fmt.Sprintf("%s  %-24s %s", icon, name, menuDescStyle.Render(desc))
			menuContent.WriteString(normalItemStyle.Render(line))
		}
		menuContent.WriteString("\n")
	}

	menuPanel := panelStyle.Width(w - 4).Render(menuContent.String())
	s.WriteString(menuPanel)
	s.WriteString("\n\n")

	// Quick squeeze bar
	quickContent := "  Quick Squeeze:  " + m.quickInput.View()
	quickPanel := panelStyle.Width(w - 4).Render(quickContent)
	s.WriteString(quickPanel)
	s.WriteString("\n\n")

	// Key hints
	hints := keyHint("↑↓", "Navigate") + "   " +
		keyHint("Enter", "Select") + "   " +
		keyHint("/", "Quick squeeze") + "   " +
		keyHint("q", "Quit")
	s.WriteString("  " + hints)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s.String())
}

// ════ EXTRACT INPUT VIEW ════

func (m Model) viewExtractInput() string {
	w := m.width
	if w > 76 {
		w = 76
	}

	var s strings.Builder

	// Header
	item := menuItems[m.selectedSource]
	header := headerStyle(w).Render(fmt.Sprintf("  pulp  ›  %s", strings.ToLower(item.name)))
	s.WriteString(header)
	s.WriteString("\n\n")

	// URL input
	urlContent := "\n  " + m.urlInput.View() + "\n"
	urlPanel := activePanelStyle.Width(w - 4).
		BorderForeground(colorOrange).
		Render(urlContent)
	s.WriteString("  " + lipgloss.NewStyle().Bold(true).Foreground(colorOrange).Render("Enter URL") + "\n")
	s.WriteString(urlPanel)
	s.WriteString("\n\n")

	// Platform detection
	url := m.urlInput.Value()
	if url != "" {
		detected := detectSourceName(url)
		engine := detectEngine(url)
		detContent := fmt.Sprintf("  ◉ Detected: %s\n  Engine:   %s", detected, engine)
		detPanel := panelStyle.Width(w - 4).Render(detContent)
		s.WriteString(detPanel)
		s.WriteString("\n\n")
	}

	// CTA
	cta := lipgloss.NewStyle().
		Background(colorOrange).
		Foreground(colorSurface).
		Bold(true).
		Padding(0, 3).
		Render("  Squeeze  🍊  ")
	s.WriteString(lipgloss.PlaceHorizontal(w, lipgloss.Center, cta))
	s.WriteString("\n\n")

	// Hints
	hints := keyHint("Enter", "Squeeze") + "   " + keyHint("ESC", "Back")
	s.WriteString("  " + hints)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s.String())
}

// ════ SQUEEZING VIEW ════

func (m Model) viewSqueezing() string {
	w := m.width
	if w > 76 {
		w = 76
	}

	var s strings.Builder

	header := headerStyle(w).Render("  pulp  ›  squeezing...")
	s.WriteString(header)
	s.WriteString("\n\n")

	// Info
	item := menuItems[m.selectedSource]
	s.WriteString(statLabelStyle.Render("  Source   ") + lipgloss.NewStyle().Foreground(colorBright).Render(m.squeezeURL) + "\n")
	s.WriteString(statLabelStyle.Render("  Engine   ") + lipgloss.NewStyle().Foreground(colorBright).Render(detectEngine(m.squeezeURL)) + "\n\n")

	// Spinner
	s.WriteString("  " + m.spinner.View() + " " + lipgloss.NewStyle().Foreground(colorOrange).Bold(true).Render("Squeezing "+item.icon+"...") + "\n\n")

	// Progress
	s.WriteString("  " + m.progress.ViewAs(0.5) + "\n\n")

	// Log
	logContent := fmt.Sprintf("  %s  Fetching content...\n  %s  Running extraction engine...",
		m.spinner.View(),
		lipgloss.NewStyle().Foreground(colorMuted).Render("◌"),
	)
	logPanel := panelStyle.Width(w - 4).Render(logContent)
	s.WriteString(logPanel)
	s.WriteString("\n\n")

	hints := keyHint("Ctrl+C", "Cancel")
	s.WriteString("  " + hints)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s.String())
}

// ════ RESULT VIEW ════

func (m Model) viewResult() string {
	w := m.width
	if w > 76 {
		w = 76
	}

	var s strings.Builder

	// Header
	badge := ""
	if m.squeezeErr != nil {
		badge = errorBadge.Render(" ✗ Error ")
	} else {
		badge = successBadge.Render(fmt.Sprintf(" 🍊 Squeezed in %s ", m.squeezeDur.Round(time.Millisecond)))
	}
	header := headerStyle(w).Render("  pulp  ›  result") + "  " + badge
	s.WriteString(header)
	s.WriteString("\n\n")

	if m.squeezeErr != nil {
		errBox := lipgloss.NewStyle().
			Foreground(colorRed).
			Render("  " + m.squeezeErr.Error())
		s.WriteString(errBox)
	} else {
		// Preview
		words := len(strings.Fields(m.squeezeOutput))
		lines := len(strings.Split(m.squeezeOutput, "\n"))

		// Juice Report
		juiceContent := fmt.Sprintf("  Words: %s   Lines: %s",
			statValueStyle.Render(fmt.Sprintf("%d", words)),
			statValueStyle.Render(fmt.Sprintf("%d", lines)),
		)
		juicePanel := panelStyle.Width(w - 4).Render(juiceContent)
		s.WriteString(juicePanel)
		s.WriteString("\n\n")

		// Viewport preview
		vw := w - 6
		vh := m.height - 18
		if vh < 5 {
			vh = 5
		}
		m.viewport.Width = vw
		m.viewport.Height = vh

		previewPanel := panelStyle.Width(w - 4).Render(m.viewport.View())
		s.WriteString(previewPanel)
	}

	s.WriteString("\n\n")

	// Actions
	actionContent := keyHint("S", "Save") + "   " +
		keyHint("C", "Copy") + "   " +
		keyHint("R", "Re-squeeze") + "   " +
		keyHint("ESC", "Back")
	s.WriteString("  " + actionContent)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, s.String())
}

// ════ HISTORY VIEW ════

func (m Model) viewHistory() string {
	w := m.width
	if w > 76 {
		w = 76
	}

	entries := m.history.Recent(30)

	var s strings.Builder

	// Total reflects all entries, not just the visible cap of 30
	total := len(m.history.Entries)
	header := headerStyle(w).Render(fmt.Sprintf("  pulp  ›  history                              %d squeezes", total))
	s.WriteString(header)
	s.WriteString("\n\n")

	icons := map[string]string{
		"extract": "🌐", "youtube": "▶ ", "instagram": "📸", "reddit": "🟠", "pdf": "📄",
	}

	if len(entries) == 0 {
		s.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Padding(1, 3).Render(
			"No squeezes yet. Run  pulp extract <url>  to get started."))
	} else {
		var listContent strings.Builder
		maxShow := m.height - 10
		if maxShow > len(entries) {
			maxShow = len(entries)
		}
		if maxShow < 1 {
			maxShow = 1
		}

		for i := 0; i < maxShow; i++ {
			e := entries[i]
			icon := icons[e.Source]
			if icon == "" {
				icon = "📝"
			}

			title := e.Title
			if title == "" {
				title = e.URL
			}
			if len(title) > 40 {
				title = title[:37] + "..."
			}

			// Relative time
			ago := relTime(e.Timestamp)
			meta := fmt.Sprintf("%d words  ·  %s", e.WordCount, ago)

			if i == m.cursor {
				line := fmt.Sprintf("%s  %-42s %s", icon, title, lipgloss.NewStyle().Foreground(colorMuted).Render(e.URL))
				listContent.WriteString(historySelectedStyle.Render(line))
				listContent.WriteString("\n")
				listContent.WriteString(historyMetaStyle.Render("   " + meta))
			} else {
				line := fmt.Sprintf("%s  %-42s", icon, title)
				listContent.WriteString(historyNormalStyle.Render(line))
				listContent.WriteString("\n")
				listContent.WriteString(historyMetaStyle.Render("   " + meta))
			}
			listContent.WriteString("\n\n")
		}

		panel := panelStyle.Width(w - 4).Render(listContent.String())
		s.WriteString(panel)
	}

	s.WriteString("\n\n")
	hints := keyHint("↑↓", "Navigate") + "   " +
		keyHint("D", "Delete") + "   " +
		keyHint("R", "Re-squeeze") + "   " +
		keyHint("ESC", "Back")
	s.WriteString("  " + hints)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, s.String())
}

// ════ SETTINGS VIEW ════

func (m Model) viewSettings() string {
	w := m.width
	if w > 76 {
		w = 76
	}

	var s strings.Builder

	header := headerStyle(w).Render("  pulp  ›  settings")
	s.WriteString(header)
	s.WriteString("\n\n")

	// Field labels and values
	type settingsField struct {
		label string
		value string
		hint  string
	}
	fields := []settingsField{
		{"Output directory", m.settingsValues[0], "free text path"},
		{"Default format", m.settingsValues[1], "md · skillzip · single"},
	}

	var defContent strings.Builder
	for i, f := range fields {
		isCurrent := i == m.settingsCursor
		isEditingThis := isCurrent && m.settingsEditing

		cursorStr := "  "
		if isCurrent {
			cursorStr = lipgloss.NewStyle().Foreground(colorOrange).Bold(true).Render("▸ ")
		}

		var valueStr string
		if isEditingThis {
			// Show the live text input
			valueStr = m.settingsInput.View()
		} else if isCurrent {
			valueStr = lipgloss.NewStyle().Foreground(colorBright).Bold(true).Render(f.value)
		} else {
			valueStr = lipgloss.NewStyle().Foreground(colorText).Render(f.value)
		}

		hintStr := ""
		if isCurrent && !m.settingsEditing {
			hintStr = "  " + lipgloss.NewStyle().Foreground(colorMuted).Italic(true).Render(f.hint)
		}

		defContent.WriteString(fmt.Sprintf("  %s%-20s %s%s\n", cursorStr, f.label, valueStr, hintStr))
	}

	defPanel := panelStyle.Width(w - 4).Render(defContent.String())
	s.WriteString("  " + lipgloss.NewStyle().Bold(true).Foreground(colorOrange).Render("Defaults") + "\n")
	s.WriteString(defPanel)
	s.WriteString("\n\n")

	// Tools
	var toolContent strings.Builder
	tools := []struct {
		name string
		bin  string
	}{
		{"defuddle", "defuddle"},
		{"yt-dlp", "yt-dlp"},
	}
	for _, t := range tools {
		_, err := exec.LookPath(t.bin)
		if err == nil {
			toolContent.WriteString(fmt.Sprintf("  %-18s %s\n", t.name, statValueStyle.Render("✓ installed")))
		} else {
			toolContent.WriteString(fmt.Sprintf("  %-18s %s\n", t.name, lipgloss.NewStyle().Foreground(colorRed).Render("✗ not found")))
		}
	}

	toolPanel := panelStyle.Width(w - 4).Render(toolContent.String())
	s.WriteString("  " + lipgloss.NewStyle().Bold(true).Foreground(colorOrange).Render("Tools") + "\n")
	s.WriteString(toolPanel)
	s.WriteString("\n\n")

	var hints string
	if m.settingsEditing {
		hints = keyHint("Enter", "Confirm") + "   " + keyHint("ESC", "Cancel edit")
	} else {
		hints = keyHint("↑↓", "Navigate") + "   " +
			keyHint("Enter", "Edit") + "   " +
			keyHint("←→", "Cycle format") + "   " +
			keyHint("S", "Save & exit") + "   " +
			keyHint("ESC", "Cancel")
	}
	s.WriteString("  " + hints)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, s.String())
}

// ── Squeeze command ──

func (m Model) startSqueeze() tea.Cmd {
	return func() tea.Msg {
		cmdName := menuItems[m.selectedSource].cmd
		start := time.Now()

		exe, err := os.Executable()
		if err != nil {
			exe = "pulp"
		}

		cmd := exec.Command(exe, cmdName, m.squeezeURL, "-q")
		out, err := cmd.CombinedOutput()
		dur := time.Since(start)

		if err != nil {
			return squeezeDoneMsg{
				output: "",
				err:    fmt.Errorf("%s", strings.TrimSpace(string(out))),
				dur:    dur,
			}
		}
		return squeezeDoneMsg{
			output: string(out),
			dur:    dur,
		}
	}
}

// ── Helpers ──

func detectSource(url string) int {
	u := strings.ToLower(url)
	switch {
	case strings.Contains(u, "youtube.com") || strings.Contains(u, "youtu.be"):
		return 1
	case strings.Contains(u, "instagram.com"):
		return 2
	case strings.Contains(u, "reddit.com"):
		return 3
	case strings.HasSuffix(u, ".pdf"):
		return 4
	default:
		return 0
	}
}

func detectSourceName(url string) string {
	idx := detectSource(url)
	return menuItems[idx].icon + " " + menuItems[idx].name
}

func detectEngine(url string) string {
	idx := detectSource(url)
	switch idx {
	case 1, 2:
		return "yt-dlp"
	case 3:
		return "reddit JSON API"
	case 4:
		return "go-pdf"
	default:
		return "defuddle → markdown cleaner"
	}
}

func relTime(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("Jan 02")
	}
}
