# 统一生成应用、前端和托盘图标；在项目根目录执行此脚本。
$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Drawing

$source = [System.Drawing.Bitmap]::new(1024, 1024)
$graphics = [System.Drawing.Graphics]::FromImage($source)
try {
    $graphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
    $graphics.Clear([System.Drawing.Color]::Transparent)
    $background = [System.Drawing.SolidBrush]::new([System.Drawing.ColorTranslator]::FromHtml('#334155'))
    $pen = [System.Drawing.Pen]::new([System.Drawing.Color]::White, 44)
    $pen.LineJoin = [System.Drawing.Drawing2D.LineJoin]::Round
    $pen.StartCap = [System.Drawing.Drawing2D.LineCap]::Round
    $pen.EndCap = [System.Drawing.Drawing2D.LineCap]::Round
    try {
        $graphics.FillEllipse($background, 16, 16, 992, 992)
        $outline = [System.Drawing.PointF[]]@(
            [System.Drawing.PointF]::new(512, 220),
            [System.Drawing.PointF]::new(770, 365),
            [System.Drawing.PointF]::new(770, 660),
            [System.Drawing.PointF]::new(512, 805),
            [System.Drawing.PointF]::new(254, 660),
            [System.Drawing.PointF]::new(254, 365)
        )
        $graphics.DrawPolygon($pen, $outline)
        $graphics.DrawLine($pen, 254, 365, 512, 510)
        $graphics.DrawLine($pen, 770, 365, 512, 510)
        $graphics.DrawLine($pen, 512, 510, 512, 805)
    } finally {
        $pen.Dispose()
        $background.Dispose()
    }

    $sizes = @(16, 20, 24, 32, 40, 48, 64, 128, 256)
    $images = [System.Collections.Generic.List[byte[]]]::new()
    foreach ($size in $sizes) {
        $bitmap = [System.Drawing.Bitmap]::new($size, $size)
        $canvas = [System.Drawing.Graphics]::FromImage($bitmap)
        $stream = [System.IO.MemoryStream]::new()
        try {
            $canvas.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
            $canvas.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
            $canvas.DrawImage($source, 0, 0, $size, $size)
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
    $graphics.Dispose()
    $source.Dispose()
}
