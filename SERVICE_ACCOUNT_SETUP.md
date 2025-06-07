# Simple Google Service Account Setup for Bookify

This guide will help you set up Google Drive access for Bookify using a service account. This is much simpler than OAuth and perfect for personal use.

## Why Service Accounts?

- **No OAuth complexity** - No refresh tokens or OAuth playground
- **Set once, forget** - Keys don't expire
- **Perfect for personal use** - Ideal for 2-3 users
- **Direct authentication** - Just use a JSON key file

## Step-by-Step Setup

### 1. Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click "Select a project" → "New Project"
3. Name it something like "Bookify" and create it
4. Make sure your new project is selected in the top dropdown

### 2. Enable Google Drive API

1. In the sidebar, go to "APIs & Services" → "Library"
2. Search for "Google Drive API"
3. Click on it and press "ENABLE"

### 3. Create Service Account

1. Go to "APIs & Services" → "Credentials"
2. Click "+ CREATE CREDENTIALS" → "Service account"
3. Fill in:
   - Service account name: `bookify-service`
   - Service account ID: (auto-fills)
   - Description: `Bookify Drive access`
4. Click "CREATE AND CONTINUE"
5. Skip the optional steps (click "DONE")

### 4. Download the Key

1. Click on your new service account email (bookify-service@...)
2. Go to the "KEYS" tab
3. Click "ADD KEY" → "Create new key"
4. Choose "JSON" format
5. Click "CREATE" - a file will download
6. **Save this file securely!** You'll need it to run Bookify

### 5. Get the Service Account Email

Open the downloaded JSON file and find the `client_email` field. It looks like:
```
"client_email": "bookify-service@your-project.iam.gserviceaccount.com"
```

Copy this email address.

### 6. Share Your Google Drive Folder

1. Go to [Google Drive](https://drive.google.com)
2. Create a new folder called "Bookify" (or use existing)
3. Right-click the folder → "Share"
4. Paste the service account email from step 5
5. Make sure it has "Editor" permission
6. Click "Send"

### 7. Get the Folder ID

1. Open the folder in Google Drive
2. Look at the URL in your browser
3. Copy the ID from: `https://drive.google.com/drive/folders/YOUR_FOLDER_ID_HERE`

## Using with Bookify

### Option 1: Environment Variable (Recommended)

```bash
export SERVICE_ACCOUNT_KEY_PATH="/path/to/your-downloaded-key.json"
./bin/bookify
```

### Option 2: .env File

Create a `.env` file:
```bash
SERVICE_ACCOUNT_KEY_PATH=/path/to/your-downloaded-key.json
```

### Option 3: Docker

```bash
docker run -d \
  -p 8080:8080 \
  -v /path/to/key.json:/app/key.json:ro \
  -e SERVICE_ACCOUNT_KEY_PATH="/app/key.json" \
  -v $(pwd)/data:/root \
  bookify
```

## Setting Up Accounts in Bookify

1. Start Bookify and go to `http://localhost:8080`
2. You'll be redirected to setup
3. Enter:
   - **Account Name**: e.g., "My Kobo Books"
   - **Folder ID**: The ID from step 7
4. Click "Create Account"

That's it! You can now drag and drop EPUB files and they'll be converted and uploaded to your Google Drive.

## Multiple Users

For your partner:
1. Create another folder in Google Drive
2. Share it with the same service account email
3. Add it as a second account in Bookify with a different name

## Troubleshooting

### "Permission denied" errors
- Make sure the folder is shared with the service account email
- Check that the service account has "Editor" permission

### "Invalid credentials"
- Verify the key file path is correct
- Check that the JSON file is valid and complete

### Can't see uploaded files
- Files are uploaded by the service account, not your personal account
- They should still appear in the shared folder
- Check the folder permissions

## Security Notes

- Keep your JSON key file secure
- Don't commit it to version control
- The service account only has access to folders you explicitly share with it
- Perfect for personal/family use where you trust all users
