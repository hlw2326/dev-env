param(
    [Parameter(Mandatory = $true)][string]$OutputFile,
    [Parameter(Mandatory = $true)][string]$IconPath
)

$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing
[System.Windows.Forms.Application]::EnableVisualStyles()
$dataRoot = Join-Path ([Environment]::GetFolderPath('LocalApplicationData')) 'DevEnv'
$softwareRoot = Join-Path $dataRoot 'software'
$environments = @()
if (Test-Path -LiteralPath $softwareRoot) {
    foreach ($kind in Get-ChildItem -LiteralPath $softwareRoot -Directory) {
        if ($kind.Attributes -band [IO.FileAttributes]::ReparsePoint) { throw "环境目录不能是链接：$($kind.FullName)" }
        foreach ($version in Get-ChildItem -LiteralPath $kind.FullName -Directory) {
            $environments += [pscustomobject]@{ kind = $kind.Name; version = $version.Name }
        }
    }
}

$form = New-Object System.Windows.Forms.Form
$form.Text = '卸载 DevEnv - 选择清理内容'
$brandIcon = [System.Drawing.Icon]::new($IconPath)
$form.Icon = $brandIcon
$form.Size = New-Object System.Drawing.Size(640, 620)
$form.MinimumSize = $form.Size
$form.StartPosition = 'CenterScreen'
$form.Font = New-Object System.Drawing.Font('Microsoft YaHei UI', 9)
$form.MaximizeBox = $false
$form.MinimizeBox = $false

$layout = New-Object System.Windows.Forms.TableLayoutPanel
$layout.Dock = 'Fill'
$layout.Padding = New-Object System.Windows.Forms.Padding(18)
$layout.ColumnCount = 1
$layout.RowCount = 8
foreach ($height in @(52, 30, 0, 32, 32, 48, 36, 44)) {
    $row = New-Object System.Windows.Forms.RowStyle
    if ($height -eq 0) { $row.SizeType = 'Percent'; $row.Height = 100 } else { $row.SizeType = 'Absolute'; $row.Height = $height }
    [void]$layout.RowStyles.Add($row)
}
$form.Controls.Add($layout)

$intro = New-Object System.Windows.Forms.Label
$intro.Text = "默认保留所有环境和数据。仅删除勾选内容，删除后无法恢复。`r`n项目源码目录不会删除。"
$intro.Dock = 'Fill'
$layout.Controls.Add($intro, 0, 0)

$all = New-Object System.Windows.Forms.CheckBox
$all.Text = '删除全部已安装环境（软件及其配置），或在下方单独勾选版本'
$all.ThreeState = $false
$all.Dock = 'Fill'
$layout.Controls.Add($all, 0, 1)

$list = New-Object System.Windows.Forms.CheckedListBox
$list.Dock = 'Fill'
$list.CheckOnClick = $true
$list.IntegralHeight = $false
$list.HorizontalScrollbar = $true
foreach ($environment in $environments) { [void]$list.Items.Add("$($environment.kind)  $($environment.version)", $false) }
if ($environments.Count -eq 0) { $all.Text = '未发现已安装环境'; $all.Enabled = $false; $list.Enabled = $false }
$script:syncingSelection = $false
$all.Add_CheckStateChanged({
    if ($script:syncingSelection) { return }
    $script:syncingSelection = $true
    try {
        for ($index = 0; $index -lt $list.Items.Count; $index++) { $list.SetItemChecked($index, $all.Checked) }
    } finally { $script:syncingSelection = $false }
})
$list.Add_ItemCheck({
    if ($script:syncingSelection) { return }
    $count = $list.CheckedItems.Count
    if ($_.CurrentValue -eq 'Checked') { $count-- }
    if ($_.NewValue -eq 'Checked') { $count++ }
    $script:syncingSelection = $true
    try {
        if ($count -eq 0) { $all.CheckState = 'Unchecked' }
        elseif ($count -eq $list.Items.Count) { $all.CheckState = 'Checked' }
        else { $all.CheckState = 'Indeterminate' }
    } finally { $script:syncingSelection = $false }
})
$layout.Controls.Add($list, 0, 2)

