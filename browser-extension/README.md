# Mathalama Focus Mode - Browser Extension

A browser extension to block distracting websites during your Mathalama Focus sessions.

## Features

- **Focus Mode**: Block distracting sites (YouTube, Reddit, Twitter, etc.) for a set duration
- **Customizable Duration**: Choose from 25, 45, 90 minutes or set custom duration
- **Site Management**: Enable/disable which sites to block
- **Visual Timer**: See remaining time in the blocked page overlay
- **Easy Control**: Start/stop focus sessions from the popup

## Blocked Sites by Default

- YouTube
- Reddit
- Twitter/X
- Facebook
- Instagram
- TikTok
- Twitch

## Installation

### Development Installation

1. Build the extension:
```bash
cd browser-extension
npm install
npm run build
```

2. Load into Chrome:
   - Go to `chrome://extensions/`
   - Enable "Developer mode"
   - Click "Load unpacked"
   - Select the `browser-extension/dist` folder

3. Load into Firefox:
   - Go to `about:debugging#/runtime/this-firefox`
   - Click "Load Temporary Add-on"
   - Select any file in `browser-extension/dist` folder

### Publishing to Stores

**Chrome Web Store:**
1. Create a Google account and developer account
2. Zip the contents of `dist/` folder
3. Upload to Chrome Web Store developer dashboard
4. Submit for review

**Firefox Add-ons:**
1. Create Mozilla account
2. Zip the contents of `dist/` folder
3. Upload to Firefox Add-ons developer hub
4. Submit for review

## Usage

1. Click the extension icon in your browser
2. Select desired focus duration (or enter custom)
3. Click "Start Focus"
4. Visit any blocked site - you'll see a blocking overlay
5. Click "Go Back" or confirm you want to continue
6. Click "End Focus" to stop early

## Architecture

- **manifest.json**: Extension configuration for Chrome/Firefox
- **service-worker.ts**: Background script managing focus sessions
- **content-script.ts**: Runs on every page to enforce blocking
- **blocker.ts**: Logic to detect and block sites
- **popup.html/ts**: User interface for starting/managing focus sessions

## How It Works

1. When you start a focus session, the background script records the session state
2. The content script checks every page load against blocked sites
3. If a blocked site matches, a full-page overlay prevents access
4. Timer counts down remaining session time
5. Session automatically ends after the set duration

## Development

```bash
# Watch mode for development
npm run dev

# Type checking
npm run type-check

# Build for production
npm run build
```

## Privacy

This extension:
- Stores all data locally in browser storage
- Never sends data to external servers (except for session start/end tracking to Mathalama back
