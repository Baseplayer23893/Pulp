<div align="center">

```
██████╗ ██╗   ██╗██╗     ██████╗ 
██╔══██╗██║   ██║██║     ██╔══██╗
██████╔╝██║   ██║██║     ██████╔╝
██╔═══╝ ██║   ██║██║     ██╔═══╝ 
██║     ╚██████╔╝███████╗██║     
╚═╝      ╚═════╝ ╚══════╝╚═╝     
```

**squeeze the web into clean markdown for AI**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-orange.svg)](LICENSE)
[![Stars](https://img.shields.io/github/stars/Baseplayer23893/Pulp?style=flat&color=orange)](https://github.com/Baseplayer23893/Pulp/stargazers)

</div>

---

Pulp is a **free, open-source, local-first** alternative to createskills. Extract clean markdown from any webpage, YouTube video, Reddit post, Instagram Reel, or PDF — then package it as a skill for your AI workflows.

No API keys. No monthly fees. No data leaving your machine.

```bash
curl -sSL https://raw.githubusercontent.com/Baseplayer23893/Pulp/main/install.sh | bash
```

---

## Why Pulp?

| | Pulp | createskills |
|---|---|---|
| Price | **Free forever** | $10+/month |
| Open source | ✅ | ❌ |
| Local-first | ✅ | ❌ |
| YouTube transcripts | ✅ | Paywalled |
| MCP server | ✅ | Paywalled |
| API keys needed | **None** | Required |

---

## Features

- 🌐 **Web pages** — any URL → clean markdown via defuddle
- ▶️ **YouTube** — transcripts from any video or Short, no API key
- 🟠 **Reddit** — post + top comments as markdown
- 📸 **Instagram Reels** — audio transcription + caption
- 📄 **PDF** — extract text from any PDF file
- 📦 **Skill packaging** — bundle extractions into `skill.zip` for Cursor, Antigravity, or any AI IDE
- 🧠 **brain CLI hook** — integrates with your second brain workflow
- 🍊 **Beautiful TUI** — built with Bubble Tea + Lip Gloss

---

## Install

**One-line install (Linux/macOS):**
```bash
curl -sSL https://raw.githubusercontent.com/Baseplayer23893/Pulp/main/install.sh | bash
```

**Build from source:**
```bash
git clone https://github.com/Baseplayer23893/Pulp
cd Pulp
go build -o pulp ./main.go
sudo mv pulp /usr/local/bin/
```

**Dependencies:**
```bash
npm install -g defuddle   # web extraction
pip install yt-dlp        # YouTube transcripts
```

---

## Usage

```bash
pulp                          # open TUI
pulp extract <url>            # extract webpage
pulp youtube <url>            # YouTube transcript
pulp reddit <url>             # Reddit post + comments
pulp instagram <url>          # Instagram Reel
pulp pdf <file>               # extract PDF
pulp package <name>           # create skill.zip
```

**Quick squeeze any URL straight from the TUI** — just paste and hit Enter.

---

## Skill Packaging

Pulp outputs a `skill.zip` that works with any AI IDE:

```
my-research/
├── SKILL.md        ← clean content + frontmatter
└── references/     ← images, PDFs, audio
```

Drop it in your skills directory and your AI has the context.

---

## Screenshots

> TUI screenshots coming soon

---

## Roadmap

- [x] Web page extraction
- [x] YouTube transcripts (no API key)
- [x] Reddit posts
- [x] Instagram Reels
- [x] PDF extraction
- [x] Beautiful TUI (Bubble Tea)
- [x] Skill packaging
- [x] MCP server (v0.2)
- [ ] Web dashboard (v0.3)
- [ ] Cloud sync (future)

---

## Contributing

PRs, issues, and feature requests welcome. Check [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

MIT — free forever, do whatever you want with it.

---

<div align="center">
built with 🍊 by <a href="https://github.com/Baseplayer23893">Baseplayer23893</a>
</div>
