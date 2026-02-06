// Workflow Dashboard - Cache First with Optional Live Refresh
// This implementation prioritizes cached data for fast loads
// Users can optionally refresh to get live GitHub API data

const owner = 'urunc-dev';
const repo = 'urunc';
let cachedData = null;
let allWorkflows = [];

async function loadWorkflows() {
  const container = document.getElementById('jobs-container');
  
  try {
    // Step 1: Load from cached JSON first (fast, no API calls)
    console.log('Loading workflows from cache...');
    const cacheResponse = await fetch('../data/workflows.json');
    
    if (!cacheResponse.ok) {
      throw new Error('Cache file not available. Please try the live API.');
    }
    
    const cacheData = await cacheResponse.json();
    cachedData = cacheData;
    
    // Extract workflows from cache
    const allWorkflowsList = [];
    if (cacheData.categories) {
      Object.keys(cacheData.categories).forEach(name => {
        const workflows = cacheData.categories[name];
        if (workflows.length > 0) {
          allWorkflowsList.push({
            name: name,
            ...workflows[0]
          });
        }
      });
    }
    
    allWorkflows = allWorkflowsList.filter(w => w.state === 'active');
    
    if (allWorkflows.length === 0) {
      container.innerHTML = '<div class="no-data">No active workflows found in cache.</div>';
      return;
    }
    
    // Build workflow dropdown
    const select = document.getElementById('workflow-select');
    select.innerHTML = allWorkflows.map((w, idx) => 
      `<option value="${idx}">${escapeHtml(w.name || w.path)}</option>`
    ).join('');
    
    // Show from cache and auto-load job data
    displayWorkflowInfo(0);
    fetchLiveData(allWorkflows[0]);
    
    select.addEventListener('change', (e) => {
      const idx = parseInt(e.target.value);
      displayWorkflowInfo(idx);
      fetchLiveData(allWorkflows[idx]);
    });
    
  } catch (error) {
    console.error('Error:', error);
    container.innerHTML = `<div class="error-message">
      <strong>⚠️ Failed to load workflow cache</strong>
      <p>${escapeHtml(error.message)}</p>
      <p style="font-size: 0.9em; margin-top: 10px;">Try refreshing the page or check if the cache file exists.</p>
    </div>`;
  }
}

function displayWorkflowInfo(workflowIndex) {
  const container = document.getElementById('jobs-container');
  const workflow = allWorkflows[workflowIndex];
  
  if (!workflow) return;
  
  const lastUpdated = cachedData?.generated_at 
    ? new Date(cachedData.generated_at).toLocaleString() 
    : 'Unknown';
  
  let html = `
    <div class="info-message">
      <strong>${escapeHtml(workflow.name || workflow.path)}</strong>
      <div style="font-size: 0.9em; margin-top: 8px; color: #555;">
        📅 <strong>Cache updated:</strong> ${lastUpdated}
      </div>
    </div>
    <div class="loading"><p>⏳ Loading job run history...</p></div>
  `;
  
  container.innerHTML = html;
}

