# 🚀 Manual de Despliegue en AWS EC2 + MySQL + HTTPS

Guía completa para desplegar la API ISmartSell en EC2 con base de datos MySQL local y HTTPS.

## 📋 Requisitos Previos

- Cuenta de AWS
- Dominio o subdominio apuntando a tu EC2 (para HTTPS)
- Ejemplo: `api.tudominio.com`

---

## 1️⃣ Crear y Configurar Instancia EC2

### Paso 1.1: Lanzar EC2
1. Ve a AWS Console → EC2 → Launch Instance
2. Configuración:
   - **Nombre**: `ismartsell-api`
   - **AMI**: Ubuntu Server 22.04 LTS
   - **Tipo**: `t2.small` (mínimo recomendado) o `t2.micro` (gratis)
   - **Key pair**: Crea o selecciona una (descarga el .pem)
   - **Security Group**: Crea nuevo con estas reglas:
     - SSH (22) - Tu IP
     - HTTP (80) - 0.0.0.0/0
     - HTTPS (443) - 0.0.0.0/0
   - **Storage**: 20 GB mínimo

3. Launch Instance

### Paso 1.2: Configurar DNS
1. Copia la IP pública de tu EC2 (ej: `3.145.78.90`)
2. En tu proveedor de dominio, crea registro A:
   ```
   api.tudominio.com → 3.145.78.90
   ```
3. Espera 5-10 minutos para propagación

---

## 2️⃣ Conectar a EC2 e Instalar Dependencias

### Paso 2.1: Conectar por SSH
```bash
chmod 400 tu-key.pem
ssh -i tu-key.pem ubuntu@3.145.78.90
```

### Paso 2.2: Actualizar Sistema
```bash
sudo apt update && sudo apt upgrade -y
```

### Paso 2.3: Instalar Python 3.10+
```bash
sudo apt install -y python3 python3-pip python3-venv
python3 --version  # Debe ser 3.10+
```

### Paso 2.4: Instalar MySQL
```bash
sudo apt install -y mysql-server
sudo systemctl start mysql
sudo systemctl enable mysql
```

### Paso 2.5: Instalar Nginx y Certbot (para HTTPS)
```bash
sudo apt install -y nginx certbot python3-certbot-nginx
sudo systemctl start nginx
sudo systemctl enable nginx
```

### Paso 2.6: Instalar Git
```bash
sudo apt install -y git
```

---

## 3️⃣ Configurar MySQL

### Paso 3.1: Configurar MySQL (sin crear usuarios adicionales)
```bash
sudo mysql
```

Dentro de MySQL ejecuta:
```sql
-- Crear base de datos
CREATE DATABASE ismartsell;

-- Configurar root para usar contraseña (reemplaza 'TuPassword123!' con tu contraseña)
ALTER USER 'root'@'localhost' IDENTIFIED WITH mysql_native_password BY 'TuPassword123!';
FLUSH PRIVILEGES;

-- Verificar
SELECT user, host, plugin FROM mysql.user WHERE user='root';

-- Salir
EXIT;
```

### Paso 3.2: Probar conexión
```bash
mysql -u root -p
# Ingresa tu contraseña
# Si entra, salir con: EXIT;
```

---

## 4️⃣ Desplegar la Aplicación

### Paso 4.1: Clonar o Subir el Código

**Opción A: Clonar desde Git (si tienes repo)**
```bash
cd /home/ubuntu
git clone https://github.com/tuusuario/Api_ISmartSell.git
cd Api_ISmartSell
```

**Opción B: Subir desde tu máquina local**
En tu máquina local:
```bash
cd /home/rodrigo/Documentos
scp -i tu-key.pem -r Api_ISamrtSell ubuntu@3.145.78.90:/home/ubuntu/
```

Luego en EC2:
```bash
cd /home/ubuntu/Api_ISamrtSell
```

### Paso 4.2: Crear Entorno Virtual e Instalar Dependencias
```bash
python3 -m venv venv
source venv/bin/activate
pip install --upgrade pip
pip install -r requirements.txt
```

### Paso 4.3: Configurar Variables de Entorno
```bash
nano .env
```

Pega esto (ajusta los valores):
```env
# Database
DATABASE_URL=mysql+aiomysql://root:TuPassword123!@localhost:3306/ismartsell

# JWT
JWT_SECRET_KEY=genera-un-string-aleatorio-super-largo-aqui-12345678
JWT_ALGORITHM=HS256
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=30

# CORS (permite tu dominio)
CORS_ORIGINS=["https://api.tudominio.com","http://localhost:3000"]

# Reservation
RESERVATION_TIMEOUT_MINUTES=30

# Payment
PAYMENT_PROVIDER_API_KEY=tu-api-key-stripe-o-payu
PLATFORM_COMMISSION_RATE=0.01
```

Guarda con `Ctrl+O`, `Enter`, `Ctrl+X`

### Paso 4.4: Actualizar requirements.txt (agregar MySQL driver)
```bash
nano requirements.txt
```

Agrega esta línea al final:
```
aiomysql==0.2.0
```

Instala:
```bash
pip install aiomysql
```

### Paso 4.5: Crear Tablas de Base de Datos
```bash
# Opción 1: Con Alembic
alembic upgrade head

# Opción 2: Auto-creación (las tablas se crean al iniciar la app)
# No hacer nada, se crean automáticamente
```

### Paso 4.6: Probar la Aplicación
```bash
uvicorn src.main:app --host 0.0.0.0 --port 8000
```

