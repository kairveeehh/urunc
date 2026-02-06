# CI Workflows Viewer Documentation

## Overview

The CI Workflows Viewer provides a unified, centralized view of all GitHub Actions workflows configured for the urunc project. This feature fetches workflow information directly from the GitHub API and displays it in an organized, categorized manner on the project's documentation site.

## Features

- **Live Updates**: Fetches workflow data directly from GitHub API (primary method)
- **Offline Support**: Falls back to cached JSON data if API is unavailable
- **Categorized Display**: Workflows are automatically grouped by their configured name
- **Status Indicators**: Shows whether each workflow is active or disabled
- **Direct Links**: Quick links to view each workflow in GitHub Actions
- **Metadata**: Displays workflow path, ID, creation date, and last update date
- **Automatic Updates**: Weekly scheduled job to keep the cached data fresh

## Implementation Details

### Components

1. **Documentation Page** (`docs/workflows.md`)
   - Interactive HTML page served as part of MkDocs documentation
   - Client-side JavaScript that fetches and displays workflows
   - Responsive grid layout for workflow cards

2. **Python Script** (`script/workflows.py`)
   - Command-line tool to fetch and display workflows locally
   - Can export workflow data to JSON format
   - Useful for debugging and manual updates

3. **GitHub Actions Workflow** (`.github/workflows/update-workflows.yml`)
   - Runs weekly to keep cached data fresh
   - Can be manually triggered
   - Exports workflow data to `docs/assets/workflows.json`

### Data Flow

```
GitHub API (Primary)
    ↓
Browser/Page loads
    ↓
Parse and display
    ↓
If API unavailable → Fallback to JSON cache
```

## Usage

### View Workflows

Simply navigate to the "CI Workflows" page in the urunc documentation at:
```
https://urunc.io/workflows/
```

### Update Workflows Cache Manually

#### Option 1: Use the Python Script

```bash
# Display all workflows
python3 script/workflows.py

# Export to JSON file
python3 script/workflows.py --export workflows.json

# Specify different repository
python3 script/workflows.py --owner org-name --repo repo-name
```

#### Option 2: Run GitHub Actions Workflow

Manually trigger the "Update CI Workflows List" workflow from the GitHub Actions tab.

### Integration with CI/CD

The `update-workflows.yml` workflow:
- Runs automatically every Monday at 00:00 UTC
- Can be manually triggered via GitHub Actions UI
- Automatically commits updated `docs/assets/workflows.json` to main branch

## Customization

### Adding to Other Repositories

To add this feature to another repository:

1. Copy `docs/workflows.md` to your docs folder
2. Update repository references in the JavaScript:
   ```javascript
   const owner = 'your-owner';
   const repo = 'your-repo';
   ```
3. Copy `.github/workflows/update-workflows.yml` if you want automated updates
4. Update your navigation config to include the workflows page

### Styling

The workflows page uses the Material MkDocs theme and custom CSS. You can customize the appearance by modifying the `<style>` section in `docs/workflows.md`.

### API Rate Limiting

GitHub API allows 60 requests per hour for unauthenticated requests, and 5,000 per hour with authentication. The workflows page makes one API call per page load. If you hit rate limits:
- The page will automatically fall back to cached data
- Increase the cache update frequency
- Use GitHub token authentication (requires workflow updates)

## Troubleshooting

### Workflows Not Loading

1. **Check browser console**: Open DevTools (F12) to see error messages
2. **Verify GitHub API accessibility**: Try the API URL directly in browser
3. **Check cache file**: Verify `docs/assets/workflows.json` exists and is valid
4. **API rate limit hit**: Wait an hour, then reload (will use cached data)

### Missing Workflows

If workflows don't appear after adding them:
1. Ensure workflow file is in `.github/workflows/` directory
2. Wait for the automatic weekly update or manually run the Python script
3. Clear browser cache and reload

## Future Enhancements

Potential improvements:

- [ ] Real-time workflow run status and recent execution history
- [ ] Filter workflows by status, trigger type, or custom tags
- [ ] Search functionality for large workflow lists
- [ ] Workflow metrics and statistics
- [ ] Integration with other CI systems (not just GitHub Actions)
- [ ] Advanced categorization options (by trigger type, purpose, etc.)
- [ ] Workflow dependency visualization
- [ ] Performance metrics and execution time tracking

## API Reference

### GitHub Workflows API Endpoint

```
GET /repos/{owner}/{repo}/actions/workflows
```

Response includes:
- `id`: Workflow ID
- `name`: Workflow display name
- `path`: Relative path to workflow file
- `state`: Current state (active/disabled)
- `created_at`: Creation timestamp
- `updated_at`: Last update timestamp

## Contributing

When adding or modifying CI workflows:

1. Ensure workflow files follow naming conventions
2. Update workflow names/descriptions to be meaningful
3. Test the workflows page after changes
4. Consider the frequency of updates for performance
