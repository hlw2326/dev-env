param(
    [Parameter(Mandatory = $true)][string]$HelperPath,
    [Parameter(Mandatory = $true)][string]$InstallExecutable,
    [Parameter(Mandatory = $true)][string]$OptionsFile
)

$ErrorActionPreference = 'Stop'
$dataRoot = [IO.Path]::GetFullPath((Join-Path ([Environment]::GetFolderPath('LocalApplicationData')) 'DevEnv'))
$dataRootPrefix = $dataRoot.TrimEnd('\') + '\'
$commandRoot = Join-Path $dataRoot 'bin'
if (-not (Test-Path -LiteralPath $HelperPath -PathType Leaf)) { throw 'Uninstall helper is missing.' }
if (-not (Test-Path -LiteralPath $OptionsFile -PathType Leaf)) { throw 'Uninstall options are missing.' }

Get-Process | Where-Object {
    $_.Path -and $_.Path.Equals($InstallExecutable, [StringComparison]::OrdinalIgnoreCase)
} | Stop-Process -Force

$serviceRoot = 'HKLM:\SYSTEM\CurrentControlSet\Services'
foreach ($entry in Get-ChildItem -LiteralPath $serviceRoot) {
    $properties = Get-ItemProperty -LiteralPath $entry.PSPath
    $imagePath = [Environment]::ExpandEnvironmentVariables([string]$properties.ImagePath)
    $usesInstalledHost = $imagePath.IndexOf($InstallExecutable, [StringComparison]::OrdinalIgnoreCase) -ge 0
    $usesDataRoot = $imagePath.IndexOf($dataRootPrefix, [StringComparison]::OrdinalIgnoreCase) -ge 0
    if (-not $usesInstalledHost -or -not $usesDataRoot) { continue }
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
        $value = [Environment]::ExpandEnvironmentVariables([string]$key.GetValue($name))
        $usesInstalledApp = $value.IndexOf($InstallExecutable, [StringComparison]::OrdinalIgnoreCase) -ge 0
        $usesDataRoot = $value.IndexOf($dataRootPrefix, [StringComparison]::OrdinalIgnoreCase) -ge 0
        if ($usesInstalledApp -or $usesDataRoot) { Remove-ItemProperty -LiteralPath $runKey -Name $name }
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
