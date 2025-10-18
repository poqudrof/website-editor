# Frontend Tutorial - Inline Editing

## Quick Setup

### 1. Add editable attributes to your HTML/JSX

```jsx
<h1 data-editable="home:title">Welcome to my site</h1>
<p data-editable="home:subtitle">Your digital partner</p>
```

### 2. Create the EditableHandler component

Create `components/EditableHandler.tsx`:

```tsx
'use client';

import { useEffect } from 'react';

const API_URL = 'http://localhost:9000';

export default function EditableHandler() {
  useEffect(() => {
    // Load content from API and store original content
    document.querySelectorAll('[data-editable]').forEach(async (el) => {
      const id = el.getAttribute('data-editable');
      const originalContent = el.textContent || ''; // Store original HTML content

      // Store original in data attribute for later use
      el.setAttribute('data-original', originalContent);

      const res = await fetch(`${API_URL}/api/content/${id}`);
      const data = await res.json();

      // If content exists and has been edited, use it
      if (data.is_edited && data.content) {
        el.textContent = data.content;
      }
      // Otherwise keep the original HTML content
    });

    // Enable editing on double-click
    const handleDblClick = (e: Event) => {
      const el = e.target as HTMLElement;
      if (!el.hasAttribute('data-editable')) return;

      el.contentEditable = 'true';
      el.focus();
    };

    // Save on blur
    const handleBlur = async (e: Event) => {
      const el = e.target as HTMLElement;
      if (el.contentEditable !== 'true') return;

      el.contentEditable = 'false';
      const id = el.getAttribute('data-editable');
      const content = el.textContent || '';
      const originalContent = el.getAttribute('data-original') || '';

      // Send both edited and original content
      await fetch(`${API_URL}/api/content/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          content,           // The new edited content
          original_content: originalContent  // The original HTML content
        }),
      });
    };

    document.addEventListener('dblclick', handleDblClick);
    document.addEventListener('blur', handleBlur, true);

    return () => {
      document.removeEventListener('dblclick', handleDblClick);
      document.removeEventListener('blur', handleBlur, true);
    };
  }, []);

  return null;
}
```

### 3. Add to your layout

In `app/layout.tsx`:

```tsx
import EditableHandler from '@/components/EditableHandler';

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        <EditableHandler />
        {children}
      </body>
    </html>
  );
}
```

### 4. Run both servers

```bash
# Terminal 1 - Go backend
go run .

# Terminal 2 - Next.js frontend
npm run dev
```

## Usage

1. Double-click any element with `data-editable`
2. Edit the text
3. Click outside to save
4. Refresh - your changes persist!

## How It Works

The system differentiates between **original** and **edited** content:

### Original Content
- Comes from your HTML/JSX code
- Stored in the database when you first edit
- Never modified after initial save
- Used as fallback if no edits exist

### Edited Content
- User modifications via double-click editing
- Stored separately in the database
- Takes precedence when displaying
- Can be reverted to original (future feature)

### Database Structure
Each editable element stores:
- `id` - The data-editable identifier (e.g., "home:title")
- `original_content` - Initial HTML content
- `edited_content` - User-modified content
- `is_edited` - Boolean flag
- `updated_at` - Timestamp

### Example Flow
1. **First load**: "Welcome to my site" (from HTML)
2. **User edits**: "Welcome to our amazing site!"
3. **Database stores**:
   - original: "Welcome to my site"
   - edited: "Welcome to our amazing site!"
   - is_edited: true
4. **Next load**: Shows "Welcome to our amazing site!"

## That's it!
