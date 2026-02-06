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

  .pass-rate-badge {
    display: inline-block;
    padding: 4px 10px;
    border-radius: 12px;
    color: white;
    font-size: 0.9em;
    font-weight: 600;
  }

  .run-status {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
    align-items: center;
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    font-size: 14px;
    font-weight: bold;
    cursor: help;
    transition: transform 0.2s;
  }

  .status-badge:hover {
    transform: scale(1.2);
  }

  .status-pass {
    color: #fff;
    background-color: #28a745;
  }

  .status-fail {
    color: #fff;
    background-color: #dc3545;
  }

  .status-skip {
    color: #666;
    background-color: #ffc107;
  }

  .status-in-progress {
    color: #fff;
    background-color: #007bff;
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
    line-height: 1.6;
  }

  .error-message strong {
    display: block;
    margin-bottom: 8px;
    font-size: 1.1em;
  }

  .error-message p {
    margin: 8px 0;
  }

  .error-message ul {
    margin: 8px 0;
  }

  .error-message code {
    background: #fff;
    padding: 2px 6px;
    border-radius: 3px;
    font-family: monospace;
    font-size: 0.9em;
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
      font-size: 0.85em;
    }

    .jobs-table th,
    .jobs-table td {
      padding: 8px;
    }

    .status-badge {
      width: 20px;
      height: 20px;
      font-size: 12px;
    }

    .pass-rate-badge {
      font-size: 0.8em;
      padding: 3px 8px;
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
