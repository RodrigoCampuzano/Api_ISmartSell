#!/usr/bin/env bash
# ============================================================
#  deploy_update.sh — Actualiza la API en el servidor
#  Uso: ssh root@tu-servidor 'bash -s' < deploy_update.sh
#        o bien copiarlo al servidor y ejecutarlo:
#        chmod +x deploy_update.sh && sudo ./deploy_update.sh
# ============================================================

set -euo pipefail

# ── Configuración ─────────────────────────────────────────────
APP_DIR="/var/www/Api_ISmartSell"
SERVICE_NAME="pos-api"
DB_NAME="pos_app"
DB_USER="postgres"
BINARY_NAME="pos-api"
GIT_BRANCH="main"   # Cambia si usas otra rama

# ── Colores ───────────────────────────────────────────────────
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

info()  { echo -e "${GREEN}[✓]${NC} $1"; }
warn()  { echo -e "${YELLOW}[!]${NC} $1"; }
fail()  { echo -e "${RED}[✗]${NC} $1"; exit 1; }

# ── 1. Pull del código ───────────────────────────────────────
info "Actualizando código desde GitHub..."
cd "$APP_DIR" || fail "No existe el directorio $APP_DIR"

# Permitir a Git ejecutarse como root en un directorio no propio
git config --global --add safe.directory "$APP_DIR"

git fetch origin
git reset --hard "origin/$GIT_BRANCH"
info "Código actualizado a la última versión de '$GIT_BRANCH'."

# ── 2. Migración de la Base de Datos ─────────────────────────
MIGRATION_FILE="$APP_DIR/migrations/001_mercadopago.sql"
if [ -f "$MIGRATION_FILE" ]; then
    info "Ejecutando migración de Mercado Pago..."
    sudo -u "$DB_USER" psql -d "$DB_NAME" -f "$MIGRATION_FILE" 2>&1 || warn "Migración ya aplicada o con advertencias (normal si se corre dos veces)."
    info "Migración completada."
else
    warn "No se encontró archivo de migración en $MIGRATION_FILE, saltando..."
fi

# ── 3. Compilar el binario ────────────────────────────────────
info "Compilando la API..."
cd "$APP_DIR"
go build -o "$BINARY_NAME" ./main.go || fail "Error de compilación"
info "Binario '$BINARY_NAME' compilado exitosamente."

# Restaurar permisos al usuario 'ubuntu' para evitar problemas con git pull futuros
chown -R ubuntu:ubuntu "$APP_DIR"

# ── 4. Reiniciar el servicio ──────────────────────────────────
info "Reiniciando servicio $SERVICE_NAME..."
systemctl daemon-reload
systemctl restart "$SERVICE_NAME" || fail "No se pudo reiniciar $SERVICE_NAME"

# Esperar 2 segundos y verificar
sleep 2
if systemctl is-active --quiet "$SERVICE_NAME"; then
    info "¡Servicio $SERVICE_NAME corriendo correctamente!"
else
    fail "El servicio $SERVICE_NAME no arrancó. Revisa: journalctl -u $SERVICE_NAME -n 50"
fi

# ── 5. Verificación rápida ────────────────────────────────────
info "Verificando respuesta de la API..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/auth/login 2>/dev/null || echo "000")
if [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "200" ]; then
    info "API respondiendo correctamente (HTTP $HTTP_CODE)."
else
    warn "API respondió con HTTP $HTTP_CODE — revisa los logs."
fi

echo ""
echo -e "${GREEN}═══════════════════════════════════════════${NC}"
echo -e "${GREEN}   ¡Deploy completado exitosamente! 🚀${NC}"
echo -e "${GREEN}═══════════════════════════════════════════${NC}"
echo ""
echo "  Comandos útiles:"
echo "    sudo journalctl -u $SERVICE_NAME -f      # Ver logs"
echo "    sudo systemctl status $SERVICE_NAME       # Ver estado"
echo ""
