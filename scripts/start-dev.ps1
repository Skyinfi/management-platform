$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$backendDir = Join-Path $root 'app-manager'
$frontendDir = Join-Path $root 'manager-frontend-app'

$backendJob = Start-Job -Name 'app-manager-backend' -ScriptBlock {
    param($dir)
    Set-Location $dir
    go run ./cmd/app-manager
} -ArgumentList $backendDir

$frontendJob = Start-Job -Name 'app-manager-frontend' -ScriptBlock {
    param($dir)
    Set-Location $dir
    npm run dev
} -ArgumentList $frontendDir

Write-Host "Backend job started: $($backendJob.Id)"
Write-Host "Frontend job started: $($frontendJob.Id)"
Write-Host 'Backend:  http://localhost:8080'
Write-Host 'Frontend: http://localhost:5173'
Write-Host 'Press Ctrl+C to stop this script. Jobs will continue running until the shell exits.'

while ($true) {
    Start-Sleep -Seconds 5
    $backendState = Get-Job -Id $backendJob.Id -ErrorAction SilentlyContinue
    $frontendState = Get-Job -Id $frontendJob.Id -ErrorAction SilentlyContinue

    if (-not $backendState -or -not $frontendState) {
        Write-Host 'A job was removed or stopped.'
        break
    }

    if ($backendState.State -eq 'Failed' -or $frontendState.State -eq 'Failed') {
        Write-Host 'A job failed. Check job output with Get-Job and Receive-Job.'
        break
    }
}
