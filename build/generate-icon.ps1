# 统一生成应用、前端和托盘图标；在项目根目录执行此脚本。
$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Drawing

$source = [System.Drawing.Image]::FromFile((Join-Path $PSScriptRoot 'logo.png'))
try {
    $sizes = @(16, 20, 24, 32, 40, 48, 64, 128, 256)
    $images = [System.Collections.Generic.List[byte[]]]::new()
    foreach ($size in $sizes) {
        $bitmap = [System.Drawing.Bitmap]::new($size, $size)
        $canvas = [System.Drawing.Graphics]::FromImage($bitmap)
        $stream = [System.IO.MemoryStream]::new()
        try {
            $canvas.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
            $canvas.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
            $canvas.Clear([System.Drawing.Color]::Transparent)
            $scale = $size / [double][Math]::Max($source.Width, $source.Height)
            $width = [single]($source.Width * $scale)
            $height = [single]($source.Height * $scale)
            $bounds = [System.Drawing.RectangleF]::new(($size - $width) / 2, ($size - $height) / 2, $width, $height)
            $canvas.DrawImage($source, $bounds)
            $bitmap.Save($stream, [System.Drawing.Imaging.ImageFormat]::Png)
            $images.Add($stream.ToArray())
        } finally {
            $stream.Dispose()
            $canvas.Dispose()
            $bitmap.Dispose()
        }
    }
    [System.IO.File]::WriteAllBytes((Join-Path $PSScriptRoot 'appicon.png'), $images[$images.Count - 1])
    $icon = [System.IO.MemoryStream]::new()
    $writer = [System.IO.BinaryWriter]::new($icon)
    try {
        $writer.Write([uint16]0)
        $writer.Write([uint16]1)
        $writer.Write([uint16]$sizes.Count)
        $offset = 6 + 16 * $sizes.Count
        for ($index = 0; $index -lt $sizes.Count; $index++) {
            $dimension = $sizes[$index] % 256
            $writer.Write([byte]$dimension)
            $writer.Write([byte]$dimension)
            $writer.Write([byte]0)
            $writer.Write([byte]0)
            $writer.Write([uint16]1)
            $writer.Write([uint16]32)
            $writer.Write([uint32]$images[$index].Length)
            $writer.Write([uint32]$offset)
            $offset += $images[$index].Length
        }
        foreach ($bytes in $images) { $writer.Write([byte[]]$bytes) }
        $writer.Flush()
        [System.IO.File]::WriteAllBytes((Join-Path $PSScriptRoot 'windows/icon.ico'), $icon.ToArray())
    } finally {
        $writer.Dispose()
        $icon.Dispose()
    }
} finally {
    $source.Dispose()
}
