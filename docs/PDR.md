# SkillForge - Product Design Document

> Open-source createskills alternative — extract clean markdown from web content and package for AI workflows.

---

## 1. Problem Statement

Current tools like createskills solve a real problem: turning noisy web content into clean, token-efficient markdown for AI models. However:

- Createskills is a **closed SaaS** with monthly fees ($10+/month for essential features)
- Data flows through their servers — no local-only option
- Some features (YouTube transcripts, MCP server) are **paywalled**
- No open-source alternative exists

**SkillForge** solves this by being fully open-source, local-first, and free forever.

---

## 2. Product Overview

### What is SkillForge?

A utility ecosystem that extracts clean markdown from web content and packages it for AI models, custom agents, and local LLM workflows.

### Target Users

- Developers using Cursor, Claude Code, OpenCode
- Researchers building AI knowledge bases
- Students using Obsidian for study notes
- Anyone running local LLM workflows

---

## 3. Feature Specification

### 3.1 Core Extraction Features

| Feature | Description | Priority |
|---------|-------------|----------|
| **Web Page Extraction** | Convert any webpage to clean markdown (uses defuddle under the hood) | P0 |
| **YouTube Transcripts** | Extract video transcripts (including Shorts, no API key needed) | P0 |
| **Instagram Reels** | Extract audio transcription + caption/description | P0 |
| **Reddit Posts** | Get Reddit post + top comments as markdown | P0 |
| **PDF Text Extraction** | Extract text from PDF files | P0 |
| **Element Picker** | Select specific DOM elements from a page | P1 |

### 3.2 Packaging Features

| Feature | Description | Priority |
|---------|-------------|----------|
| **Skill Zip** | Create `skill.zip` = `SKILL.md` + `references/` | P0 |
| **Single File Mode** | Append to a single skills.md file | P1 |
| **Hybrid Mode** | Flexible storage based on CLI flags | P1 |

### 3.3 Integration Features

| Feature | Description | Priority |
|---------|-------------|----------|
| **CLI** | Terminal-first command-line tool | P0 |
| **brain CLI Hook** | Integrate with existing brain CLI | P0 |
| **Web Dashboard** | Optional local web UI | P1 |
| **API Server** | REST API for automation | P1 |
| **MCP Server** | Model Context Protocol for Cursor/Claude Code | P2 |

---

## 4. CLI Commands

### Core Commands

```bash
skillforge extract <url>              # Extract web page → markdown
skillforge youtube <url>               # Get YouTube transcript
skillforge instagram <url>            # Extract Instagram Reel
skillforge reddit <url>               # Get Reddit post
skillforge pdf <file>                 # Extract text from PDF
skillforge package <name>             # Create skill.zip from extracted content
skillforge serve                      # Start local dashboard
skillforge install                    # Install CLI to PATH
```

### Optional Flags

```bash
--output, -o           # Output file location
--format, -f           # Output format: md, skillzip, single
--references, -r       # Include images/PDFs in references/
--quiet, -q            # Suppress verbose output
```

---

## 5. Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────┐
│                  User Interface                 │
│  ┌─────────┐  ┌─────────┐  ┌─────────────────┐ │
│  │   CLI   │  │ Web UI   │  │    API Server   │ │
│  └────┬────��  └────┬────┘  └───────┬───────┘ │
└───────┼─────────────┼────────────────┼─────────┘
        │             │                │
        └─────────────┼────────────────┘
                     │
              ┌──────┴──────┐
              │  Core Engine │
              │  ┌────────┐ │
              │  │Extract │ │
              │  │ +Clean │ │
              │  │   ↓   │ │
              │  │Package│ │
              │  └────────┘ │
              └─────────────┘
```

### Directory Structure

```
skillforge/
├── cli/                    # Main CLI tool
│   ├── main.go            # Entry point
│   ├── cmd/
│   │   ├── extract.go    # Web → markdown
│   │   ├── youtube.go     # YouTube transcript
│   │   ├── instagram.go  # Instagram Reel
│   │   ├── reddit.go     # Reddit post
│   │   ├── pdf.go        # PDF extraction
│   │   ├── package.go    # skill.zip creation
│   │   └── serve.go      # Dashboard server
│   └── internal/
│       ├── defuddle.go   # Wrapper for defuddle CLI
│       ├── cleaner.go    # Markdown post-processing
│       └── storage.go    # File management
├── dashboard/             # Optional web UI
│   ├── server.go
│   └── static/
├── mcp/                   # MCP server (future)
└── brain-integration/     # Hooks into brain CLI
    └── skillforge-process # brain skill
```

### Data Flow

```
URL Input
    ↓
[Platform Detector]
    ↓
┌─────────┬──────────┬────────┬─────────┐
│ YouTube │ Instagram│ Reddit │  PDF   │
│  ↓      │    ↓     │   ↓    │   ↓    │
└─────────┴──────────┴────────┴─────────┘
    ↓
