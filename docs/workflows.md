# Workflow Dashboard

View the status and history of all CI/CD workflows and their jobs.

<style>
  .dashboard-container {
    margin-top: 20px;
  }

  .controls-section {
    display: flex;
    gap: 20px;
    margin-bottom: 20px;
    flex-wrap: wrap;
  }

  .control-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 250px;
  }

  .control-group label {
    font-weight: 600;
    font-size: 0.95em;
  }

  .control-group select,
  .control-group input {
    padding: 8px 12px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 0.95em;
    font-family: inherit;
  }

  .control-group select:focus,
  .control-group input:focus {
    outline: none;
    border-color: #007bff;
    box-shadow: 0 0 0 3px rgba(0, 123, 255, 0.1);
  }

  .jobs-table {
    width: 100%;
    border-collapse: collapse;
    margin-top: 20px;
    background: white;
    border: 1px solid #ddd;
    border-radius: 4px;
    overflow: hidden;
  }

  .jobs-table thead {
    background-color: #f5f5f5;
  }

  .jobs-table th {
    padding: 12px;
    text-align: left;
    font-weight: 600;
    border-bottom: 2px solid #ddd;
    white-space: nowrap;
  }

  .jobs-table td {
    padding: 12px;
    border-bottom: 1px solid #eee;
  }

  .jobs-table tbody tr:hover {
    background-color: #f9f9f9;
  }

  .job-name {
    color: #007bff;
    text-decoration: none;
    font-weight: 500;
  }

  .job-name:hover {
    text-decoration: underline;
  }

  .pass-rate {
    font-weight: 600;
    text-align: center;
  }

  .run-status {
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
  }

  .status-badge {
    display: inline-block;
    width: 20px;
    height: 20px;
    border-radius: 3px;
    font-size: 0.7em;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-weight: bold;
    cursor: help;
  }

  .status-pass {
    background-color: #28a745;
  }

  .status-fail {
    background-color: #dc3545;
  }

  .status-skip {
    background-color: #ffc107;
    color: #333;
  }

  .status-in-progress {
    background-color: #007bff;
    animation: pulse 1.5s infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.6; }
  }

  .loading {
    text-align: center;
    padding: 40px;
    color: #666;
  }

  .error-message {
    padding: 16px;
    background: #ffebee;
    border: 1px solid #c62828;
    border-radius: 4px;
    color: #b71c1c;
    margin-top: 20px;
  }

  .info-message {
    padding: 12px;
    background: #e3f2fd;
    border-left: 4px solid #1976d2;
    border-radius: 4px;
    margin-bottom: 20px;
    font-size: 0.95em;
  }

  .no-data {
    padding: 20px;
    text-align: center;
    color: #666;
    background: #f9f9f9;
    border-radius: 4px;
    margin-top: 20px;
  }

  @media (max-width: 768px) {
    .controls-section {
      flex-direction: column;
    }

    .control-group {
      min-width: 100%;
    }

    .jobs-table {
      font-size: 0.9em;
    }

    .jobs-table th,
    .jobs-table td {
      padding: 8px;
    }

    .status-badge {
      width: 16px;
      height: 16px;
    }
  }
</style>

<div class="dashboard-container">
  <div class="controls-section">
    <div class="control-group">
      <label for="workflow-select">Select Workflow:</label>
      <select id="workflow-select">
        <option value="">Loading workflows...</option>
      </select>
    </div>
    <div class="control-group">
      <label for="job-search">Search Jobs:</label>
      <input type="text" id="job-search" placeholder="Search by job name...">
    </div>
  </div>

  <div id="jobs-container" class="loading">
    <p>⏳ Loading workflow data...</p>
  </div>
</div>

