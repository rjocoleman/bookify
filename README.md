# Bookify - EPUB to KEPUB Converter

Bookify is a web-based service that automatically converts EPUB files to KEPUB format (optimized for Kobo e-readers) and uploads them to Google Drive. It features a drag-and-drop interface, background processing, and real-time status updates.

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
- Google Cloud Console account with a service account
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

### 2. Google Drive Service Account Setup

See [SERVICE_ACCOUNT_SETUP.md](SERVICE_ACCOUNT_SETUP.md) for detailed instructions.

**Quick Setup:**

1. Create a service account in Google Cloud Console
2. Download the JSON key file
3. Enable Google Drive API
4. Share your Google Drive folder with the service account email
5. Set the environment variable with the key file path or content

### 3. Environment Variables

Set one of these for authentication:

```bash
# Option 1: Path to service account key file (recommended)
export GOOGLE_SERVICE_ACCOUNT_KEY_PATH="/path/to/service-account-key.json"

# Option 2: Service account key JSON content
export GOOGLE_SERVICE_ACCOUNT_KEY='{"type":"service_account",...}'
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
# Set service account key path
export GOOGLE_SERVICE_ACCOUNT_KEY_PATH="/path/to/service-account-key.json"

# Run the application
./bin/bookify
```

Navigate to `http://localhost:8080`

### Docker

```bash
# Build the Docker image
docker build -t bookify .

# Run with Docker (using key file)
docker run -d \
  -p 8080:8080 \
  -v /path/to/service-account-key.json:/app/key.json:ro \
  -e GOOGLE_SERVICE_ACCOUNT_KEY_PATH="/app/key.json" \
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
3. Click "Create Account"

**Note**: Make sure you've shared the Google Drive folder with your service account email before setting up.

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
| `GOOGLE_SERVICE_ACCOUNT_KEY_PATH` | Path to service account JSON key file | - |
| `GOOGLE_SERVICE_ACCOUNT_KEY` | Service account JSON key content | - |
| `PORT` | Server port | 8080 |
| `DB_PATH` | SQLite database path | ./kepub.db |
| `TEMP_DIR` | Temporary file directory | ./temp |
| `MAX_FILE_SIZE` | Maximum upload size | 100MB |

**Note**: You must set either `GOOGLE_SERVICE_ACCOUNT_KEY_PATH` or `GOOGLE_SERVICE_ACCOUNT_KEY`.

## Troubleshooting

### "Cannot access folder"

The service account doesn't have access to the folder.

**Solution**:
1. Find your service account email in the JSON key file (look for `client_email`)
2. Share the Google Drive folder with this email address
3. Make sure the service account has "Editor" permissions

### "No service account key configured"

Bookify can't find the service account credentials.

**Solution**: Make sure you've set either:
- `GOOGLE_SERVICE_ACCOUNT_KEY_PATH` with a valid file path, or
- `GOOGLE_SERVICE_ACCOUNT_KEY` with the JSON content

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
