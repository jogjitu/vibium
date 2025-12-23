# CLAUDE.md

Browser automation for AI agents and humans.

## Key Docs

- V1-ROADMAP.md — Implementation plan with daily prompts
- FILESYSTEM.md — Target project structure
- V2-ROADMAP.md — Deferred features (don't build these yet)

## Reference Docs
- docs/reference/WebDriver-Bidi-Spec.md — Defines the BiDirectional WebDriver Protocol, a mechanism for remote control of user agents

## Current Goal

V1 by Christmas. Start with Day 1 in V1-ROADMAP.md.

## Tech Stack

- Go (clicker binary)
- TypeScript (JS client)
- WebDriver BiDi protocol
- MCP server (stdio)

## Design Philosophy

Optimize for first-time user/developer joy. Defaults should create an "aha!" moment:
- Browser visible by default (see what the AI is doing)
- Screenshots save to a sensible location automatically
- Zero config needed to get started

Power users can override defaults (headless mode, custom paths, etc.) when needed.

## Rules

- Follow V1-ROADMAP.md one milestone at a time
- Follow FILESYSTEM.md for filename and file paths guidance (unless there's a good reason not to.)
- Verify each checkpoint before moving on
- Don't add code from future milestones
- After each completed milestone:
  - Have a human verify the milestone is complete.
  - Git commit the changes
- When adding new command line options to the bin/clicker binary, be sure to add a simple example and sample output (or short description)
