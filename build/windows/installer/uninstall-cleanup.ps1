param(
    [switch]$DeleteData
)

$ErrorActionPreference = 'SilentlyContinue'
$dataRoot = Join-Path ([Environment]::GetFolderPath('LocalApplicationData')) 'dev-env'
$commandRoot = Join-Path $dataRoot 'bin'
$serviceRoot = 'HKLM:\SYSTEM\CurrentControlSet\Services'

Get-ChildItem -Path $serviceRoot | ForEach-Object {
    $serviceName = $_.PSChildName
    $properties = Get-ItemProperty -LiteralPath $_.PSPath
    $imagePath = [string]$properties.ImagePath
    $managedService = $serviceName.StartsWith('DevEnv', [StringComparison]::OrdinalIgnoreCase)
    $usesDataRoot = $imagePath -and $imagePath.IndexOf($dataRoot, [StringComparison]::OrdinalIgnoreCase) -ge 0
    if (-not $managedService -and -not $usesDataRoot) {
        return
    }
    Stop-Service -Name $serviceName -Force
    & sc.exe delete $serviceName | Out-Null
}

$dataRootPrefix = $dataRoot.TrimEnd('\') + '\'
Get-Process | Where-Object {
    $_.Path -and $_.Path.StartsWith($dataRootPrefix, [StringComparison]::OrdinalIgnoreCase)
} | Stop-Process -Force

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
$pathItems = @($userPath -split ';' | Where-Object {
    $pathItem = $_.Trim().Trim('"')
    $pathItem -and -not $pathItem.Equals($commandRoot, [StringComparison]::OrdinalIgnoreCase)
})
[Environment]::SetEnvironmentVariable('Path', ($pathItems -join ';'), 'User')

if ($DeleteData) {
    Remove-Item -LiteralPath $dataRoot -Recurse -Force
}
