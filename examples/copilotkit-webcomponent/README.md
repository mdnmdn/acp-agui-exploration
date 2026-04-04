# CopilotKit Web Component

A web component wrapper for CopilotKit frontend that allows easy integration of AI chat capabilities into any web page.

## Features

- Easy to use as a custom HTML element
- Configurable via HTML attributes
- Supports three modes: `chat` (inline), `sidebar`, `popup`
- Connects to CopilotKit runtime or Copilot Cloud
- Event callbacks via property assignment or DOM events
- TypeScript types included

## Installation

```bash
npm install @acp-agui-bridge/copilotkit-webcomponent
```

## Usage

### Basic Usage

```html
<acp-copilotkit 
  url="http://localhost:4000/copilotkit"
  apikey="your-public-api-key"
  mode="sidebar"
  labels-title="My AI Assistant"
  labels-initial="Hello! How can I help you today?"
  labels-placeholder="Ask me anything..."
  default-open="true">
</acp-copilotkit>

<script type="module">
  import '@acp-agui-bridge/copilotkit-webcomponent';
</script>
```

### As a Script Tag

```html
<script type="module" src="https://unpkg.com/@acp-agui-bridge/copilotkit-webcomponent/dist/index.js"></script>

<acp-copilotkit 
  url="http://localhost:4000/copilotkit"
  mode="chat">
</acp-copilotkit>
```

## Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `url` | string | CopilotKit runtime URL |
| `apikey` | string | Public API key for Copilot Cloud (required for event callbacks) |
| `mode` | `'chat'` \| `'sidebar'` \| `'popup'` | Display mode (default: `sidebar`) |
| `labels-title` | string | Title of the assistant |
| `labels-initial` | string | Initial greeting message |
| `labels-placeholder` | string | Input placeholder text |
| `default-open` | boolean | Whether sidebar/popup opens by default |

## Event Callbacks

Two ways to listen for chat events. Both fire simultaneously when both are configured.

**Note:** `apikey` must be set for event callbacks to work (CopilotKit requirement).

### Option 1: Property Assignment

```js
const el = document.querySelector('acp-copilotkit');

el.callbacks = {
  onOpen: () => console.log('Chat opened'),
  onClose: () => console.log('Chat closed'),
  onMessageSent: (message) => console.log('User sent:', message),
  onFeedback: (messageId, type) => console.log('Feedback:', messageId, type),
  onCopy: (messageId) => console.log('Message copied:', messageId),
  onRegenerate: (messageId) => console.log('Regenerate:', messageId),
  onRewrite: (messageId) => console.log('Rewrite:', messageId),
  onGenerationChanged: (isGenerating) => console.log('Generating:', isGenerating),
};
```

### Option 2: DOM Event Listeners

```js
const el = document.querySelector('acp-copilotkit');

el.addEventListener('copilotkit:open', () => console.log('Chat opened'));
el.addEventListener('copilotkit:close', () => console.log('Chat closed'));
el.addEventListener('copilotkit:message-sent', (e) => console.log('Sent:', e.detail.message));
el.addEventListener('copilotkit:feedback', (e) => console.log('Feedback:', e.detail));
el.addEventListener('copilotkit:copy', (e) => console.log('Copied:', e.detail.messageId));
el.addEventListener('copilotkit:regenerate', (e) => console.log('Regenerate:', e.detail.messageId));
el.addEventListener('copilotkit:rewrite', (e) => console.log('Rewrite:', e.detail.messageId));
el.addEventListener('copilotkit:generation-changed', (e) => console.log('Generating:', e.detail.isGenerating));
```

## Modes

### Sidebar

```html
<acp-copilotkit mode="sidebar" url="..." default-open="true"></acp-copilotkit>
```

Renders a collapsible sidebar with CopilotKit's `CopilotSidebar` component.

### Chat (Inline)

```html
<acp-copilotkit mode="chat" url="..."></acp-copilotkit>
```

Renders an inline chat panel using CopilotKit's `CopilotChat` component. Requires a container with defined height.

### Popup

```html
<acp-copilotkit mode="popup" url="..." default-open="false"></acp-copilotkit>
```

Renders a floating popup button with chat using CopilotKit's `CopilotPopup` component.

## Development

```bash
# Install dependencies
npm install

# Build (types + bundle)
npm run build

# Dev server with HMR
npm run dev

# Preview production build
npm run preview
```

## Static Server Usage

The production bundle (`dist/index.js`) is fully self-contained with all dependencies bundled. No build tools or module resolvers needed.

### With miniserve

```bash
# Serve the example directory
miniserve examples/copilotkit-webcomponent/

# Then open http://localhost:8080/sample.html
```

### With any static server (nginx, Caddy, python, etc.)

```html
<!DOCTYPE html>
<html>
<head>
  <script type="module" src="/dist/index.js"></script>
</head>
<body>
  <acp-copilotkit 
    url="http://localhost:4000/copilotkit"
    apikey="your-public-key"
    mode="sidebar">
  </acp-copilotkit>
</body>
</html>
```

**Requirements:**
- Server must serve `.js` files with `Content-Type: application/javascript`
- No module bundler or import map needed — everything is pre-bundled

## CDN Usage

```html
<script type="module" src="https://unpkg.com/@acp-agui-bridge/copilotkit-webcomponent@1.0.0/dist/index.js"></script>

<acp-copilotkit 
  url="http://localhost:4000/copilotkit"
  mode="sidebar">
</acp-copilotkit>
```

## License

MIT
