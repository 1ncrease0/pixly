
### Сборка

```bash
docker build -f services/auth/Dockerfile -t pixly-auth .
docker build -f services/gateway/Dockerfile -t pixly-gateway .
docker build -f services/notification/Dockerfile -t pixly-notification .
```

### Запуск зависимостей

```bash
docker-compose up -d
```

### Запуск каждого сервиса

**Auth**

```bash
docker run --rm -it \
  --network host \
  -v "$(pwd)/services/auth/local.yaml:/app/config.yaml:ro" \
  -e CONFIG_PATH=/app/config.yaml \
  pixly-auth
```

**Gateway**

```bash
docker run --rm -it \
  --network host \
  -v "$(pwd)/services/gateway/local.yaml:/app/config.yaml:ro" \
  -e CONFIG_PATH=/app/config.yaml \
  pixly-gateway
```

**Notification**

```bash
docker run --rm -it \
  --network host \
  -v "$(pwd)/services/notification/local.yaml:/app/config.yaml:ro" \
  -v "$(pwd)/services/notification/.env:/app/.env:ro" \
  -e CONFIG_PATH=/app/config.yaml \
  pixly-notification
```

TODO:
- Добавить полный запуск через docker compose
- Исправить логгирование
- Дописать art service
- Добавить тесты
- Добавить линтер в пайплайн, исправить код с предупреждениями линтера


