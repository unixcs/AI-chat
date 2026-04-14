# Project Quick Reference

## Purpose

This repository is a front-end/back-end separated AI chat platform with membership management.

Primary business capabilities:
- User registration and login via phone number
- Membership activation via redeem code
- AI chat with SSE streaming output
- Conversation and message history
- Admin backend for users, roles, menus, membership, redeem codes, redeem records, and conversations

The current prompt configuration indicates the AI experience has been customized toward tarot interpretation.

## Tech Stack

### Frontend
- Vue 3
- Vue Router 4
- Pinia
- Vite 8
- Axios
- Day.js
- Sass

### Backend
- Node.js
- Express 5
- JWT with `jsonwebtoken`
- `bcryptjs`
- `dotenv`
- `cors`
- Day.js
- SQLite via Node built-in `node:sqlite`

## Runtime Requirements

- Node.js `>= 22`
  Reason: backend uses `node:sqlite`
- npm `>= 10` recommended

## Local Startup

### Backend
Working directory: `backend/`

Install:
```bash
npm install
```

Run:
```bash
npm run dev
```

Default URL:
```text
http://localhost:3001
```

Required environment file:
`backend/.env`

Important variables:
```env
DEEPSEEK_API_KEY=
DEEPSEEK_MODEL=deepseek-chat
DEEPSEEK_BASE_URL=https://api.deepseek.com/v1
MODEL_CONCURRENCY=6
MODEL_QUEUE_MAX=50
DEEPSEEK_SYSTEM_PROMPT_FILE=
```

### Frontend
Working directory: `frontend/`

Install:
```bash
npm install
```

Run:
```bash
npm run dev
```

Default URL:
```text
http://localhost:5173
```

Vite proxy:
- `/api` -> `http://localhost:3001`

## Top-Level Structure

```text
demo/
├─ README.md
├─ PROJECT_QUICK_REFERENCE.md
├─ restart-backend.bat
├─ backend/
│  ├─ server.js
│  ├─ db.js
│  ├─ auth.js
│  ├─ package.json
│  ├─ prompts/
│  │  └─ Prompt.md
│  └─ data.sqlite
└─ frontend/
   ├─ package.json
   ├─ vite.config.js
   ├─ src/
   │  ├─ main.js
   │  ├─ router/
   │  │  └─ index.js
   │  ├─ stores/
   │  ├─ api/
   │  └─ views/
   └─ tests/
```

## Key Entry Points

### Backend
- `backend/server.js`
  Main Express app, routing, DeepSeek integration, SSE streaming, queueing, membership checks
- `backend/db.js`
  SQLite initialization, schema creation, seed data, read/write helpers
- `backend/auth.js`
  JWT signing, verification, user/admin auth middleware, user session validation
- `backend/prompts/Prompt.md`
  Current business-specific system prompt content

### Frontend
- `frontend/src/main.js`
  Vue app bootstrap, Pinia, router, theme init
- `frontend/src/router/index.js`
  Route map and auth guards
- `frontend/src/views/user/ChatView.vue`
  Main user chat UI
- `frontend/src/stores/auth.js`
  Token and profile state
- `frontend/src/stores/chat.js`
  Conversation state, streaming lifecycle, message state
- `frontend/src/stores/chat-streaming.js`
  Streaming message helper utilities

## Route Map

### Public
- `/`
- `/login`
- `/register`

### User
- `/app/chat`
- `/app/profile`
- `/app/settings` redirects to `/app/profile`

### Admin
- `/admin/login`
- `/admin`
- `/admin/users`
- `/admin/roles`
- `/admin/menus`
- `/admin/members`
- `/admin/redeem-codes`
- `/admin/redeem-records`
- `/admin/conversations`

## Core Backend Flows

### Authentication
- User and admin both use JWT
- User requests pass through `userAuth`
- Admin requests pass through `adminAuth`
- User session validity also checks `currentSessionId`
- If token session does not match stored session, user is kicked out as multi-device login conflict

Files to inspect:
- `backend/auth.js`
- relevant login endpoints in `backend/server.js`

### Membership
- Membership is stored on user as `memberExpireAt`
- Chat is blocked when membership is missing or expired
- Redeem code activation extends or sets membership duration

Files to inspect:
- `backend/server.js`
- `frontend/src/views/user/ChatView.vue`

### AI Chat Streaming
- Frontend sends chat request and opens a stream
- Backend calls DeepSeek `/chat/completions` with `stream: true`
- Backend relays output as SSE
- Frontend appends partial deltas to the local assistant placeholder message

Files to inspect first:
- `backend/server.js`
- `frontend/src/stores/chat.js`
- `frontend/src/stores/chat-streaming.js`
- `frontend/src/views/user/ChatView.vue`

### Conversation History
- Conversations and messages are persisted in SQLite
- Frontend fetches conversation list first, then active conversation messages
- Admin side can inspect complete conversation contents

