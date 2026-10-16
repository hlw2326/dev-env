param(
    [Parameter(Mandatory = $true)][string]$HelperPath,
    [Parameter(Mandatory = $true)][string]$InstallExecutable,
    [Parameter(Mandatory = $true)][string]$OptionsFile
)

$ErrorActionPreference = 'Stop'
$dataRoot = [IO.Path]::GetFullPath((Join-Path ([Environment]::GetFolderPath('LocalApplicationData')) 'dev-env'))
$dataRootPrefix = $dataRoot.TrimEnd('\') + '\'
$commandRoot = Join-Path $dataRoot 'bin'
$options = Get-Content -LiteralPath $OptionsFile -Raw -Encoding UTF8 | ConvertFrom-Json
if (-not (Test-Path -LiteralPath $HelperPath -PathType Leaf)) { throw 'Uninstall helper is missing.' }

# Stop only this installed application; the temporary helper is outside its directory.
Get-Process | Where-Object {
    $_.Path -and $_.Path.Equals($InstallExecutable, [StringComparison]::OrdinalIgnoreCase)
} | Stop-Process -Force

$serviceRoot = 'HKLM:\SYSTEM\CurrentControlSet\Services'
foreach ($entry in Get-ChildItem -LiteralPath $serviceRoot) {
    $properties = Get-ItemProperty -LiteralPath $entry.PSPath
    $imagePath = ([string]$properties.ImagePath).TrimStart().TrimStart('"')
    # A service name alone is not ownership evidence. Its executable must be inside this data root.
    if (-not $imagePath.StartsWith($dataRootPrefix, [StringComparison]::OrdinalIgnoreCase)) { continue }
    $serviceName = $entry.PSChildName
    $service = Get-Service -Name $serviceName
    if ($service.Status -ne 'Stopped') {
        Stop-Service -Name $serviceName
        $service.WaitForStatus('Stopped', [TimeSpan]::FromSeconds(30))
    }
    & "$env:SystemRoot\System32\sc.exe" delete $serviceName
    if ($LASTEXITCODE -ne 0) { throw "Cannot unregister service: $serviceName" }
}

Get-Process | Where-Object {
    $_.Path -and $_.Path.StartsWith($dataRootPrefix, [StringComparison]::OrdinalIgnoreCase)
} | Stop-Process -Force

$runKey = 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Run'
if (Test-Path -LiteralPath $runKey) {
    $key = Get-Item -LiteralPath $runKey
    foreach ($name in $key.GetValueNames()) {
        $value = ([string]$key.GetValue($name)).TrimStart().TrimStart('"')
        $applicationEntry = $name -eq 'DevEnv' -and $value.Trim('"').Equals($InstallExecutable, [StringComparison]::OrdinalIgnoreCase)
        if ($applicationEntry -or $value.StartsWith($dataRootPrefix, [StringComparison]::OrdinalIgnoreCase)) {
            Remove-ItemProperty -LiteralPath $runKey -Name $name
        }
    }
}

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
$pathItems = @($userPath -split ';' | Where-Object {
    $pathItem = [Environment]::ExpandEnvironmentVariables($_.Trim().Trim('"')).TrimEnd('\')
    $pathItem -and -not $pathItem.Equals($commandRoot, [StringComparison]::OrdinalIgnoreCase)
})
[Environment]::SetEnvironmentVariable('Path', ($pathItems -join ';'), 'User')

& $HelperPath --uninstall-cleanup $OptionsFile
if ($LASTEXITCODE -ne 0) { throw "Managed data cleanup failed (exit $LASTEXITCODE). Files not removed successfully may remain." }

if (Test-Path -LiteralPath $commandRoot) {
    $resolved = [IO.Path]::GetFullPath($commandRoot)
    if (-not $resolved.StartsWith($dataRootPrefix, [StringComparison]::OrdinalIgnoreCase)) { throw 'Invalid command directory.' }
    if ((Get-Item -LiteralPath $commandRoot).Attributes -band [IO.FileAttributes]::ReparsePoint) { throw 'Linked command directory is not allowed.' }
    $linked = Get-ChildItem -LiteralPath $commandRoot -Recurse -Force | Where-Object { $_.Attributes -band [IO.FileAttributes]::ReparsePoint }
    if ($linked) { throw 'Linked command directory is not allowed.' }
    Remove-Item -LiteralPath $commandRoot -Recurse -Force
}