[Content Extractor]
    ↓
[Markdown Cleaner]  ← Removes clutter, normalizes
    ↓
[Output Formatter]   ← SKILL.md or skill.zip
    ↓
File Output / CLI / API Response
```

---

## 6. Technical Implementation

### 6.1 Technology Stack

| Layer | Technology |
|-------|------------|
| Language | Go 1.21+ |
| CLI Framework | Cobra |
| Config | Viper |
| Web Framework | Chi |
| HTML Parsing | goquery |
| PDF Extraction | ledongthuc/pdf |
| Markdown | Custom + mmark |

### 6.2 External Dependencies

| Tool | Purpose | Source |
|------|---------|--------|
| defuddle | Web page → markdown | npm install -g defuddle |
| yt-dlp | YouTube transcript | pip install yt-dlp |
| rg | Reddit scraping | Native Go |
| instagram-scraper | IG Reels | Native Go |

### 6.3 Key Functions

```go
// Extract from any URL
func Extract(url string, opts Options) (string, error)

// YouTube transcript
func GetYouTubeTranscript(videoID string) (string, error)

// Instagram Reel
func GetInstagramReel(url string) (ReelContent, error)

// Reddit post
func GetRedditPost(url string) (RedditContent, error)

// PDF extraction
func ExtractPDF(filepath string) (string, error)

// Package as skill.zip
func PackageSkill(name string, content string, refs []string) error
```

---

## 7. Storage Format

### 7.1 SKILL.md Format

```markdown
---
name: skill-name
description: Brief description of what this is
created: 2026-05-04
source: https://example.com
tags: [research, ai]
---

# Skill Name

[Clean extracted content goes here]

## References

- [[references/image1.png]]
- [[references/document.pdf]]
```

### 7.2 skill.zip Structure

```
skill-name/
├── SKILL.md           # Frontmatter + content
└── references/       # Downloaded assets
    ├── image1.png
    ├── document.pdf
    └── audio.mp3
```

### 7.3 Single File Mode

```markdown
---
name: skills-collection
updated: 2026-05-04
---

# Skills

## 2026-05-04: Topic Name

Content here...

---
```

---

## 8. Integration with Second Brain

### 8.1 brain CLI Hook

Add to `~/.local/bin/brain`:

```bash
skillforge extract "$NOTE" -o ~/brain/00-Inbox/
```

### 8.2 Workflow

```
Web Research
    ↓
skillforge extract <url>
    ↓
skillforge package my-research
    ↓
Copy skill.zip to ~/.claude/skills/
    ↓
Claude Code / Cursor has context
```

### 8.3 AGENTS.md Integration

Update brain's AGENTS.md to use skillforge:

```markdown
## Tools

- skillforge: Extract web content for AI context
- defuddle: Alternative web extraction
```

---

## 9. Pricing & Business Model

| Tier | Price | Features |
|------|------|----------|
| **Open Source** | Free | All CLI features, self-hosted |
| **Cloud** | TBD | Managed hosting, sync across devices |
| **Enterprise** | TBD | Custom integrations, support |

**No paywalls. Everything is free forever.**

---

## 10. Milestones

### Phase 1: Core CLI (MVP)

- [x] Web page extraction (defuddle)
- [x] YouTube transcript
- [x] PDF extraction
- [x] Basic CLI structure
- [ ] Install command

### Phase 2: Multi-Platform

- [ ] Instagram Reel extraction
- [ ] Reddit post extraction
- [ ] Skill packaging (skill.zip)
- [ ] Element picker

### Phase 3: Integration

- [ ] brain CLI hook
- [ ] MCP server
- [ ] Web dashboard
- [ ] API server

---

## 11. Risks & Mitigation

| Risk | Mitigation |
|------|-----------|
| YouTube transcript rate limits | Use yt-dlp, implement retries + backoff |
| Instagram scraping blocks | Use official API or fallback to web scraping |
| PDF extraction quality | Test multiple libraries, choose best per format |
| Maintenance burden | Keep dependencies minimal, use proven tools |

---

## 12. Success Metrics

### User Acquisition

- GitHub stars: 100+ in first month
- HN/Reddit mentions: 10+ posts
- CLI installs: 500+ in first month

### Engagement

- Daily active users: 50+
- Commands per user: 10+/month
- skill.zip usage: 1000+/month

### Technical

- Uptime: 99.9%
- Extraction success rate: 95%+
- Average extraction time: <5s

---

## 13. Code of Conduct

be excellent to each other.

---

## 14. License

MIT License - See LICENSE file.

---

## 15. Contributing

Open to contributions! Submit PRs, issues, and feature requests.

---

## 16. Contact

- GitHub: github.com/Baseplayer23893/skillforge
- Issues: github.com/Baseplayer23893/skillforge/issues