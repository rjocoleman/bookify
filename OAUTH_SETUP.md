# OAuth Setup for Bookify

This guide explains how to set up OAuth authentication for Bookify. This change was made to address Google's storage quota limitations with service accounts - by using OAuth, files are uploaded to your personal Google Drive storage quota instead of the service account's limited quota.

## Prerequisites

- A Google account
- Access to Google Cloud Console
- Bookify running on localhost (http://localhost:8080)

## Step 1: Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click "Select a project" → "New Project"
3. Enter a project name (e.g., "Bookify")
4. Click "Create"

## Step 2: Enable Google Drive API

1. In your Google Cloud project, go to "APIs & Services" → "Library"
2. Search for "Google Drive API"
3. Click on it and then click "Enable"

## Step 3: Create OAuth 2.0 Credentials

1. Go to "APIs & Services" → "Credentials"
2. Click "Create Credentials" → "OAuth client ID"
3. If prompted, configure the OAuth consent screen:
   - Choose "External" for user type
   - Fill in the required fields:
     - App name: "Bookify"
     - User support email: Your email
     - Developer contact: Your email
   - Add scopes: Click "Add or Remove Scopes" and add: `https://www.googleapis.com/auth/drive`
   - Add your email as a test user (if in testing mode)
4. For Application type, select "Web application"
5. Name it "Bookify OAuth"
6. Add authorized redirect URI: `http://localhost:8080/oauth/callback`
7. Click "Create"
8. Copy the Client ID and Client Secret

## Step 4: Configure Bookify

1. Create a `.env` file in your Bookify directory (if it doesn't exist)
2. Add the OAuth credentials:
   ```
   GOOGLE_CLIENT_ID=your_client_id_here.apps.googleusercontent.com
   GOOGLE_CLIENT_SECRET=your_client_secret_here
   ```

## Step 5: Set Up Your Google Drive Folder

1. Go to [Google Drive](https://drive.google.com)
2. Create a new folder for your EPUB files (or use an existing one)
3. Open the folder and copy the folder ID from the URL:
   - URL format: `https://drive.google.com/drive/folders/[FOLDER_ID]`
   - Copy only the `[FOLDER_ID]` part

## Step 6: Create Your Bookify Account

1. Start Bookify: `go run cmd/server/main.go`
2. Navigate to http://localhost:8080
3. You'll be redirected to the setup page
4. Enter:
   - Account Name: A friendly name for this account (e.g., "Personal Library")
   - Google Drive Folder ID: The folder ID you copied in Step 5
5. Click "Authorize with Google"
6. You'll be redirected to Google's OAuth consent page
7. Sign in with your Google account
8. Grant Bookify permission to upload files to your Google Drive
9. You'll be redirected back to Bookify

## Troubleshooting

### "OAuth denied" error
- Make sure you've added your email as a test user in the OAuth consent screen configuration
- Ensure you're signing in with the correct Google account

### "Folder access failed" error
- Verify the folder ID is correct
- Make sure the folder exists in your Google Drive
- Try creating a new folder and using its ID

### "Token exchange failed" error
- Check that your Client ID and Client Secret are correctly set in the `.env` file
- Ensure the redirect URI in Google Cloud Console exactly matches `http://localhost:8080/oauth/callback`

### Token expiration
- OAuth tokens expire after a period of time
- If uploads start failing, return to the setup page to re-authenticate
- Bookify will attempt to refresh tokens automatically when possible

## Migration from Service Account

If you were previously using a service account:

1. Your existing uploaded files remain in the service account's Drive
2. You'll need to set up OAuth as described above
3. New uploads will go to your personal Google Drive
4. Consider manually transferring important files from the service account's Drive to your personal Drive

## Privacy & Security

- Your Google credentials are stored locally in Bookify's database
- Bookify requests full Google Drive access to be able to upload files to any folder you specify
- You can revoke access at any time from your [Google Account settings](https://myaccount.google.com/permissions)
- Only use Bookify on trusted systems as it has access to your Google Drive

## For Developers

The OAuth flow uses the standard Google OAuth 2.0 implementation:
- Authorization endpoint: `https://accounts.google.com/o/oauth2/v2/auth`
- Token endpoint: `https://oauth2.googleapis.com/token`
- Scopes: `https://www.googleapis.com/auth/drive`

The refresh token is stored to enable automatic token renewal without requiring re-authentication.