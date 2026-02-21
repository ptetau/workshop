#!/usr/bin/env pwsh
#
# Run all verification checks (same as CI and pre-commit hook).
# Usage: pwsh scripts/check-all.ps1
#
# Exit codes: 0 = all passed, 1 = failure (govulncheck is warn-only)
#

$ErrorActionPreference = "Stop"
$failed = $false

function Invoke-Check {
    param([string]$Name, [string]$Command, [bool]$WarnOnly = $false)
    Write-Host "`n=== $Name ===" -ForegroundColor Cyan
    try {
        Invoke-Expression $Command
        if ($LASTEXITCODE -ne 0) { throw "exit code $LASTEXITCODE" }
        Write-Host "PASS" -ForegroundColor Green
    } catch {
        if ($WarnOnly) {
            Write-Host "WARN: $Name failed (non-blocking)" -ForegroundColor Yellow
        } else {
            Write-Host "FAIL: $Name" -ForegroundColor Red
            $script:failed = $true
        }
    }
}

Invoke-Check "go build" "go build ./..."
Invoke-Check "go vet" "go vet ./..."

Write-Host "`n=== gofmt ===" -ForegroundColor Cyan
$unformatted = gofmt -l .
if ($unformatted) {
    Write-Host "FAIL: unformatted files:" -ForegroundColor Red
    $unformatted | ForEach-Object { Write-Host "  $_" }
    $failed = $true
} else {
    Write-Host "PASS" -ForegroundColor Green
}

Write-Host "`n=== go test ===" -ForegroundColor Cyan
$cgoEnabled = (go env CGO_ENABLED).Trim()
if ($cgoEnabled -eq "1") {
    Invoke-Check "go test -race" "go test -race -count=1 ./..."
} else {
    Write-Host "SKIP: go test -race (CGO_ENABLED=$cgoEnabled). Running non-race tests instead." -ForegroundColor Yellow
    Invoke-Check "go test" "go test -count=1 ./..."
}
Invoke-Check "lintguidelines" "go run ./tools/lintguidelines --root . --strict"
Invoke-Check "govulncheck" "govulncheck ./..." -WarnOnly $true

Write-Host ""
if ($failed) {
    Write-Host "Some checks FAILED." -ForegroundColor Red
    exit 1
} else {
    Write-Host "All checks passed." -ForegroundColor Green
    exit 0
}
