#!/bin/bash
# Script de despliegue rápido para EC2
# Ejecutar como: sudo bash deploy.sh

set -e

echo "🚀 Iniciando despliegue de ISmartSell API..."

# Colores
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Variables
DOMAIN=${1:-"api.tudominio.com"}
DB_PASSWORD=${2:-"Password123!"}

echo -e "${YELLOW}Dominio: $DOMAIN${NC}"
echo -e "${YELLOW}MySQL Password: $DB_PASSWORD${NC}"

# 1. Actualizar sistema
echo -e "${GREEN}[1/9] Actualizando sistema...${NC}"
apt update && apt upgrade -y

# 2. Instalar dependencias
echo -e "${GREEN}[2/9] Instalando dependencias...${NC}"
apt install -y python3 python3-pip python3-venv git mysql-server nginx certbot python3-certbot-nginx

# 3. Configurar MySQL
echo -e "${GREEN}[3/9] Configurando MySQL...${NC}"
systemctl start mysql
systemctl enable mysql

mysql -e "CREATE DATABASE IF NOT EXISTS ismartsell;"
mysql -e "ALTER USER 'root'@'localhost' IDENTIFIED WITH mysql_native_password BY '$DB_PASSWORD';"
mysql -e "FLUSH PRIVILEGES;"

# 4. Clonar o verificar código
echo -e "${GREEN}[4/9] Verificando código...${NC}"
cd /home/ubuntu
if [ ! -d "Api_ISamrtSell" ]; then
    echo "Por favor, sube el código manualmente con scp"
    exit 1
fi

# 5. Instalar dependencias Python
echo -e "${GREEN}[5/9] Instalando dependencias Python...${NC}"
cd Api_ISamrtSell
python3 -m venv venv
source venv/bin/activate
pip install --upgrade pip
pip install -r requirements.txt

# 6. Configurar .env
echo -e "${GREEN}[6/9] Configurando .env...${NC}"
cat > .env <<EOF
DATABASE_URL=mysql+aiomysql://root:$DB_PASSWORD@localhost:3306/ismartsell
JWT_SECRET_KEY=$(openssl rand -hex 32)
JWT_ALGORITHM=HS256
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=30
CORS_ORIGINS=["https://$DOMAIN","http://localhost:3000"]
RESERVATION_TIMEOUT_MINUTES=30
PAYMENT_PROVIDER_API_KEY=your-key-here
PLATFORM_COMMISSION_RATE=0.01
EOF

# 7. Configurar Nginx
echo -e "${GREEN}[7/9] Configurando Nginx...${NC}"
cat > /etc/nginx/sites-available/ismartsell <<EOF
server {
    listen 80;
    server_name $DOMAIN;

    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
    }
}
EOF

ln -sf /etc/nginx/sites-available/ismartsell /etc/nginx/sites-enabled/
nginx -t && systemctl reload nginx

# 8. Configurar systemd
echo -e "${GREEN}[8/9] Configurando servicio systemd...${NC}"
cat > /etc/systemd/system/ismartsell.service <<EOF
[Unit]
Description=ISmartSell FastAPI Application
After=network.target mysql.service

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/Api_ISamrtSell
Environment="PATH=/home/ubuntu/Api_ISamrtSell/venv/bin"
ExecStart=/home/ubuntu/Api_ISamrtSell/venv/bin/uvicorn src.main:app --host 127.0.0.1 --port 8000 --workers 2
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable ismartsell
systemctl start ismartsell

# 9. Configurar HTTPS
echo -e "${GREEN}[9/9] Configurando HTTPS...${NC}"
echo "Ejecuta manualmente: sudo certbot --nginx -d $DOMAIN"

echo -e "${GREEN}✅ Despliegue completado!${NC}"
echo -e "${YELLOW}Próximo paso: Ejecuta 'sudo certbot --nginx -d $DOMAIN' para HTTPS${NC}"
echo -e "${YELLOW}Luego visita: https://$DOMAIN/docs${NC}"