<script>
  const owner = 'urunc-dev';
  const repo = 'urunc';
  let allWorkflows = [];
  let workflowRunsCache = {};
  let jobsCache = {};

  async function loadWorkflows() {
    try {
      const response = await fetch(
        `https://api.github.com/repos/${owner}/${repo}/actions/workflows`,
        { headers: { 'Accept': 'application/vnd.github.v3+json' } }
      );

      if (!response.ok) throw new Error(`API error: ${response.status}`);

      const data = await response.json();
      allWorkflows = (data.workflows || []).filter(w => w.state === 'active');

      if (allWorkflows.length === 0) {
        document.getElementById('jobs-container').innerHTML = 
          '<div class="no-data">No active workflows found.</div>';
        return;
      }

      const select = document.getElementById('workflow-select');
      select.innerHTML = allWorkflows.map((w, idx) => 
        `<option value="${idx}">${escapeHtml(w.name)}</option>`
      ).join('');

      // Load first workflow
      await loadWorkflowJobs(0);

      select.addEventListener('change', (e) => {
        loadWorkflowJobs(parseInt(e.target.value));
      });

      document.getElementById('job-search').addEventListener('input', filterJobs);

    } catch (error) {
      console.error('Error:', error);
      document.getElementById('jobs-container').innerHTML = 
        `<div class="error-message"><strong>Failed to load workflows:</strong> ${escapeHtml(error.message)}</div>`;
    }
  }

  async function loadWorkflowJobs(workflowIndex) {
    const container = document.getElementById('jobs-container');
    const workflow = allWorkflows[workflowIndex];

    container.innerHTML = '<div class="loading"><p>⏳ Loading jobs and run history...</p></div>';

    try {
      const runsResponse = await fetch(
        `https://api.github.com/repos/${owner}/${repo}/actions/workflows/${workflow.id}/runs?per_page=10`,
        { headers: { 'Accept': 'application/vnd.github.v3+json' } }
      );

      if (!runsResponse.ok) throw new Error('Failed to fetch runs');

      const runsData = await runsResponse.json();
      const runs = runsData.workflow_runs || [];

      // Collect all jobs from all runs
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
            jobMap.set(job.name, {
              name: job.name,
              runs: []
            });
          }
          jobMap.get(job.name).runs.push({
            status: job.conclusion || job.status,
            runId: run.id,
            htmlUrl: job.html_url
          });
        }
      }

      // Calculate pass rates and format data
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

      // Sort by pass rate
      jobsData.sort((a, b) => a.passRate - b.passRate);

      renderJobsTable(jobsData, workflow);
      document.getElementById('job-search').value = '';

    } catch (error) {
      console.error('Error loading jobs:', error);
      container.innerHTML = `<div class="error-message">Failed to load jobs: ${escapeHtml(error.message)}</div>`;
    }
  }

  function renderJobsTable(jobsData, workflow) {
    const container = document.getElementById('jobs-container');

    if (jobsData.length === 0) {
      container.innerHTML = '<div class="no-data">No job data available for this workflow.</div>';
      return;
    }

    let html = `<div class="info-message">
      <strong>${escapeHtml(workflow.name)}</strong> - Showing ${jobsData.length} job(s)
    </div>`;

    html += `
      <table class="jobs-table">
        <thead>
          <tr>
            <th style="width: 40%;">Job Name</th>
            <th style="width: 15%; text-align: center;">Pass Rate</th>
            <th style="width: 45%;">Last 10 Runs</th>
          </tr>
        </thead>
        <tbody>
    `;

    jobsData.forEach(job => {
      const runsHtml = job.runs.map(run => {
        let title, className;
        switch (run.status) {
          case 'success':
            title = 'Pass';
            className = 'status-pass';
            break;
          case 'failure':
            title = 'Fail';
            className = 'status-fail';
            break;
          case 'skipped':
            title = 'Skip';
            className = 'status-skip';
            break;
          default:
            title = 'In Progress';
            className = 'status-in-progress';
        }
        return `<span class="status-badge ${className}" title="${title}"></span>`;
      }).join('');

      html += `
        <tr>
          <td><a href="https://github.com/${owner}/${repo}/actions" class="job-name">${escapeHtml(job.name)}</a></td>
          <td class="pass-rate">${job.passRate.toFixed(2)}%</td>
          <td><div class="run-status">${runsHtml}</div></td>
        </tr>
      `;
    });

    html += `
        </tbody>
      </table>
    `;

    container.innerHTML = html;
    window.currentJobsData = jobsData;
  }

  function filterJobs() {
    const searchTerm = document.getElementById('job-search').value.toLowerCase();
    if (!window.currentJobsData) return;

    const rows = document.querySelectorAll('.jobs-table tbody tr');
    rows.forEach(row => {
      const jobName = row.querySelector('.job-name').textContent.toLowerCase();
      row.style.display = jobName.includes(searchTerm) ? '' : 'none';
    });
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
</script>