async function fetchLiveData(workflow) {
  const container = document.getElementById('jobs-container');
  
  try {
    const runsResponse = await fetch(
      `https://api.github.com/repos/${owner}/${repo}/actions/workflows/${workflow.id}/runs?per_page=10`,
      { headers: { 'Accept': 'application/vnd.github.v3+json' } }
    );
    
    if (!runsResponse.ok) {
      if (runsResponse.status === 403) {
        throw new Error('GitHub API rate limit exceeded (403). You\'ve made too many requests. Wait an hour and try again, or add GitHub token authentication.');
      }
      throw new Error(`GitHub API error: ${runsResponse.status}`);
    }
    
    const runsData = await runsResponse.json();
    const runs = runsData.workflow_runs || [];
    
    // Collect job data
    const jobMap = new Map();
    
    for (const run of runs) {
      const jobsResponse = await fetch(run.jobs_url, {
        headers: { 'Accept': 'application/vnd.github.v3+json' }
      });
      
      if (!jobsResponse.ok) continue;
      
      const jobsData = await jobsResponse.json();
      const jobs = jobsData.jobs || [];
      
      for (const job of jobs) {
        if (!jobMap.has(job.name)) {
          jobMap.set(job.name, { name: job.name, runs: [] });
        }
        jobMap.get(job.name).runs.push({
          status: job.conclusion || job.status,
          runId: run.id,
          htmlUrl: job.html_url
        });
      }
    }
    
    // Calculate pass rates
    const jobsData = Array.from(jobMap.values()).map(job => {
      const recentRuns = job.runs.slice(0, 10);
      const passed = recentRuns.filter(r => r.status === 'success').length;
      const total = recentRuns.length;
      const passRate = total > 0 ? ((passed / total) * 100).toFixed(2) : 0;
      
      return {
        name: job.name,
        passRate: parseFloat(passRate),
        runs: recentRuns
      };
    });
    
    jobsData.sort((a, b) => a.passRate - b.passRate);
    renderJobsTable(jobsData, workflow);
    
  } catch (error) {
    console.error('Error fetching live data:', error);
    container.innerHTML = `<div class="error-message">
      <strong>⚠️ Failed to load live job data</strong>
      <p>${escapeHtml(error.message)}</p>
      <div style="margin-top: 10px; padding: 10px; background: #fff3cd; border: 1px solid #ffc107; border-radius: 4px; color: #856404;">
        <strong>GitHub API Rate Limiting:</strong>
        <ul style="margin: 8px 0 0 20px; padding: 0;">
          <li>Unauthenticated requests: 60 per hour</li>
          <li>Each workflow loads ~10-12 requests</li>
          <li>Limit resets every hour</li>
          <li>Use cached data above or wait an hour</li>
        </ul>
      </div>
    </div>`;
  }
}

function renderJobsTable(jobsData, workflow) {
  const container = document.getElementById('jobs-container');
  
  if (jobsData.length === 0) {
    container.innerHTML = '<div class="no-data">No job data available for this workflow.</div>';
    return;
  }
  
  const lastUpdated = cachedData?.generated_at 
    ? new Date(cachedData.generated_at).toLocaleString() 
    : 'Unknown';
  
  let html = `
    <div class="info-message">
      <strong>${escapeHtml(workflow.name || workflow.path)}</strong>
      <div style="font-size: 0.9em; margin-top: 8px; color: #555;">
        <span style="margin-right: 15px;">📅 <strong>Cache:</strong> ${lastUpdated}</span>
        <span>📊 <strong>Jobs:</strong> ${jobsData.length}</span>
      </div>
    </div>
  `;
  
  html += `
    <table class="jobs-table">
      <thead>
        <tr>
          <th style="width: 35%;">Job Name</th>
          <th style="width: 12%; text-align: center;">Pass Rate</th>
          <th style="width: 53%;">Last 10 Runs (Most Recent → Oldest)</th>
        </tr>
      </thead>
      <tbody>
  `;
  
  jobsData.forEach(job => {
    const runsHtml = job.runs.map(run => {
      let text, className, emoji;
      switch (run.status) {
        case 'success':
          text = '✓';
          className = 'status-pass';
          emoji = '✓';
          break;
        case 'failure':
          text = '✗';
          className = 'status-fail';
          emoji = '✗';
          break;
        case 'skipped':
          text = '○';
          className = 'status-skip';
          emoji = '○';
          break;
        default:
          text = '●';
          className = 'status-in-progress';
          emoji = '●';
      }
      return `<span class="status-badge ${className}" title="${run.status}">${emoji}</span>`;
    }).join(' ');
    
    html += `
      <tr>
        <td><a href="https://github.com/${owner}/${repo}/actions/workflows/${escapeHtml(workflow.path)}" target="_blank" class="job-name">${escapeHtml(job.name)}</a></td>
        <td class="pass-rate"><span class="pass-rate-badge" style="background-color: ${getPassRateColor(job.passRate)}">${job.passRate.toFixed(1)}%</span></td>
        <td><div class="run-status">${runsHtml}</div></td>
      </tr>
    `;
  });
  
  html += `
      </tbody>
    </table>
  `;
  
  container.innerHTML = html;
}

function getPassRateColor(passRate) {
  if (passRate >= 90) return '#28a745'; // green
  if (passRate >= 70) return '#ffc107'; // yellow
  if (passRate >= 50) return '#fd7e14'; // orange
  return '#dc3545'; // red
}

function escapeHtml(text) {
  const map = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#039;'
  };
  return String(text).replace(/[&<>"']/g, m => map[m]);
}

// Load on page ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', loadWorkflows);
} else {
  loadWorkflows();
}
