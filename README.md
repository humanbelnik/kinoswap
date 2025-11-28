# Kinoswap

## ТиОПО ЛР №7 (Mock-сервер)

Мок-сервер
```
docker compose -f docker-compose.prod.yml up -d mock-s3-server
```

Инстанс приложения (с использованием мока и без)
```
S3_CLIENT_TYPE=mock docker compose -f docker-compose.prod.yml up -d core-migrate core-app --build
```

```
S3_CLIENT_TYPE=real docker compose -f docker-compose.prod.yml up -d core-migrate core-app --build
```

Тесты
```
env=CI docker compose -f docker-compose.prod.yml up e2e-tests --build
```

Или локально
```
cd services/e2e
go run main.go
```

Протестировать:
- [x] Поднять реальный S3, прогнать тест
- [x] Поднять мок S3, прогнать тест
- [x] Сконфигурировать приложение на работу с моком, мок не поднимать. Словить ошибку запуска
- [ ] В `Pipeline` добавить прогон теста с моком и без мока






