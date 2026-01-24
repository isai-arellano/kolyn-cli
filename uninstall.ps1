# Configuración
$Binary = "kolyn"
$InstallDir = "$env:USERPROFILE\bin"
$KolynDir = "$env:USERPROFILE\.kolyn"

# Colores
$Blue = [ConsoleColor]::Blue
$Green = [ConsoleColor]::Green
$Red = [ConsoleColor]::Red
$Yellow = [ConsoleColor]::Yellow
$Reset = [ConsoleColor]::White

Write-Host "Uninstalling Kolyn CLI..." -ForegroundColor $Blue
Write-Host ""

# 1. Check if installed
$BinaryPath = Join-Path $InstallDir "$Binary.exe"
if (-not (Test-Path $BinaryPath)) {
    Write-Host "Kolyn binary not found in $InstallDir" -ForegroundColor $Yellow
    
    # Check PATH just in case
    if (Get-Command $Binary -ErrorAction SilentlyContinue) {
        $FoundPath = (Get-Command $Binary).Source
        Write-Host "Found in PATH: $FoundPath" -ForegroundColor $Green
        $BinaryPath = $FoundPath
    } else {
        Write-Host "Kolyn is not installed." -ForegroundColor $Red
        exit 1
    }
}

# 2. Remove Binary
Write-Host "Removing binary from $BinaryPath..."
try {
    # Try to stop if running
    if (Get-Process $Binary -ErrorAction SilentlyContinue) {
        Stop-Process -Name $Binary -Force -ErrorAction SilentlyContinue
    }
    Remove-Item -Path $BinaryPath -Force -ErrorAction Stop
    Write-Host "Kolyn binary removed successfully!" -ForegroundColor $Green
} catch {
    Write-Host "Error removing binary: $_" -ForegroundColor $Red
    Write-Host "Try running as Administrator." -ForegroundColor $Yellow
    exit 1
}

Write-Host ""

# 3. Docker Services
Write-Host "Do you want to remove Docker services created by Kolyn? [y/N]" -ForegroundColor $Yellow -NoNewline
$response = Read-Host " "

if ($response -match "^(y|yes|s|si)$") {
    $ServicesDir = Join-Path $KolynDir "services"
    if (Test-Path $ServicesDir) {
        Write-Host "Removing services in $ServicesDir..."

        # Stop services loop
        $Services = Get-ChildItem -Path $ServicesDir -Directory
        foreach ($Service in $Services) {
            $ComposeFile = Join-Path $Service.FullName "docker-compose.yml"
            if (Test-Path $ComposeFile) {
                Write-Host "  Stopping $($Service.Name)..."
                # Run docker compose down inside the directory
                Push-Location $Service.FullName
                try {
                    docker compose down -v | Out-Null
                } catch {
                    Write-Host "  Warning: Failed to stop $($Service.Name)" -ForegroundColor $Red
                }
                Pop-Location
            }
        }

        Remove-Item -Recurse -Force $ServicesDir
        Write-Host "Docker services removed!" -ForegroundColor $Green
    } else {
        Write-Host "No Docker services found." -ForegroundColor $Blue
    }
}

Write-Host ""

# 4. Configuration & Skills
Write-Host "Do you want to remove configuration files? [y/N]" -ForegroundColor $Yellow -NoNewline
$response = Read-Host " "

if ($response -match "^(y|yes|s|si)$") {
    if (Test-Path $KolynDir) {
        Write-Host "Do you want to KEEP your downloaded skills/sources? (Recommended if re-installing) [Y/n]" -ForegroundColor $Yellow -NoNewline
        $keepSkills = Read-Host " "

        if ($keepSkills -match "^(n|no)$") {
            Write-Host "Removing entire .kolyn directory..."
            Remove-Item -Recurse -Force $KolynDir
            Write-Host "All configuration and skills removed!" -ForegroundColor $Green
        } else {
            Write-Host "Cleaning up configuration but keeping skills..."
            # Remove everything except 'skills' and 'sources' directories
            Get-ChildItem -Path $KolynDir | Where-Object { $_.Name -ne "skills" -and $_.Name -ne "sources" } | Remove-Item -Recurse -Force
            Write-Host "Configuration removed. Skills preserved in .kolyn\skills" -ForegroundColor $Green
        }
    } else {
        Write-Host "No configuration files found." -ForegroundColor $Blue
    }
}

Write-Host ""
Write-Host "✓ Kolyn has been uninstalled successfully!" -ForegroundColor $Green
Write-Host "Thank you for using Kolyn CLI."
Write-Host ""
