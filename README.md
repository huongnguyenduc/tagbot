# Werewolf Bot

A Telegram bot for managing daily attendance polls for Werewolf game groups.

## Features

✅ **User Management**
- Automatic user registration on messages
- Member mentions with `@all` or `/tagall`
- User cleanup on leave/kick

## Commands

- `/help` - Show help message
- `/tagall` - Mention all members

## Project Structure

The codebase has been refactored into smaller, maintainable modules:

### Core Files

- **`main.go`** - Entry point and main event loop
- **`globals.go`** - Global variables and shared state
- **`database.go`** - Database operations and schema
- **`utils.go`** - Utility functions and helpers
- **`handlers.go`** - Command and event handlers
- **`webhook.go`** - Webhook server and HTTP handling

### File Responsibilities

#### `main.go`
- Bot initialization
- Command registration
- Main event loop
- Update routing

#### `globals.go`
- Global bot and database instances
- Scheduler state management
- Manual poll tracking state

#### `database.go`
- Database connection and initialization
- User management (save/delete)
- Group settings (timezone, poll times)
- Vote storage and retrieval
- Database schema creation

#### `utils.go`
- Text formatting (MarkdownV2 escaping)
- User mention generation
- Help text
- Group setting updates

#### `handlers.go`
- Command processing
- @all mention handling
- Chat member updates

#### `webhook.go`
- HTTP server for webhook mode
- Webhook setup and removal
- Request parsing and validation
- Health check endpoint

## Environment Variables

### Required
- `TELEGRAM_BOT_TOKEN` - Your Telegram bot token
- `DATABASE_URL` - PostgreSQL connection string

### Optional (Webhook Mode)
- `USE_WEBHOOK` - Set to "true" or "1" to enable webhook mode (default: polling)
- `WEBHOOK_URL` - Your public webhook URL (required if USE_WEBHOOK=true)
- `PORT` - Server port for webhook mode (default: 8080)

## Database Schema

The bot uses PostgreSQL with the following tables:

- `members` - Chat members and their information

## Building and Running

```bash
# Build the bot
go build -o werewolf-bot .

# Run in polling mode (default)
./werewolf-bot

# Run in webhook mode
USE_WEBHOOK=true WEBHOOK_URL=https://yourdomain.com/webhook ./werewolf-bot
```

## Docker

### Polling Mode
```bash
# Build Docker image
docker build -t werewolf-bot .

# Run with environment variables (polling mode)
docker run -e TELEGRAM_BOT_TOKEN=your_token -e DATABASE_URL=your_db_url werewolf-bot
```

### Webhook Mode
```bash
# Run with webhook (expose port 8080)
docker run -p 8080:8080 \
  -e TELEGRAM_BOT_TOKEN=your_token \
  -e DATABASE_URL=your_db_url \
  -e USE_WEBHOOK=true \
  -e WEBHOOK_URL=https://yourdomain.com/webhook \
  werewolf-bot
```

## Webhook Setup

### Requirements for Webhook Mode
1. **Public HTTPS URL** - Telegram requires HTTPS for webhooks
2. **Valid SSL Certificate** - Self-signed certificates won't work
3. **Accessible Port** - Default is 8080, configurable via PORT env var

### Webhook Endpoints
- `/webhook` - Receives Telegram updates
- `/health` - Health check endpoint (returns "OK")

### Example Webhook URLs
- `https://yourdomain.com/webhook`
- `https://your-app.herokuapp.com/webhook`
- `https://your-app.railway.app/webhook`

### Switching Between Modes
The bot automatically handles switching between polling and webhook modes:
- **Polling → Webhook**: Removes existing webhook, sets up new one
- **Webhook → Polling**: Removes webhook, starts polling

### Benefits of Webhook Mode
- **Lower Latency** - Instant message delivery vs polling delay
- **Better Performance** - No constant polling requests
- **Resource Efficient** - Less CPU and network usage
- **Scalable** - Better for high-traffic bots

## Key Features Implementation

### Database Integration
- Automatic schema creation
- Conflict resolution for user updates
