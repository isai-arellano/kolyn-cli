# Configuración
$Repo = "isai-arellano/kolyn-cli"
$Binary = "kolyn"
$InstallDir = "$env:USERPROFILE\bin"

# Colores (aproximación para PowerShell)
$Blue = [ConsoleColor]::Blue
$Green = [ConsoleColor]::Green
$Red = [ConsoleColor]::Red
$Reset = [ConsoleColor]::White

Write-Host "Installing Kolyn CLI..." -ForegroundColor $Blue

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
$TmpDir = Join-Path $env:TEMP "kolyn_install"
if (Test-Path $TmpDir) { Remove-Item -Recurse -Force $TmpDir }
New-Item -ItemType Directory -Force -Path $TmpDir | Out-Null

# Descargar
Write-Host "Downloading from $LatestUrl..."
$ZipPath = Join-Path $TmpDir $Filename
try {
    Invoke-WebRequest -Uri $LatestUrl -OutFile $ZipPath
} catch {
    Write-Host "Error downloading release. Check if release exists for this architecture." -ForegroundColor $Red
    exit 1
}

# Extraer
Write-Host "Extracting..."
Expand-Archive -Path $ZipPath -DestinationPath $TmpDir -Force

# Crear directorio de instalación si no existe
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

# Instalar (Mover binario)
$SourceExe = Get-ChildItem -Path $TmpDir -Recurse -Filter "$Binary.exe" | Select-Object -First 1

if ($null -ne $SourceExe) {
    Write-Host "Installing to $DestExe..."
    
    # Intentar detener procesos que usen el archivo antes de sobrescribir
    if (Get-Process $Binary -ErrorAction SilentlyContinue) {
        Stop-Process -Name $Binary -Force -ErrorAction SilentlyContinue
    }
    
    Move-Item -Path $SourceExe.FullName -Destination $DestExe -Force
} else {
    Write-Host "Error: Binary not found in zip." -ForegroundColor $Red
    exit 1
}

# Agregar al PATH si no está
$UserPath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::User)
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "Adding $InstallDir to PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", [EnvironmentVariableTarget]::User)
    $env:Path += ";$InstallDir"
    Write-Host "PATH updated. You might need to restart your terminal." -ForegroundColor $Green
}

# Limpieza
Remove-Item -Recurse -Force $TmpDir

Write-Host "Kolyn installed successfully!" -ForegroundColor $Green
Write-Host "Run 'kolyn --help' to get started."
