# Set the working directory to the script location
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

# Define the paths
$ArtifactPath = Join-Path $ScriptDir "..\artifacts"
$OutputZip = Join-Path $ArtifactPath "glue-bundle.zip"
$TempFolder = Join-Path $env:TEMP "glue_etl_temp"
$RootSource = Get-Location

# Ensure the artifacts directory exists
if (!(Test-Path $ArtifactPath)) {
    New-Item -ItemType Directory -Path $ArtifactPath | Out-Null
}

# Clean up previous zip and temp folders
if (Test-Path $OutputZip) { Remove-Item $OutputZip -Force }
if (Test-Path $TempFolder) { Remove-Item $TempFolder -Recurse -Force }

# Recreate temp folder
New-Item -ItemType Directory -Path $TempFolder | Out-Null

# Select valid files only, preserving directory structure
$FilesToInclude = Get-ChildItem -Recurse -File | Where-Object {
    $_.FullName -notmatch '\\\.venv\\|\\\.idea\\|__pycache__|\.pyc$|build.ps1'
}

foreach ($file in $FilesToInclude) {
    # Skip if the file path is not under the current working directory
    if (-not ($file.FullName.StartsWith($RootSource.Path))) {
        continue
    }

    $relativePath = $file.FullName.Substring($RootSource.Path.Length).TrimStart('\')
    $destPath = Join-Path $TempFolder $relativePath
    $destDir = Split-Path $destPath -Parent

    if (!(Test-Path $destDir)) {
        New-Item -ItemType Directory -Path $destDir -Force | Out-Null
    }

    Copy-Item $file.FullName -Destination $destPath -Force
}

# Create the zip archive
Compress-Archive -Path "$TempFolder\*" -DestinationPath $OutputZip -Force

# Clean up temporary files
Remove-Item $TempFolder -Recurse -Force

Write-Host "Build completed: $OutputZip"