Files to inspect:
- `backend/server.js`
- `backend/db.js`
- `frontend/src/stores/chat.js`

## Database Notes

Database file:
- `backend/data.sqlite`

SQLite WAL side files may appear during runtime:
- `backend/data.sqlite-wal`
- `backend/data.sqlite-shm`

Current schema includes at least these tables:
- `users`
- `roles`
- `menus`
- `redeemCodes`
- `redeemRecords`
- `conversations`
- `messages`

The database is initialized in `backend/db.js`.

## Prompt Configuration

Prompt source behavior in `backend/server.js`:
- If `DEEPSEEK_SYSTEM_PROMPT_FILE` is set, backend loads that file
- Otherwise it falls back to a generic Chinese assistant prompt

Current project-specific prompt file:
- `backend/prompts/Prompt.md`

Current prompt content is tarot-reading oriented, with strict plain-text formatting constraints.

## Frontend State Notes

### Auth Store
File: `frontend/src/stores/auth.js`

Responsibilities:
- Persist `userToken` in localStorage
- Persist `adminToken` in localStorage
- Hold current `profile`
- Logout handlers for user and admin

### Chat Store
File: `frontend/src/stores/chat.js`

Responsibilities:
- Manage conversation list
- Track active conversation
- Cache messages by conversation id
- Track streaming state and stream handle
- Reset UI state on logout/session invalidation

Important behaviors:
- Creates a local user message immediately before stream starts
- Creates an empty local assistant placeholder
- Appends streaming deltas into that placeholder
- Removes the placeholder on stream error if it is still empty
- Clears tokens and redirects to `/login` on forced session invalidation

### Streaming Helpers
File: `frontend/src/stores/chat-streaming.js`

Responsibilities:
- Append text deltas to the correct assistant message
- Remove an empty assistant placeholder safely

There is a matching test file:
- `frontend/tests/chat-streaming.test.js`

## Current Worktree Situation

At the time this note was written, the repository was not clean.

Observed modified files:
- `backend/data.sqlite-shm`
- `backend/data.sqlite-wal`
- `backend/prompts/Prompt.md`
- `frontend/src/stores/chat.js`

Observed untracked files:
- `frontend/package-lock.json`
- `frontend/src/stores/chat-streaming.js`
- `frontend/tests/`

Interpretation:
- There is in-progress work around frontend streaming behavior and tests
- SQLite WAL files are runtime artifacts, not necessarily meaningful source edits

Before making changes next time, run:
```bash
git status --short --branch
```

## Likely Files To Open First By Task

### Fix login or auth issue
1. `backend/auth.js`
2. `backend/server.js`
3. `frontend/src/stores/auth.js`
4. `frontend/src/router/index.js`

### Fix chat stream issue
1. `frontend/src/stores/chat.js`
2. `frontend/src/stores/chat-streaming.js`
3. `frontend/src/views/user/ChatView.vue`
4. `backend/server.js`

### Fix membership or redeem logic
1. `backend/server.js`
2. `backend/db.js`
3. admin/user related views in `frontend/src/views/`

### Change AI persona or output style
1. `backend/prompts/Prompt.md`
2. `backend/server.js`
3. `backend/.env`

### Fix admin page navigation or page access
1. `frontend/src/router/index.js`
2. relevant file under `frontend/src/views/admin/`
3. `frontend/src/stores/auth.js`

### Investigate database or seed data issues
1. `backend/db.js`
2. `backend/data.sqlite`
3. any related endpoint in `backend/server.js`

## Notable Risks And Caveats

### Hard-coded credentials and secrets
- `backend/auth.js` contains a hard-coded JWT secret: `demo_secret_key_2026`
- `backend/db.js` seed data contains default user/admin credentials

These are acceptable for local demo work but unsafe for production.

### No obvious unified test script
- Root `package.json` does not provide meaningful project scripts
- Frontend currently has a test file, but `frontend/package.json` does not yet show a `test` script

### Large backend file
- `backend/server.js` is the main integration point and is large
- Most behavior changes will end up touching this file

## Fast Re-Orientation Checklist

When returning to this repo later, use this sequence:

1. Check worktree state with `git status --short --branch`
2. Read this file first
3. Open `README.md` only if deployment or environment details are needed
4. Open the task-specific entry files listed in "Likely Files To Open First By Task"
5. If the issue is chat-related, inspect `frontend/src/stores/chat.js` and `backend/server.js` before anything else

## Suggested Next Documentation Updates

If the codebase changes later, update this file when any of the following changes:
- Route structure
- Auth flow
- Membership rules
- Prompt source or AI provider
- Database schema
- Primary files for chat streaming

## Document Intent

This file is not a full specification.
It is a rapid handoff and re-orientation note so future work can start with targeted reads instead of re-scanning the whole codebase.
