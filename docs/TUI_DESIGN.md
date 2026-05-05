# Pulp — TUI Design Specification

> Built with Bubble Tea + Lip Gloss (Charmbracelet ecosystem)
> Developed in Antigravity IDE by Google

---

## Design Philosophy

**Aesthetic: Terminal Noir** — Dark, focused, command-driven. Inspired by classic UNIX tools but elevated with modern Lip Gloss styling. Every pixel of terminal real estate is intentional.

Colors lean into Pulp's identity — orange is the soul of pulp (citrus, raw material, energy).

- **Primary accent:** `#F97316` (orange) — pulp, raw, alive
- **Secondary accent:** `#10B981` (emerald) — extraction success
- **Warning:** `#F59E0B` (amber)
- **Error:** `#EF4444` (red)
- **Muted text:** `#6B7280`
- **Surface:** `#111827` (near-black bg assumed)
- **Border:** `#374151`
- **Highlight border:** `#F97316`

---

## Screen Map

```
┌─────────────────┐
│   Home / Menu   │──────────────────────────────────┐
└────────┬────────┘                                   │
         │                                            │
    ┌────┴────┐                                       │
    │ Squeeze │                                       │
    └────┬────┘                                       │
         │                                            ▼
    ┌────┴──────────────┐              ┌──────────────────────┐
    │  Squeezing View   │              │    History Panel     │
    │  (progress + log) │              │  (past squeezes)     │
    └────┬──────────────┘              └──────────────────────┘
         │
    ┌────┴──────────┐
    │  Result View  │
    │ (preview + save)│
    └───────────────┘
```

---

## 1. Home Screen (Main Menu)

```
╔══════════════════════════════════════════════════════════════════════╗
║                                                                      ║
║      ██████╗ ██╗   ██╗██╗     ██████╗                               ║
║      ██╔══██╗██║   ██║██║     ██╔══██╗                              ║
║      ██████╔╝██║   ██║██║     ██████╔╝                              ║
║      ██╔═══╝ ██║   ██║██║     ██╔═══╝                               ║
║      ██║     ╚██████╔╝███████╗██║                                   ║
║      ╚═╝      ╚═════╝ ╚══════╝╚═╝                                   ║
║                                                                      ║
║            squeeze the web into clean markdown for AI               ║
╚══════════════════════════════════════════════════════════════════════╝

  ╭──────────────────────────────────────────────────────────────────╮
  │                                                                  │
  │   ▸  Extract Web Page       Convert any URL to clean markdown   │
  │      YouTube Transcript     Pull video/shorts transcript        │
  │      Instagram Reel         Extract audio + caption             │
  │      Reddit Post            Post + top comments as markdown     │
  │      PDF Extraction         Extract text from PDF file          │
  │      Package Skill          Create skill.zip from content       │
  │      History                View past squeezes                  │
  │      Settings               Configure defaults                  │
  │                                                                  │
  ╰──────────────────────────────────────────────────────────────────╯

  ╭──────────────────────────────────────────────────────────────────╮
  │  Quick Squeeze:  ›  _                                            │
  ╰──────────────────────────────────────────────────────────────────╯

  [↑↓] Navigate   [Enter] Select   [/] Quick squeeze   [q] Quit
```

**Implementation notes:**
- Logo rendered in `#F97316` bold — orange pops on any dark terminal
- Selected menu item: orange thick left border + bold white text
- Unselected: muted gray `#9CA3AF`
- Quick Squeeze input always focused at bottom — paste URL and hit Enter from anywhere
- Tagline: *"squeeze the web into clean markdown for AI"* — matches the brand

---

## 2. Extract View — URL Input

```
╔══════════════════════════════════════════════════════════════════════╗
║  pulp  ›  extract                                     [ESC] Back    ║
╚══════════════════════════════════════════════════════════════════════╝

  ╭── Enter URL ─────────────────────────────────────────────────────╮
  │                                                                  │
  │  🔗  https://docs.anthropic.com/en/docs/│                       │
  │                                                                  │
  ╰──────────────────────────────────────────────────────────────────╯

  ╭── Options ───────────────────────────────────────────────────────╮
  │                                                                  │
  │   Output format    ● Markdown   ○ skill.zip   ○ Single file     │
  │   References       [✓] Include images/PDFs                      │
  │   Output path      ~/brain/00-Inbox/                            │
  │   Tags             [ai] [docs]  + add tag                       │
  │                                                                  │
  ╰──────────────────────────────────────────────────────────────────╯

  ╭── Platform Detection ────────────────────────────────────────────╮
  │                                                                  │
  │  ◉ Detected: Web Page                                           │
  │    Engine:   defuddle                                            │
  │                                                                  │
  ╰──────────────────────────────────────────────────────────────────╯

                           [ Squeeze  🍊  ]

  [Tab] Next field   [Enter] Squeeze   [ESC] Back
```

