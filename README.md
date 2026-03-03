# Margin

_Write in the margins of the web_

A web annotation layer built on [AT Protocol](https://atproto.com) that lets you annotate, highlight, and bookmark any URL on the internet.

## Project Structure

```
margin/
├── lexicons/           # AT Protocol lexicon schemas
│   └── at/margin/
│       ├── annotation.json
│       ├── bookmark.json
│       ├── collection.json
│       ├── collectionItem.json
│       ├── highlight.json
│       ├── like.json
│       ├── reply.json
│       ├── apikey.json
│       ├── preferences.json
│       └── profile.json
├── backend/            # Go API server
│   ├── cmd/server/
│   └── internal/
├── web/                # Astro SSR + React web app
│   └── src/
├── extension/          # Browser extension (WXT)
│   └── src/
└── avatar/             # Cloudflare Worker for avatar proxying
```

## Getting Started

### Docker (Recommended)

Run the full stack with Docker:

```bash
docker compose up -d --build
```

This builds both the Go backend and the Astro frontend into a single container. The Astro SSR server handles all frontend routing, static assets, and OG image generation, while the Go backend serves the API internally.

### Development

#### Backend

```bash
cd backend
go mod tidy
go run ./cmd/server
```

API server runs on http://localhost:8081

#### Web App

```bash
cd web
bun install
bun run dev
```

Dev server runs on http://localhost:4321 and proxies API requests to the backend.

#### Browser Extension

Built with [WXT](https://wxt.dev):

```bash
cd extension
bun install
bun run dev          # Chrome dev mode
bun run dev:firefox  # Firefox dev mode
```

## Architecture

In production, a single Docker container runs both services:

- **Astro SSR** (port 8080, public) — serves the web app, handles SSR for OG meta tags, generates dynamic OG images via satori, and proxies API/auth requests to the backend.
- **Go API** (port 8081, internal) — handles all API endpoints, OAuth, firehose ingestion, and data storage.

## Domain

**Domain**: `margin.at`
**Lexicon Namespace**: `at.margin.*`

## Tech Stack

- **Backend**: Go + Chi + SQLite
- **Frontend**: Astro 5 (SSR) + React 19 + Tailwind CSS
- **OG Images**: satori + @resvg/resvg-js
- **Extension**: WXT + React + Tailwind CSS
- **Protocol**: AT Protocol (Bluesky)

## Sponsors

Thank you for making Margin possible!

<!-- sponsors --><!-- sponsors -->

## License

MIT
