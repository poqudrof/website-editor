# Site Editor - Inline Editing System

A simple inline content editing system with Go backend and Next.js frontend.

## Prerequisites

- Go 1.21+ installed
- Node.js 18+ installed (for frontend)

## Backend Setup & Run

```bash
# Install dependencies (first time only)
go mod download

# Run the server
go run .
```

Server starts on `http://localhost:9000`

### Build binary (optional)

```bash
go build -o site-editor
./site-editor
```

## Frontend Setup

See `frontend-tutorial.md` for complete Next.js integration.

Quick version:

```bash
# In your Next.js project
npm install

# Run dev server
npm run dev
```

Frontend runs on `http://localhost:3000`

## Usage

1. Start Go backend server (port 9000)
2. Start Next.js frontend (port 3000)
3. Add `data-editable="page:element"` to any HTML element
4. Double-click to edit, click outside to save
5. Changes persist in SQLite database

## API Endpoints

### GET `/api/content/:id`
Returns content with differentiation between original and edited versions.

**Response:**
```json
{
  "id": "home:title",
  "content": "Edited Text",           // Display this (edited if exists, else original)
  "original_content": "Original Text", // From HTML
  "edited_content": "Edited Text",    // User modifications
  "is_edited": true,                  // Whether user has edited
  "updated_at": 1760820931
}
```

### PUT `/api/content/:id`
Save or update content. Automatically differentiates original vs edited.

**Request:**
```json
{
  "content": "User Edited Text",       // Required: The edited content
  "original_content": "Original HTML"  // Optional: Send only on first edit
}
```

**Examples:**
```bash
# Get content
curl http://localhost:9000/api/content/home:title

# First edit (save both original and edited)
curl -X PUT http://localhost:9000/api/content/home:title \
  -H "Content-Type: application/json" \
  --data-raw '{"content":"My New Title","original_content":"Welcome"}'

# Subsequent edits (only update edited content)
curl -X PUT http://localhost:9000/api/content/home:title \
  -H "Content-Type: application/json" \
  --data-raw '{"content":"Updated Title"}'
```

## Database

SQLite database file: `content.db` (auto-created on first run)

### Schema
Each editable element stores:
- `id` - Unique identifier (e.g., "home:title")
- `original_content` - Initial content from HTML/JSX
- `edited_content` - User-modified content
- `is_edited` - Boolean flag indicating if user has edited
- `updated_at` - Unix timestamp

### How It Works
The system preserves the original HTML content while tracking user edits separately:

1. **Before editing**: Shows original HTML content
2. **First edit**: Stores both original + edited versions
3. **After editing**: Shows edited version
4. **Future feature**: Can revert to original anytime

## Project Structure

```
site-editor/
├── backend/          # This directory
│   ├── main.go      # Server setup (Fiber + port 9000)
│   ├── db.go        # Database models (GORM + SQLite)
│   ├── handlers.go  # API handlers (GET/PUT)
│   ├── go.mod       # Dependencies
│   └── content.db   # SQLite database (auto-generated)
├── frontend-tutorial.md  # Next.js integration guide
└── README.md        # This file
```
