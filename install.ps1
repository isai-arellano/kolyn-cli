# Configuración
$ErrorActionPreference = "Stop"
$Repo = "isai-arellano/kolyn-cli"
$Binary = "kolyn"

# Definir directorios
if ($env:USERPROFILE) {
    $HomeDir = $env:USERPROFILE
} else {
    $HomeDir = $HOME
}
$InstallDir = Join-Path $HomeDir "bin"
$DestExe = Join-Path $InstallDir "$Binary.exe"

# Colores
$Blue = [ConsoleColor]::Blue
$Green = [ConsoleColor]::Green
$Red = [ConsoleColor]::Red
$Yellow = [ConsoleColor]::Yellow
$Reset = [ConsoleColor]::White
$Gray = [ConsoleColor]::Gray

Write-Host "Installing Kolyn CLI..." -ForegroundColor $Blue
Write-Host "Install Directory: $InstallDir" -ForegroundColor $Gray

# Detectar Arquitectura
$Arch = "x86_64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $Arch = "arm64"
} elseif ($env:PROCESSOR_ARCHITECTURE -eq "x86") {
    $Arch = "i386"
}

Write-Host "Detected: Windows $Arch"

# Construir nombre del archivo
$Filename = "kolyn_Windows_$Arch.zip"
$LatestUrl = "https://github.com/$Repo/releases/latest/download/$Filename"

# Crear directorio temporal
$BaseTemp = $env:TEMP
if (-not $BaseTemp) {
    $BaseTemp = [System.IO.Path]::GetTempPath()
}
$TmpDir = Join-Path $BaseTemp "kolyn_install"

Write-Host "Temp path: $TmpDir" -ForegroundColor $Gray

if (Test-Path $TmpDir) { 
    try {
        $OldEAP = $ErrorActionPreference
        $ErrorActionPreference = "SilentlyContinue"
        Remove-Item -Recurse -Force $TmpDir
        $ErrorActionPreference = $OldEAP
    } catch {
        $ErrorActionPreference = $OldEAP
        Write-Host "Warning: Could not remove existing temp dir. Trying to proceed..." -ForegroundColor $Yellow
    }
}
New-Item -ItemType Directory -Force -Path $TmpDir | Out-Null

# Descargar
Write-Host "Downloading from $LatestUrl..."
$ZipPath = Join-Path $TmpDir $Filename
try {
    Invoke-WebRequest -Uri $LatestUrl -OutFile $ZipPath
} catch {
    Write-Host "Error downloading release. Check if release exists for this architecture." -ForegroundColor $Red
    Write-Host $_.Exception.Message -ForegroundColor $Red
    exit 1
}

# Extraer
Write-Host "Extracting..."
Expand-Archive -Path $ZipPath -DestinationPath $TmpDir -Force

# Crear directorio de instalación si no existe
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

# Buscar binario (recursivo)
$SourceExe = Get-ChildItem -Path $TmpDir -Recurse -Filter "$Binary.exe" | Select-Object -First 1

if ($null -ne $SourceExe) {
    Write-Host "Found binary at: $($SourceExe.FullName)"
    Write-Host "Installing to: $DestExe"
    
    # Detener proceso si corre
    if (Get-Process $Binary -ErrorAction SilentlyContinue) {
        Stop-Process -Name $Binary -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 1
    }
    
    # Mover
    if (Test-Path $DestExe) {
        Remove-Item -Path $DestExe -Force -ErrorAction SilentlyContinue
    }
    Move-Item -Path $SourceExe.FullName -Destination $DestExe -Force
} else {
    Write-Host "Error: Binary '$Binary.exe' not found in zip." -ForegroundColor $Red
    Write-Host "Contents of $TmpDir :"
    Get-ChildItem $TmpDir -Recurse | Select-Object FullName
    exit 1
}

# Agregar al PATH
$UserPath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::User)
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "Adding $InstallDir to PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", [EnvironmentVariableTarget]::User)
    $env:Path += ";$InstallDir"
    Write-Host "PATH updated. You might need to restart your terminal." -ForegroundColor $Green
}

# Limpieza
if (Test-Path $TmpDir) {
    try {
        $OldEAP = $ErrorActionPreference
        $ErrorActionPreference = "SilentlyContinue"
        Remove-Item -Recurse -Force $TmpDir
        $ErrorActionPreference = $OldEAP
    } catch {
        $ErrorActionPreference = $OldEAP
        Write-Host "Warning: Could not remove temp dir '$TmpDir'. You may delete it manually." -ForegroundColor $Yellow
    }
}

Write-Host "Kolyn installed successfully!" -ForegroundColor $Green
Write-Host "Run 'kolyn --help' to get started."
