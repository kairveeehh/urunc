#!/usr/bin/env python3
"""
Script to fetch and display CI/CD workflows for the urunc project.

This script can be used locally to generate a summary of all workflows,
or to verify workflows are accessible from the GitHub API.
"""

import json
import sys
import requests
from typing import Dict, List, Any
from datetime import datetime, timezone

def fetch_workflows(owner: str = "urunc-dev", repo: str = "urunc") -> List[Dict[str, Any]]:
    """Fetch workflows from GitHub API."""
    url = f"https://api.github.com/repos/{owner}/{repo}/actions/workflows"
    headers = {
        "Accept": "application/vnd.github.v3+json"
    }
    
    try:
        response = requests.get(url, headers=headers, timeout=10)
        response.raise_for_status()
        data = response.json()
        return data.get("workflows", [])
    except requests.exceptions.RequestException as e:
        print(f"Error fetching workflows: {e}", file=sys.stderr)
        sys.exit(1)

def categorize_workflows(workflows: List[Dict[str, Any]]) -> Dict[str, List[Dict[str, Any]]]:
    """Categorize workflows by name."""
    categories = {}
    for workflow in workflows:
        name = workflow.get("name", "Uncategorized")
        if name not in categories:
            categories[name] = []
        categories[name].append(workflow)
    return categories

def print_workflow_summary(owner: str = "urunc-dev", repo: str = "urunc") -> None:
    """Print a summary of all workflows."""
    print(f"Fetching workflows for {owner}/{repo}...")
    workflows = fetch_workflows(owner, repo)
    
    if not workflows:
        print("No workflows found.")
        return
    
    categories = categorize_workflows(workflows)
    
    print(f"\nFound {len(workflows)} workflow(s) in {len(categories)} categor(ies):\n")
    
    for category in sorted(categories.keys()):
        workflows_in_category = categories[category]
        latest = workflows_in_category[0]
        
        print(f"  📋 {category}")
        print(f"     Path: {latest.get('path')}")
        print(f"     State: {latest.get('state')}")
        print(f"     ID: {latest.get('id')}")
        print(f"     URL: https://github.com/{owner}/{repo}/actions/workflows/{latest.get('path')}")
        print()

def export_workflows_json(owner: str = "urunc-dev", repo: str = "urunc", output_file: str = "workflows.json") -> None:
    """Export workflow data to JSON file."""
    print(f"Fetching workflows for {owner}/{repo}...")
    workflows = fetch_workflows(owner, repo)
    
    categories = categorize_workflows(workflows)
    
    export_data = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "repository": f"{owner}/{repo}",
        "total_workflows": len(workflows),
        "categories": {}
    }
    
    for category in sorted(categories.keys()):
        workflows_in_category = categories[category]
        export_data["categories"][category] = [
            {
                "name": w.get("name"),
                "path": w.get("path"),
                "state": w.get("state"),
                "id": w.get("id"),
                "created_at": w.get("created_at"),
                "updated_at": w.get("updated_at"),
                "url": f"https://github.com/{owner}/{repo}/actions/workflows/{w.get('path')}"
            }
            for w in workflows_in_category
        ]
    
    with open(output_file, 'w') as f:
        json.dump(export_data, f, indent=2)
    
    print(f"✓ Workflows exported to {output_file}")

if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(
        description="Fetch and display CI/CD workflows for urunc"
    )
    parser.add_argument(
        "--owner",
        default="urunc-dev",
        help="GitHub repository owner (default: urunc-dev)"
    )
    parser.add_argument(
        "--repo",
        default="urunc",
        help="GitHub repository name (default: urunc)"
    )
    parser.add_argument(
        "--export",
        metavar="FILE",
        help="Export workflow data to JSON file"
    )
    
    args = parser.parse_args()
    
    if args.export:
        export_workflows_json(args.owner, args.repo, args.export)
    else:
        print_workflow_summary(args.owner, args.repo)
