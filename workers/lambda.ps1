param (
  [Parameter(Mandatory = $true)]
  [string]$FunctionName
)

# Get the absolute path of this script
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# Define the Lambda function source path
$FunctionPath = Join-Path $ScriptDir "..\workers\$FunctionName"

# Define the output zip path
$OutputDir = Join-Path $ScriptDir "..\artifacts"
$OutputZip = Join-Path $OutputDir "$FunctionName.zip"

# Ensure the function folder exists
if (-not (Test-Path $FunctionPath)) {
  Write-Error "Function folder not found: $FunctionPath"
  exit 1
}

# Create output directory if it doesn't exist
if (-not (Test-Path $OutputDir)) {
  New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# Compile the Go Lambda binary as 'bootstrap'
Push-Location $FunctionPath

$env:GOOS = "linux"
$env:GOARCH = "amd64"

Write-Host "Building Lambda function: $FunctionName"
go build -o bootstrap main.go
if ($LASTEXITCODE -ne 0) {
  Write-Error "Go build failed for $FunctionName"
  Pop-Location
  exit 1
}

# Zip the bootstrap binary to the artifacts folder
Write-Host "Zipping bootstrap to: $OutputZip"
Compress-Archive -Path bootstrap -DestinationPath $OutputZip -Force

# Clean up temporary bootstrap binary
Remove-Item bootstrap

Pop-Location

Write-Host "Build completed successfully: $OutputZip"
