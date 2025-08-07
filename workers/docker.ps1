param (
    [Parameter(Mandatory = $true)]
    [string]$ImageName
)

# Get the absolute path of this script
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# Define the Dockerfile location
$DockerfileDir = Join-Path $ScriptDir "..\workers\$ImageName"
$DockerfilePath = Join-Path $DockerfileDir "Dockerfile"

# Define the root of the repo (project base with go.mod)
$RepoRoot = Join-Path $ScriptDir ".."

# Define the output .tar path
$OutputDir = Join-Path $RepoRoot "artifacts"
$OutputImage = Join-Path $OutputDir "$ImageName.tar"

# Ensure Dockerfile exists
if (-not (Test-Path $DockerfilePath)) {
    Write-Error "Dockerfile not found: $DockerfilePath"
    exit 1
}

# Create output directory if it doesn't exist
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# Build the Docker image
Write-Host "Building Docker image for: $ImageName"
docker build -f $DockerfilePath -t $ImageName $RepoRoot

if ($LASTEXITCODE -ne 0) {
    Write-Error "Docker build failed for $ImageName"
    exit 1
}

# Save the image as .tar to the artifacts folder
Write-Host "Saving Docker image to: $OutputImage"
docker save -o $OutputImage $ImageName

Write-Host "Build completed successfully: $OutputImage"