$database = New-Object System.Windows.Forms.CheckBox
$database.Text = '删除全部数据库数据（MySQL、PostgreSQL、Redis 等 dbdata 内容）'
$database.Dock = 'Fill'
$layout.Controls.Add($database, 0, 3)

$user = New-Object System.Windows.Forms.CheckBox
$user.Text = '删除全部应用用户数据（项目记录、全局设置、日志、缓存及备份）'
$user.Dock = 'Fill'
$layout.Controls.Add($user, 0, 4)

foreach ($checkbox in @($all, $database, $user)) {
    $checkbox.AutoSize = $false
    $checkbox.CheckAlign = 'MiddleLeft'
    $checkbox.TextAlign = 'MiddleLeft'
    $checkbox.UseCompatibleTextRendering = $true
    $checkbox.Padding = New-Object System.Windows.Forms.Padding(0)
}

$hint = New-Object System.Windows.Forms.Label
$hint.Text = "未勾选的环境及安装记录会保留。数据库数据与软件删除相互独立。`r`n无论是否清理数据，卸载都会停止并注销 DevEnv 管理的服务、移除命令 PATH。"
$hint.Dock = 'Fill'
$hint.ForeColor = [System.Drawing.Color]::DimGray
$layout.Controls.Add($hint, 0, 5)

$rootLabel = New-Object System.Windows.Forms.Label
$rootLabel.Text = "清理范围：$dataRoot"
$rootLabel.AutoEllipsis = $true
$rootLabel.Dock = 'Fill'
$layout.Controls.Add($rootLabel, 0, 6)

$buttons = New-Object System.Windows.Forms.FlowLayoutPanel
$buttons.Dock = 'Fill'
$buttons.FlowDirection = 'RightToLeft'
$continue = New-Object System.Windows.Forms.Button
$continue.Text = '继续卸载'
$continue.Size = New-Object System.Drawing.Size(100, 32)
$cancel = New-Object System.Windows.Forms.Button
$cancel.Text = '取消'
$cancel.Size = New-Object System.Drawing.Size(90, 32)
$cancel.DialogResult = 'Cancel'
$buttons.Controls.Add($continue)
$buttons.Controls.Add($cancel)
$layout.Controls.Add($buttons, 0, 7)
$form.CancelButton = $cancel
$continue.Add_Click({
    $selected = @()
    foreach ($index in $list.CheckedIndices) { $selected += $environments[$index] }
    $message = "删除环境版本：$($selected.Count) 个`r`n删除数据库数据：$($database.Checked)`r`n删除应用用户数据：$($user.Checked)`r`n`r`n确认继续卸载？勾选的数据删除后无法恢复。"
    $answer = [System.Windows.Forms.MessageBox]::Show($form, $message, '确认卸载选项', 'YesNo', 'Warning', 'Button2')
    if ($answer -ne 'Yes') { return }
    try {
        $options = @{ allEnvironments = ($all.CheckState -eq 'Checked'); environments = @($selected); databaseData = $database.Checked; userData = $user.Checked }
        $json = ConvertTo-Json -InputObject $options -Depth 4
        [IO.File]::WriteAllText($OutputFile, $json, [Text.UTF8Encoding]::new($false))
        $userFlag = '0'
        if ($user.Checked) { $userFlag = '1' }
        [IO.File]::WriteAllText(($OutputFile + '.ini'), "[cleanup]`r`nuserData=$userFlag`r`n", [Text.UTF8Encoding]::new($false))
        $form.DialogResult = 'OK'
        $form.Close()
    } catch {
        [void][System.Windows.Forms.MessageBox]::Show($form, $_.Exception.Message, '无法保存卸载选项', 'OK', 'Error')
    }
})
$result = $form.ShowDialog()
$form.Dispose()
$brandIcon.Dispose()
if ($result -ne 'OK') { exit 1 }
