#!/bin/bash
echo "Reemplazando versión antigua de kolyn..."

# Compilar
go build -o kolyn

# Detectar dónde está el kolyn actual
CURRENT_LOC=$(which kolyn)

if [ -z "$CURRENT_LOC" ]; then
    echo "No se encontró instalación previa. Instalando en /usr/local/bin..."
    TARGET="/usr/local/bin/kolyn"
else
    echo "Instalación detectada en: $CURRENT_LOC"
    TARGET="$CURRENT_LOC"
fi

echo "Moviendo binario a $TARGET (requiere contraseña de administrador)..."
sudo mv kolyn "$TARGET"

echo "✅ Actualización completada."
echo "Versión instalada:"
"$TARGET" version
