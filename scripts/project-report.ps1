#!/usr/bin/env pwsh
#
# Generate a project status report from GitHub Issues.
# Usage: pwsh scripts/project-report.ps1 [-Milestone "Phase N: Name"]
#
# Outputs: milestone progress, epic status, in-progress work, recent completions.
#

param(
    [string]$Milestone = ""
)

$repo = "ptetau/workshop"

Write-Host "`n# Project Status Report" -ForegroundColor Cyan
Write-Host "Generated: $(Get-Date -Format 'yyyy-MM-dd HH:mm')`n"

# --- Milestone Progress ---
Write-Host "## Milestone Progress" -ForegroundColor Cyan
$milestones = gh api "repos/$repo/milestones" --jq '.[] | @json' | ForEach-Object { $_ | ConvertFrom-Json }
foreach ($m in $milestones | Sort-Object number) {
    $total = $m.open_issues + $m.closed_issues
    if ($total -eq 0) {
        $pct = 0
    } else {
        $pct = [math]::Floor(($m.closed_issues * 100) / $total)
    }
    $bar = ("█" * [math]::Floor($pct / 8)) + ("░" * (12 - [math]::Floor($pct / 8)))
    $status = if ($m.open_issues -eq 0 -and $total -gt 0) { "✅" } elseif ($m.closed_issues -gt 0) { "🟡" } else { "⬜" }
    Write-Host ("  {0} {1}: {2}/{3} done ({4}%) {5}" -f $status, $m.title, $m.closed_issues, $total, $pct, $bar)
}

# --- Epic Progress ---
Write-Host "`n## Epic Progress" -ForegroundColor Cyan
$epics = gh issue list --repo $repo --label "type:epic" --state all --limit 30 --json number,title,state,milestone,labels |
    ConvertFrom-Json | Sort-Object number
foreach ($e in $epics) {
    $area = ($e.labels | Where-Object { $_.name -like "area:*" } | Select-Object -First 1).name
    $ms = if ($e.milestone) { $e.milestone.title } else { "unassigned" }
    $icon = if ($e.state -eq "closed") { "✅" } else { "⬜" }
    Write-Host ("  {0} #{1} {2} [{3}] ({4})" -f $icon, $e.number, $e.title, $area, $ms)
}

# --- Open Issues by Milestone ---
if ($Milestone) {
    Write-Host "`n## Open Issues: $Milestone" -ForegroundColor Cyan
    $issues = gh issue list --repo $repo --state open --milestone $Milestone --limit 50 --json number,title,labels |
        ConvertFrom-Json | Sort-Object number
} else {
    Write-Host "`n## Open Issues (all milestones)" -ForegroundColor Cyan
    $issues = gh issue list --repo $repo --state open --limit 100 --json number,title,labels,milestone |
        ConvertFrom-Json | Sort-Object { if ($_.milestone) { $_.milestone.number } else { 999 } }, number
}

foreach ($i in $issues) {
    $labels = ($i.labels | ForEach-Object { $_.name }) -join ", "
    $ms = if ($i.milestone) { $i.milestone.title } else { "" }
    if ($Milestone) {
        Write-Host ("  #{0} {1} [{2}]" -f $i.number, $i.title, $labels)
    } else {
        Write-Host ("  #{0} {1} [{2}] ({3})" -f $i.number, $i.title, $labels, $ms)
    }
}
Write-Host ("  Total open: {0}" -f $issues.Count)

# --- In-Progress Work ---
Write-Host "`n## In-Progress Work" -ForegroundColor Cyan
$branches = git branch --list "issue-*" 2>$null
if ($branches) {
    foreach ($b in $branches) {
        Write-Host "  Branch: $($b.Trim())"
    }
} else {
    Write-Host "  No feature branches"
}

$prs = gh pr list --repo $repo --state open --json number,title,headRefName | ConvertFrom-Json
if ($prs.Count -gt 0) {
    foreach ($pr in $prs) {
        Write-Host ("  PR #{0}: {1} ({2})" -f $pr.number, $pr.title, $pr.headRefName)
    }
} else {
    Write-Host "  No open PRs"
}

# --- Recently Closed ---
Write-Host "`n## Recently Closed (last 10)" -ForegroundColor Cyan
$closed = gh issue list --repo $repo --state closed --limit 10 --json number,title,closedAt |
    ConvertFrom-Json | Sort-Object closedAt -Descending
foreach ($c in $closed) {
    $date = if ($c.closedAt) { ([datetime]$c.closedAt).ToString("yyyy-MM-dd") } else { "unknown" }
    Write-Host ("  #{0} {1} — closed {2}" -f $c.number, $c.title, $date)
}

Write-Host ""
