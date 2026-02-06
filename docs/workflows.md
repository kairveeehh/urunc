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
