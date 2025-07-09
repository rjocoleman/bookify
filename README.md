# Bookify - EPUB to KEPUB Converter

Bookify is a web-based service that automatically converts EPUB files to KEPUB format (optimized for Kobo e-readers) and uploads them to Google Drive using OAuth authentication. It features a drag-and-drop interface, background processing, and real-time status updates.

## Features

- Convert EPUB files to KEPUB format using the kepubify library
- Automatic upload to Google Drive folders
- Background job processing with real-time status updates
- Multi-account support
- Drag-and-drop file uploads
- Automatic temporary file cleanup
- Fast, lightweight Go backend with HTMX frontend

## Prerequisites

- Go 1.21 or higher
- A Google account
- Access to Google Cloud Console for OAuth setup
- Google Drive API enabled

## Installation

### 1a. Docker

```bash
docker pull ghcr.io/rjocoleman/bookify:v1.0.0
```

### 1b. Clone and Build

```bash
git clone https://github.com/rjocoleman/bookify.git
cd bookify

# Install dependencies
go mod download

# Install templ CLI tool
go install github.com/a-h/templ/cmd/templ@latest

# Generate templates
templ generate

# Build the application
go build -o bin/bookify ./cmd/server
```

### 2. Google Drive OAuth Setup

See [OAUTH_SETUP.md](OAUTH_SETUP.md) for detailed instructions.

**Quick Setup:**

1. Create OAuth 2.0 credentials in Google Cloud Console
2. Set up the OAuth consent screen
3. Add `http://localhost:8080/oauth/callback` as redirect URI
4. Set the Client ID and Client Secret as environment variables
5. Authorize Bookify through the web interface

### 3. Environment Variables

Set these for OAuth authentication:

```bash
# Required: OAuth credentials from Google Cloud Console
export GOOGLE_CLIENT_ID="your-client-id.apps.googleusercontent.com"
export GOOGLE_CLIENT_SECRET="your-client-secret"
```

Optional configuration:

```bash
export PORT=8080                    # Optional, defaults to 8080
export DB_PATH="./bookify.db"       # Optional, defaults to ./kepub.db
export TEMP_DIR="./temp"            # Optional, defaults to ./temp
```

## Running the Application

### Local Development

```bash
# Set OAuth credentials
export GOOGLE_CLIENT_ID="your-client-id.apps.googleusercontent.com"
export GOOGLE_CLIENT_SECRET="your-client-secret"

# Run the application
./bin/bookify
```

Navigate to `http://localhost:8080`

### Docker

```bash
# Build the Docker image
docker build -t bookify .

# Run with Docker
docker run -d \
  -p 8080:8080 \
  -e GOOGLE_CLIENT_ID="your-client-id.apps.googleusercontent.com" \
  -e GOOGLE_CLIENT_SECRET="your-client-secret" \
  -v $(pwd)/data:/root \
  --name bookify \
  bookify
```

## Usage

### First Time Setup

1. When you first access Bookify, you'll be redirected to the setup page
2. Enter:
   - **Account Name**: A friendly name for this Google Drive account (e.g., "My Kobo Library")
   - **Folder ID**: The Google Drive folder ID where files will be uploaded
3. Click "Authorize with Google"
4. Sign in with your Google account and grant Bookify permission to upload files
5. You'll be redirected back to Bookify once authorized

**Note**: Files will be uploaded to your personal Google Drive storage quota.

### Uploading Books

1. Select an account from the dropdown
2. Drag and drop EPUB files onto the upload area, or click to browse
3. Files will be:
   - Validated (must be valid EPUB format)
   - Queued for processing
   - Converted to KEPUB format
   - Uploaded to your Google Drive folder
4. Monitor progress in real-time in the processing queue

### API Endpoints

- `GET /` - Main page (redirects to setup if no accounts)
- `GET /setup` - Account setup page
- `POST /setup` - Create account
- `POST /upload` - Upload EPUB files
- `GET /api/queue` - Get queue status (JSON)
- `GET /api/job/:id` - Get specific job status (JSON)

## Configuration

All configuration is done through environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `GOOGLE_CLIENT_ID` | OAuth 2.0 Client ID | - |
| `GOOGLE_CLIENT_SECRET` | OAuth 2.0 Client Secret | - |
| `PORT` | Server port | 8080 |
| `DB_PATH` | SQLite database path | ./kepub.db |
| `TEMP_DIR` | Temporary file directory | ./temp |
| `MAX_FILE_SIZE` | Maximum upload size | 100MB |

**Note**: You must set both `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET`.

## Troubleshooting

### "Folder access failed"

Bookify cannot access the specified Google Drive folder.

**Solution**:
1. Verify the folder ID is correct (from the folder URL)
2. Make sure the folder exists in your Google Drive
3. Ensure you're authorizing with the correct Google account

### "OAuth configuration missing"

Bookify can't find the OAuth credentials.

**Solution**: Make sure you've set both:
- `GOOGLE_CLIENT_ID` with your OAuth client ID
- `GOOGLE_CLIENT_SECRET` with your OAuth client secret

### Token Expiration

OAuth tokens may expire after extended periods.

**Solution**: Return to the setup page and re-authorize with Google

### Upload failures

Check that:
- Files are valid EPUB format (not PDF or other formats)
- Files aren't corrupted
- You have write permissions to the Google Drive folder

## Development

### Project Structure

```
bookify/
   cmd/server/         # Application entry point
   internal/
      db/            # Database models and service
      handlers/      # HTTP request handlers
      services/      # Business logic
      templates/     # Templ HTML templates
   web/static/        # Static assets
   temp/              # Temporary file storage
```

### Building from Source

```bash
# Install just (command runner)
# macOS: brew install just
# Other platforms: see https://github.com/casey/just

# Install dependencies and dev tools
just install

# Generate templates
just generate

# Run tests
just test

# Build
just build
```

### Development Commands

```bash
just                      # Show available commands
just dev                  # Generate, build, and run server
just test                 # Run all tests
just test-verbose         # Run tests with verbose output
just coverage             # Generate HTML coverage report
just coverage-show        # Show coverage in terminal
just check                # Run fmt, lint, and tests
just clean                # Clean build artifacts
```

## License

MIT License - see LICENSE file for details

## Acknowledgments

- [kepubify](https://github.com/pgaskin/kepubify) - EPUB to KEPUB conversion
- [Echo](https://echo.labstack.com/) - Web framework
- [HTMX](https://htmx.org/) - Frontend interactivity
- [Templ](https://templ.guide/) - HTML templating
- [GORM](https://gorm.io/) - Database ORM
