# Workflow Dashboard - Enhanced Implementation

## Overview

The urunc Workflow Dashboard has been redesigned to match the kata-containers style, providing a comprehensive view of CI/CD workflow status and job history.

## Features

✨ **Key Features:**

- **Workflow Selector Dropdown** - Switch between different workflows to view their job details
- **Job Search Filter** - Search and filter jobs by name in real-time
- **Pass Rate Dashboard** - View success rate percentage for each job
- **Run History Visualization** - Last 10 runs displayed as color-coded status badges:
  - 🟢 **Green** - Pass/Success
  - 🔴 **Red** - Fail/Failure  
  - 🟠 **Orange** - Skipped
  - 🔵 **Blue** - In Progress

- **Real-time Data** - Fetches directly from GitHub Actions API
- **Responsive Design** - Works on mobile, tablet, and desktop
- **Dark Mode Support** - Compatible with Material MkDocs theme

## Technical Details

### How It Works

1. **Workflow Discovery**
   - Fetches all active workflows from GitHub API
   - Populates dropdown selector with workflow names

2. **Job Aggregation**
   - When a workflow is selected, fetches the last 10 workflow runs
   - For each run, retrieves all jobs and their statuses
   - Aggregates job data across all runs

3. **Pass Rate Calculation**
   - Counts successful runs (conclusion: 'success')
   - Calculates percentage: (passed / total) × 100
   - Displays jobs sorted by lowest pass rate first

4. **Search Filtering**
   - Client-side filtering of job names
   - Real-time as user types
   - Case-insensitive matching

### Components

**File:** `docs/workflows.md`

Contains:
- HTML structure with dropdown and search controls
- CSS styling for dashboard layout and responsive design
- JavaScript for API interaction and dynamic rendering

**Dependencies:**
- GitHub Actions REST API v3
- No external libraries required (vanilla JavaScript)

## Usage

### Accessing the Dashboard

1. Navigate to the urunc documentation site
2. Click "CI Workflows" in the sidebar navigation
3. The dashboard loads automatically at `https://urunc.io/workflows`

### Using the Dashboard

**Select a Workflow:**
- Use the "Select Workflow" dropdown to choose which workflow to analyze
- Dashboard updates automatically with that workflow's job information

**Search Jobs:**
- Type a job name in the "Search Jobs" field
- Table filters in real-time as you type
- Clear the search to see all jobs again

**View Job Details:**
- Job names are clickable links to GitHub Actions
- Pass rate shows success percentage
- Colored boxes show the status of the last 10 runs

## Performance Considerations

### API Calls

The dashboard makes the following API calls:

1. **Fetch Workflows** (1 call)
   ```
   GET /repos/urunc-dev/urunc/actions/workflows
   ```

2. **Fetch Workflow Runs** (1 call per selected workflow)
   ```
   GET /repos/urunc-dev/urunc/actions/workflows/{id}/runs?per_page=10
   ```

3. **Fetch Job Details** (1 call per run, ~10 calls per workflow)
   ```
   GET /repos/urunc-dev/urunc/actions/runs/{id}/jobs
   ```

**Total:** ~12 API calls per workflow selection

### Rate Limiting

GitHub API limits:
- **Unauthenticated:** 60 requests/hour
- **Authenticated:** 5,000 requests/hour

The dashboard should work fine within these limits. Typical usage:
- Initial page load: ~1-2 API calls
- Per workflow selection: ~11-12 API calls

## Customization

### For Other Organizations/Repositories

Update the constants in the JavaScript:

```javascript
const owner = 'urunc-dev';    // Change to your owner
const repo = 'urunc';          // Change to your repo
```

### Styling

All styles are embedded in `docs/workflows.md`. Key customizable classes:

- `.dashboard-container` - Main container
- `.controls-section` - Dropdown and search controls
- `.jobs-table` - Main table styling
- `.status-badge` - Run status indicators

### Authentication

For higher rate limits or private repositories, add authentication headers:

```javascript
const response = await fetch(url, {
  headers: {
    'Accept': 'application/vnd.github.v3+json',
    'Authorization': 'token YOUR_GITHUB_TOKEN'  // Add this
  }
});
```

## Troubleshooting

### Dashboard Not Loading

1. **Check browser console** (F12 → Console tab)
2. **Verify GitHub API accessibility** - try API URL directly
3. **Rate limit hit** - wait 1 hour and reload (GitHub API rate limit resets hourly)
4. **Corporate firewall/proxy** - may block GitHub API

### Empty Results

- Ensure the selected workflow has runs in the last activity
- Some workflows may not have recent executions
- Try a different workflow from the dropdown

### Missing Jobs

- Jobs may be conditionally skipped in workflow
- Run succeeded too quickly without job output
- Check the workflow definition for conditional job steps

## Future Enhancements

Potential improvements:

- [ ] Historical trends and metrics
- [ ] Job duration tracking
- [ ] Run timing analysis
- [ ] Failure rate alerts
- [ ] Custom date range selection  
- [ ] Export data as CSV/JSON
- [ ] Integration with other monitoring tools
- [ ] Caching for better performance
- [ ] Authentication for private repos
- [ ] Compare pass rates between workflows
