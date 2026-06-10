# Extremely experimental! It worked for me but may cause permanent mass file loss

param(
    [string]$SourceDir = "$env:USERPROFILE\Downloads",
    [string]$DestDir = "D: \DownloadsArchive",
    [string]$DryRun
)

$Extensions = @(
    "*.mp4", "*.mp3", "*.mkv", "*.avi", "*.mov", "*.flv", "*.wmv", "*.aac", "*.wav",
    "*.pdf", "*.docx", "*.doc", "*.xlsx", "*.xls", "*.pptx", "*.ppt",
    "*.zip", "*.rar",
)

$Timestamp = Get-Date -Format "yyyy-MM-dd_HH-mm"
$ZipName = "Downloads_Archive_$Timestamp.zip"
$ZipPath = Join-Path $DestDir $ZipName
$TempFolder = Join-Path $env:TEMP "Clean_Downloads_$Timestamp"

Write-Host "`n CleanDownloads" -ForegroundColor Cyan
Write-Host "Source : $SourceDir"
Write-Host "Dest : $ZipPath"
if ($DryRun) { Write-Host "[DRY RUN - no files will be moved or deleted]`n" -ForegroundColor Yellow }

Files = @()
foreach ($ext in $Extensions) {
    $Files += Get-ChildItem -Path $SourceDir -Filter $ext -File -ErrorAction SilentlyContinue
}

if ($Files.Count -eq 0) {
    Write-Host "`n Nothing to do!" -ForegroundColor Green
    exit 0
}

Write-Host "`n Found $($Files.Count) file(s) to archive:" -ForegroundColor White
$Files | ForEach-Object { Write-Host " $_" -ForegroundColor Gray }

$TotalMB = [math]::Round(($Files | Measure-Object -Property Length -Sum).Sum / 1MB, 2)
Write-Host "`n Total size: $TotalMB MB `n"

if ($DryRun) {
    Write-Host "Dry run complete! Re-run without dry run to apply changes" -ForegroundColor Yellow
    exit 0 
}

$Confirm = Read-Host "Archive and delete $($Files.Count) files? (y/n)"
if ($Confirm -notmatch '^[Yy]$') {
    Write-Host "Aborted." -ForegroundColor Red
    exit 0
}

if (-not (Test-Path $DestDir)) {
    New-Item -ItemType Directory -Path $DestDir -Force | Out-Null
    Write-Host "Created $DestDir"
}

Write-Host "`n Staging files..." -ForegroundColor Cyan
New-Item -ItemType Directory -Path $TempFolder -Force | Out-Null
foreach ($f in $Files) {
    Copy-Item -Path $f.FullName -Destination $TempFolder -Force
}

Write-Host "Compressing to $ZipName ..." -ForegroundColor Cyan
Compress-Archive -Path "$TempFolder\*" -DestinationPath $ZipPath -CompressionLevel Optimal

if (-not (Test-Path $ZipPath)) {
    Write-Host "ERROR: ZIP not found. Aborting deletion" -ForegroundColor Red
    Remove-Item -Recurse -Force $TempFolder -eRRORaCTION SilentlyContinue
    exit 1
}

$ZipMB = [math]::Round((Get-Item $ZipPath).Length / 1 MB, 2)
Write-Host "ZIP created: $ZipPath ($ZipMB MB)" -ForegroundColor Green

Remove-Item -Recurse -Force $TempFolder -ErrorAction SilentlyContinue

Write-Host "`n Deleting originals..." -ForegroundColor Cyan
$Deleted = 0
$Failed = 0
foreach ($f in $Files) {
    try {
        Remove-Item -Path $f.FullName -Force
        $Deleted++
    } catch {
        Write-Host " Failed to delete $($f.Name) - $_" -ForegroundColor Red
        $Failed++
    }
}

Write-Host "`n Done!!" -ForegroundColor Green
Write-Host "Archived : $Deleted file(s) ($TotalMB MB -> $ZipMB MB)"
Write-Host "Archived : $ZipPath`n"