**Implementation notes:**
- Platform detector runs on-the-fly as user types (debounced 400ms)
- Platform badge cycles icon: 🌐 Web · ▶ YouTube · 📸 Instagram · 🟠 Reddit · 📄 PDF
- Options panel uses `huh` from Charmbracelet for form fields
- CTA button says **Squeeze** — branded, consistent with the pulp metaphor
- Orange 🍊 emoji on the button is a subtle delight detail

---

## 3. Extraction Progress View

```
╔══════════════════════════════════════════════════════════════════════╗
║  pulp  ›  squeezing...                                               ║
╚══════════════════════════════════════════════════════════════════════╝

  Source   https://docs.anthropic.com/en/docs/build-with-claude/...
  Engine   defuddle → markdown cleaner
  Output   ~/brain/00-Inbox/build-with-claude.md

  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
                          ████████████░░░░░░░░  62%
  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  ╭── Log ───────────────────────────────────────────────────────────╮
  │                                                                  │
  │  ✓  Fetched page HTML          234ms                            │
  │  ✓  Ran defuddle extraction    891ms                            │
  │  ✓  Cleaned markdown           12ms                             │
  │  ◌  Downloading references...                                   │
  │     └── image1.png  ████████░░░░  67%                          │
  │     └── diagram.svg  pending                                    │
  │                                                                  │
  ╰──────────────────────────────────────────────────────────────────╯

  ╭── Juice Report ──────────────────────────────────────────────────╮
  │  Raw HTML     48,291 tokens    Clean MD   3,847 tokens   -92%   │
  ╰──────────────────────────────────────────────────────────────────╯

  [Ctrl+C] Cancel
```

**Implementation notes:**
- Stats panel called **Juice Report** — on-brand
- Spinner: `spinner.Dot` style, orange tint
- Progress bar: orange → emerald gradient fill
- Token reduction `-92%` rendered in emerald — the payoff stat
- Runs as a `tea.Cmd` goroutine, sends `ProgressMsg` events to update model

---

## 4. Result Preview View

```
╔══════════════════════════════════════════════════════════════════════╗
║  pulp  ›  result                          🍊 Squeezed in 1.2s       ║
╚══════════════════════════════════════════════════════════════════════╝

  ╭── build-with-claude.md ──────────────────────────────── 3,847 tok ╮
  │                                                                   │
  │  ---                                                              │
  │  name: build-with-claude                                          │
  │  source: https://docs.anthropic.com/...                           │
  │  created: 2026-05-05                                              │
  │  tags: [ai, docs, anthropic]                                      │
  │  ---                                                              │
  │                                                                   │
  │  # Build with Claude                                              │
  │                                                                   │
  │  Claude is designed to be helpful, harmless, and honest...       │
  │  The API exposes a simple interface for sending messages...       │
  │                                                                   │
  │  ## Getting Started                                               │
  │                                                                   │
  │  To begin, install the Anthropic SDK:                             │
  │                                                                   │
  │  ```bash                                                          │
  │  pip install anthropic                                            │
  │  ```                                                              │
  │                                                          [↓ more] │
  ╰───────────────────────────────────────────────────────────────────╯

  ╭── Actions ───────────────────────────────────────────────────────╮
  │                                                                  │
  │  [S] Save to output path      [P] Package as skill.zip          │
  │  [C] Copy to clipboard        [E] Edit in $EDITOR               │
  │  [R] Re-squeeze               [H] Add to history                │
  │                                                                  │
  ╰──────────────────────────────────────────────────────────────────╯

  [↑↓ / PgUp PgDn] Scroll preview   [ESC] Back to menu
```

**Implementation notes:**
- Header badge: `🍊 Squeezed in 1.2s` — orange emoji, emerald time text
- Re-extract → **Re-squeeze** throughout
- Preview uses `viewport.Model` from Bubbles — scrollable
- Frontmatter highlighted in muted orange; code blocks in amber

---

## 5. History View