Abre navegador: `http://3.145.78.90:8000/docs`

Si funciona, presiona `Ctrl+C` para detener.

---

## 5️⃣ Configurar Nginx como Reverse Proxy

### Paso 5.1: Crear Configuración de Nginx
```bash
sudo nano /etc/nginx/sites-available/ismartsell
```

Pega esto (reemplaza `api.tudominio.com`):
```nginx
server {
    listen 80;
    server_name api.tudominio.com;

    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

Guarda: `Ctrl+O`, `Enter`, `Ctrl+X`

### Paso 5.2: Activar Configuración
```bash
sudo ln -s /etc/nginx/sites-available/ismartsell /etc/nginx/sites-enabled/
sudo nginx -t  # Verificar configuración
sudo systemctl reload nginx
```

---

## 6️⃣ Configurar HTTPS con Let's Encrypt

### Paso 6.1: Obtener Certificado SSL
```bash
sudo certbot --nginx -d api.tudominio.com
```

Responde:
- Email: `tuemail@example.com`
- Términos: `Y`
- Compartir email: `N` o `Y`
- Redirect HTTP a HTTPS: `2` (Sí, redirect)

Certbot configurará automáticamente Nginx para HTTPS.

### Paso 6.2: Verificar Renovación Automática
```bash
sudo certbot renew --dry-run
```

Si dice "successful", la renovación automática está configurada.

---

## 7️⃣ Configurar Servicio Systemd (Auto-Start)

### Paso 7.1: Crear Servicio
```bash
sudo nano /etc/systemd/system/ismartsell.service
```

Pega esto:
```ini
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
```

Guarda: `Ctrl+O`, `Enter`, `Ctrl+X`

### Paso 7.2: Activar y Ejecutar Servicio
```bash
sudo systemctl daemon-reload
sudo systemctl enable ismartsell
sudo systemctl start ismartsell
sudo systemctl status ismartsell
```

Debe decir "active (running)" en verde.

### Paso 7.3: Ver Logs (si hay problemas)
```bash
sudo journalctl -u ismartsell -f
```

---

## 8️⃣ Verificación Final

### Paso 8.1: Probar la API
Abre navegador:
```
https://api.tudominio.com/docs
```

Deberías ver Swagger UI con candado verde (HTTPS).

### Paso 8.2: Probar Endpoints
1. Registrar usuario:
   - `POST /api/auth/register`
   ```json
   {
     "email": "test@test.com",
     "password": "password123",
     "role": "SELLER"
   }
   ```

2. Copiar el token

3. Click en "Authorize" → Pegar token

4. Crear tienda:
   - `POST /api/stores`

### Paso 8.3: Comandos Útiles
```bash
# Ver logs en tiempo real
sudo journalctl -u ismartsell -f

# Reiniciar servicio
sudo systemctl restart ismartsell

# Detener servicio
sudo systemctl stop ismartsell

# Ver estado
sudo systemctl status ismartsell

# Reiniciar Nginx
sudo systemctl restart nginx

# Ver logs de Nginx
sudo tail -f /var/log/nginx/error.log
```

---

## 🔧 Troubleshooting

### Problema: "Connection refused"
```bash
# Verificar que el servicio esté corriendo
sudo systemctl status ismartsell

# Ver logs
sudo journalctl -u ismartsell -n 50
```

### Problema: "502 Bad Gateway"
```bash
# Verificar que uvicorn esté escuchando en puerto 8000
sudo netstat -tlnp | grep 8000

# Reiniciar servicio
sudo systemctl restart ismartsell
```

### Problema: Error de base de datos
```bash
# Verificar MySQL
sudo systemctl status mysql

# Verificar conexión
mysql -u root -p -e "SHOW DATABASES;"

# Ver logs de la app
sudo journalctl -u ismartsell -n 100
```

### Problema: Certificado SSL no funciona
```bash
# Verificar certificado
sudo certbot certificates

# Renovar manualmente
sudo certbot renew

# Verificar configuración Nginx
sudo nginx -t
```

---

## 📊 Resumen de URLs

- **API Docs**: https://api.tudominio.com/docs
- **Health Check**: https://api.tudominio.com/health
- **ReDoc**: https://api.tudominio.com/redoc

---

## 🎯 Próximos Pasos

1. **Backups**: Configurar backups automáticos de MySQL
   ```bash
   # Backup manual
   mysqldump -u root -p ismartsell > backup.sql
   ```

2. **Monitoring**: Instalar herramientas de monitoreo

3. **Firewall**: Configurar UFW (opcional)
   ```bash
   sudo ufw allow 22
   sudo ufw allow 80
   sudo ufw allow 443
   sudo ufw enable
   ```

4. **Optimizar MySQL**:
   ```bash
   sudo nano /etc/mysql/mysql.conf.d/mysqld.cnf
   # Ajustar innodb_buffer_pool_size según RAM
   ```

---

## ✅ Checklist Final

- [ ] EC2 creada y configurada
- [ ] DNS apuntando a EC2
- [ ] MySQL instalado y configurado
- [ ] Código subido a EC2
- [ ] Dependencias instaladas
- [ ] .env configurado
- [ ] Base de datos creada
- [ ] Nginx configurado
- [ ] HTTPS configurado con Let's Encrypt
- [ ] Servicio systemd funcionando
- [ ] API accesible en https://api.tudominio.com/docs
- [ ] Endpoints probados

¡Tu API está en producción! 🚀
