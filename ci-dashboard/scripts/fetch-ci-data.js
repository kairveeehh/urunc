//
// This script fetches workflow run data from the GitHub API for the urunc
// repository and outputs a JSON summary of job statistics.
//
// The general flow is:
//   - Query the GitHub API for workflow runs (e.g. the last 10 runs per workflow).
//   - For each run, query the API for all jobs data.
//   - Reorganize and summarize results, where each entry contains information
//     about a workflow and how it has performed over recent runs.
//
// Usage:
//   TOKEN=<github_pat> node scripts/fetch-ci-data.js > data/workflow_stats.json
//

const TOKEN = process.env.TOKEN;

const OWNER = "urunc-dev";
const REPO = "urunc";

// Number of recent runs to fetch per workflow
const TOTAL_RUNS = 10;

// Number of jobs to fetch per paged request
const JOBS_PER_REQUEST = 50;

// Count of the number of fetches made
let fetchCount = 0;

// Perform a fetch request to the GitHub API
async function fetchUrl(url) {
  const response = await fetch(url, {
    headers: {
      Accept: "application/vnd.github+json",
      Authorization: `token ${TOKEN}`,
      "X-GitHub-Api-Version": "2022-11-28",
    },
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch from ${url}: ${response.status}: ${response.statusText}`);
  }

  const json = await response.json();
  fetchCount++;
  return json;
}

// Categorize a workflow based on its name
function categorizeWorkflow(name) {
  const lower = name.toLowerCase();
  if (lower.includes("ci") || lower.includes("nightly") || lower.includes("test")) {
    return "CI / Testing";
  }
  if (lower.includes("build") || lower.includes("upload") || lower.includes("deploy") || lower.includes("release")) {
    return "Build / Deploy";
  }
  if (lower.includes("lint") || lower.includes("codeql") || lower.includes("scorecard") ||
      lower.includes("dependency") || lower.includes("validate")) {
    return "Code Quality / Security";
  }
  return "Other";
}

// Get job data for a workflow run
async function getJobData(run) {
  async function fetchJobsByPage(page) {
    const jobsUrl = `${run.jobs_url}?per_page=${JOBS_PER_REQUEST}&page=${page}`;
    const response = await fetch(jobsUrl, {
      headers: {
        Accept: "application/vnd.github+json",
        Authorization: `token ${TOKEN}`,
        "X-GitHub-Api-Version": "2022-11-28",
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch jobs from ${jobsUrl}: ${response.status}: ${response.statusText}`);
    }
    const json = await response.json();
    fetchCount++;
    return json;
  }

  function fetchJobs(p) {
    return fetchJobsByPage(p).then(function (jobsRequest) {
      for (const job of jobsRequest.jobs) {
        runWithJobData.jobs.push({
          name: job.name,
          run_id: job.run_id,
          html_url: job.html_url,
          conclusion: job.conclusion,
        });
      }
      if (p * JOBS_PER_REQUEST >= jobsRequest.total_count) {
        return runWithJobData;
      }
      return fetchJobs(p + 1);
    });
  }

  const runWithJobData = {
    id: run.id,
    run_number: run.run_number,
    created_at: run.created_at,
    conclusion: null,
    jobs: [],
  };
  if (run.status === "in_progress") {
    return new Promise((resolve) => {
      resolve(runWithJobData);
    });
  }
  runWithJobData.conclusion = run.conclusion;
  return fetchJobs(1);
}

// Compute job stats across all runs
function computeJobStats(runsWithJobData) {
  const jobStats = {};
  for (const run of runsWithJobData) {
    for (const job of run.jobs) {
      if (!(job.name in jobStats)) {
        jobStats[job.name] = {
          runs: 0,
          fails: 0,
          skips: 0,
          urls: [],
          results: [],
          run_nums: [],
        };
      }
      const stat = jobStats[job.name];
      stat.runs += 1;
      stat.run_nums.push(run.run_number);
      stat.urls.push(job.html_url);
      if (job.conclusion !== "success") {
        if (job.conclusion === "skipped") {
          stat.skips += 1;
          stat.results.push("Skip");
        } else {
          stat.fails += 1;
          stat.results.push("Fail");
        }
      } else {
        stat.results.push("Pass");
      }
    }
  }
  return jobStats;
}

async function main() {
  // Fetch list of workflows
  const workflowsUrl =
    `https://api.github.com/repos/${OWNER}/${REPO}/actions/workflows`;
  const workflowsData = await fetchUrl(workflowsUrl);

  const allWorkflowStats = {};

  for (const workflow of workflowsData.workflows) {
    // Skip dynamic/internal workflows
    if (workflow.path.startsWith("dynamic/")) {
      continue;
    }

    const runsUrl =
      `https://api.github.com/repos/${OWNER}/${REPO}/actions/workflows/` +
      `${workflow.id}/runs?per_page=${TOTAL_RUNS}`;

    try {
      const runsData = await fetchUrl(runsUrl);
      if (!runsData.workflow_runs || runsData.workflow_runs.length === 0) {
        continue;
      }

      const promises = [];
      for (const run of runsData.workflow_runs) {
        promises.push(getJobData(run));
      }
      const runsWithJobData = await Promise.all(promises);
      const jobStats = computeJobStats(runsWithJobData);

      allWorkflowStats[workflow.name] = {
        category: categorizeWorkflow(workflow.name),
        path: workflow.path,
        html_url: workflow.html_url,
        jobs: jobStats,
        total_runs: runsData.workflow_runs.length,
        latest_run: runsData.workflow_runs.length > 0
          ? {
              conclusion: runsData.workflow_runs[0].conclusion,
              created_at: runsData.workflow_runs[0].created_at,
              html_url: runsData.workflow_runs[0].html_url,
            }
          : null,
      };
    } catch (err) {
      console.error(`Error fetching data for workflow ${workflow.name}: ${err.message}`);
    }
  }

  console.log(JSON.stringify(allWorkflowStats, null, 2));
  console.error(`Total API requests: ${fetchCount}`);
}

main();