```
╔══════════════════════════════════════════════════════════════════════╗
║  pulp  ›  history                               12 squeezes         ║
╚══════════════════════════════════════════════════════════════════════╝

  ╭── Filter ────────────────────────────────────────────────────────╮
  │  🔍  _                                     [all ▾] [today ▾]    │
  ╰──────────────────────────────────────────────────────────────────╯

  ╭──────────────────────────────────────────────────────────────────╮
  │                                                                  │
  │  ▸  🌐  build-with-claude          docs.anthropic.com           │
  │         3,847 tok  ·  today 14:32  ·  md                        │
  │                                                                  │
  │     ▶  cursor-ai-tips              youtube.com/watch?v=...       │
  │         12,041 tok · today 11:20  ·  skill.zip                  │
  │                                                                  │
  │     📄  llm-paper.pdf             local file                    │
  │         8,203 tok  ·  yesterday   ·  md                         │
  │                                                                  │
  │     🟠  best-prompting-tips        reddit.com/r/LocalLLaMA      │
  │         2,109 tok  ·  2 days ago  ·  md                         │
  │                                                                  │
  │     📸  ai-workflow-demo           instagram.com/reel/...        │
  │         941 tok    ·  3 days ago  ·  md                         │
  │                                                                  │
  ╰──────────────────────────────────────────────────────────────────╯

  [Enter] Preview   [D] Delete   [R] Re-squeeze   [P] Package   [ESC] Back
```

---

## 6. Skill Packager View

```
╔══════════════════════════════════════════════════════════════════════╗
║  pulp  ›  package                                                    ║
╚══════════════════════════════════════════════════════════════════════╝

  ╭── Skill Details ─────────────────────────────────────────────────╮
  │                                                                  │
  │  Name         anthropic-docs                                     │
  │  Description  Anthropic developer documentation for AI agents   │
  │  Tags         [ai] [api] [docs]  + add                          │
  │                                                                  │
  ╰──────────────────────────────────────────────────────────────────╯

  ╭── Sources (drag to reorder) ─────────────────────────────────────╮
  │                                                                  │
  │  ✓  build-with-claude.md           3,847 tok                    │
  │  ✓  api-reference.md               6,210 tok                    │
  │  ✗  llm-paper.pdf                  8,203 tok  [excluded]        │
  │                                                                  │
  │                            Total:  10,057 tok                    │
  ╰──────────────────────────────────────────────────────────────────╯

  ╭── Output ────────────────────────────────────────────────────────╮
  │                                                                  │
  │  Format    ● skill.zip   ○ Single SKILL.md                      │
  │  Path      ~/.skills/                                            │
  │                                                                  │
  ╰──────────────────────────────────────────────────────────────────╯

                         [ 🍊 Press & Package ]

  [Space] Toggle source   [Enter] Confirm   [ESC] Cancel
```

**Implementation notes:**
- CTA: **Press & Package** — "press" as in juice press, on-brand
- Output dir `~/.skills/` — IDE-agnostic, works with Antigravity or anything else

---

## 7. Settings View

```
╔══════════════════════════════════════════════════════════════════════╗
║  pulp  ›  settings                                                   ║
╚══════════════════════════════════════════════════════════════════════╝

  ╭── Defaults ──────────────────────────────────────────────────────╮
  │                                                                  │
  │  Output directory     ~/brain/00-Inbox/                         │
  │  Default format       ● Markdown  ○ skill.zip  ○ Single file   │
  │  Include references   [✓]                                       │
  │  Auto-tag             [✓]  (guesses tags from content)          │
  │                                                                  │
  ╰──────────────────────────────────────────────────────────────────╯

  ╭── Integrations ──────────────────────────────────────────────────╮
  │                                                                  │
  │  brain CLI hook       [✓]  ~/.local/bin/brain                   │
  │  Skills directory     ~/.skills/                                 │
  │  $EDITOR              nvim                                       │
  │  IDE                  Antigravity                                │
  │                                                                  │
  ╰──────────────────────────────────────────────────────────────────╯

  ╭── Tools ─────────────────────────────────────────────────────────╮
  │                                                                  │
  │  defuddle             ✓ v0.4.1 installed                        │
  │  yt-dlp               ✓ 2026.03.15 installed                    │
  │  rg (ripgrep)         ✓ 14.1.0 installed                        │
  │                                                                  │
  ╰──────────────────────────────────────────────────────────────────╯

  [Enter] Edit field   [S] Save   [ESC] Cancel
```

---

## Component Spec

### Borders
```go
// Standard panel
panelStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("#374151")).
    Padding(0, 1)

// Active / focused panel
activePanelStyle = panelStyle.
    BorderForeground(lipgloss.Color("#F97316"))
```

