$ErrorActionPreference = 'SilentlyContinue'

$jobNames = @('app-manager-backend', 'app-manager-frontend')

foreach ($name in $jobNames) {
    $jobs = Get-Job -Name $name
    foreach ($job in $jobs) {
        if ($job.State -ne 'Stopped' -and $job.State -ne 'Completed' -and $job.State -ne 'Failed') {
            Stop-Job -Job $job | Out-Null
        }
        Remove-Job -Job $job -Force | Out-Null
        Write-Host "Stopped and removed job: $name (Id: $($job.Id))"
    }
}

Write-Host 'All app-manager jobs have been stopped.'
