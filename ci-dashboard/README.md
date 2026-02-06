# urunc CI Dashboard

A web dashboard that provides a unified view of CI workflow statuses for the
urunc project. Inspired by the
[Kata Containers CI Dashboard](https://kata-containers.github.io/).

## Features

- Fetches workflow data from the GitHub Actions API
- Categorizes workflows by type (CI/Testing, Build/Deploy, Code Quality, etc.)
- Displays weather-style icons reflecting workflow health
- Search and filter workflows by name or category
- Expandable rows showing per-job details
- Works with both pre-fetched data and live API calls

## Quick Start

### Using Pre-fetched Data (recommended)

1. Set a GitHub token:
   ```bash
   export TOKEN=<your_github_pat>
   ```

2. Fetch the data:
   ```bash
   mkdir -p data
   node scripts/fetch-ci-data.js > data/workflow_stats.json
   ```

3. Serve the dashboard:
   ```bash
   python3 -m http.server 8000
   # Open http://localhost:8000
   ```

### Using Live API

Simply open `index.html` in a browser. Note that unauthenticated GitHub API
requests are rate-limited to 60 requests per hour.

## Project Structure

```
ci-dashboard/
├── index.html                   # Main dashboard page
├── README.md                    # This file
├── public/
│   ├── sunny.svg                # Weather icon: all passing
│   ├── partially-sunny.svg      # Weather icon: mostly passing
│   ├── cloudy.svg               # Weather icon: some failures
│   ├── rainy.svg                # Weather icon: many failures
│   └── stormy.svg               # Weather icon: mostly failing
└── scripts/
    └── fetch-ci-data.js         # Script to pre-fetch workflow data
```

## Deployment

The dashboard is deployed automatically via GitHub Actions. The
`ci-dashboard-deploy.yml` workflow:

1. Runs `fetch-ci-data.js` to gather the latest workflow data
2. Commits the data to the `latest-ci-dashboard-data` branch
3. Deploys the dashboard to GitHub Pages

The data is refreshed daily via a scheduled workflow run.