### Menu Item
```go
// Selected
selectedItem = lipgloss.NewStyle().
    BorderLeft(true).
    BorderStyle(lipgloss.ThickBorder()).
    BorderForeground(lipgloss.Color("#F97316")).
    PaddingLeft(1).
    Bold(true).
    Foreground(lipgloss.Color("#F9FAFB"))

// Normal
normalItem = lipgloss.NewStyle().
    PaddingLeft(3).
    Foreground(lipgloss.Color("#9CA3AF"))
```

### Header Bar
```go
headerStyle = lipgloss.NewStyle().
    Background(lipgloss.Color("#F97316")).
    Foreground(lipgloss.Color("#111827")).  // dark text on orange
    Bold(true).
    Padding(0, 2).
    Width(termWidth)
```

### Status Badges
```go
successBadge = lipgloss.NewStyle().
    Background(lipgloss.Color("#10B981")).
    Foreground(lipgloss.Color("#F9FAFB")).
    Bold(true).
    Padding(0, 1)

errorBadge = lipgloss.NewStyle().
    Background(lipgloss.Color("#EF4444")).
    Foreground(lipgloss.Color("#F9FAFB")).
    Bold(true).
    Padding(0, 1)

infoBadge = lipgloss.NewStyle().
    Background(lipgloss.Color("#374151")).
    Foreground(lipgloss.Color("#9CA3AF")).
    Padding(0, 1)
```

### Progress Bar
```go
prog = progress.New(
    progress.WithScaledGradient("#F97316", "#10B981"),  // orange → emerald
    progress.WithoutPercentage(),
)
```

### Key Hint Footer
```go
keyStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#F97316")).
    Bold(true)

descStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#6B7280"))
```

---

## Pulp Brand Language

Consistent copy throughout the TUI — all extraction language maps to the juice/pulp metaphor.

| Generic term       | Pulp term        |
|--------------------|------------------|
| Extract            | Squeeze          |
| Extracting...      | Squeezing...     |
| Re-extract         | Re-squeeze       |
| Extraction history | Squeezes         |
| Token stats panel  | Juice Report     |
| Package CTA        | Press & Package  |
| Quick extract bar  | Quick Squeeze    |

---

## State Machine (Bubble Tea Model)

```go
type state int

const (
    stateHome state = iota
    stateExtractInput
    stateSqueezing
    stateResult
    stateHistory
    statePackager
    stateSettings
)

type Model struct {
    state      state
    menuCursor int
    input      textinput.Model
    spinner    spinner.Model
    progress   progress.Model
    viewport   viewport.Model
    result     *SqueezeResult
    history    []SqueezeEntry
    width      int
    height     int
    err        error
}
```

---

## Responsive Behavior

| Terminal Width | Layout Adjustment |
|----------------|-------------------|
| < 80 cols      | Collapse options panel; hide stats sidebar |
| 80–120 cols    | Standard layout (all mockups above) |
| > 120 cols     | Two-column: preview left, options right |

Always handle `tea.WindowSizeMsg` and store `m.width / m.height` to reflow panels.

---

## Keyboard Reference

| Key | Action |
|-----|--------|
| `↑ / ↓` | Navigate menu |
| `Enter` | Select / confirm |
| `Tab` | Next field |
| `Shift+Tab` | Previous field |
| `/` | Focus quick squeeze input (from anywhere) |
| `ESC` | Go back |
| `Ctrl+C` | Cancel / quit |
| `q` | Quit from home |
| `h` | Jump to history |
| `s` | Jump to settings |
| `?` | Toggle help overlay |

---

## Help Overlay (modal)

```
  ╭── Help ──────────────────────────────────────────────────╮
  │                                                          │
  │  Navigation                                              │
  │  ↑ ↓         Move selection                             │
  │  Enter        Confirm / select                           │
  │  ESC          Back / cancel                              │
  │  /            Quick squeeze input                        │
  │                                                          │
  │  Global                                                  │
  │  h            History                                    │
  │  s            Settings                                   │
  │  ?            This help menu                             │
  │  Ctrl+C / q   Quit                                       │
  │                                                          │
  │                         [ Close ]                        │
  ╰──────────────────────────────────────────────────────────╯
```

Rendered as a centered overlay using `lipgloss.Place()`.

---

## Recommended Packages

```go
import (
    "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/spinner"
    "github.com/charmbracelet/bubbles/progress"
    "github.com/charmbracelet/bubbles/viewport"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/huh"
)
```
