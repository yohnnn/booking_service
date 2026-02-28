# Concert Booking Service

Сервис бронирования билетов на концерты.

REST API сервис по бронированию с Kafka для отправки уведомлений о бронировании, Redis для кэширования списка концертов, JWT для авторизации.


## Технический стек

*   Язык: Go 1.25
*   База данных: PostgreSQL 
*   Кэширование: Redis 
*   Messaging: Apache Kafka + Zookeeper
*   Роутер: Chi 
*   Миграции: Goose
*   Тестирование: gomock + testify
*   Линтер: golangci-lint v2
*   Контейнеризация: Docker + Docker Compose

## Как запустить
Для работы сервиса необходимы установленные Docker и Docker Compose.

1. Склонируйте репозиторий и перейдите в директорию проекта:
```bash
git clone https://github.com/yohnnn/booking_service.git
cd booking_service
```

2. Создайте `.env` файл на основе шаблона:
```bash
cp .env.template .env
```

3. Запустите проект командой:
```bash
make docker-up
```

После успешного запуска будут доступны:
*   API Сервиса: http://localhost:8080
*   Kafka (внешний): localhost:29092
*   PostgreSQL: localhost:5432
*   Redis: localhost:6379

## Команды Makefile

В проекте предусмотрен Makefile для удобства разработки:

| Команда | Описание |
| :--- | :--- |
| `make build` | Собрать бинарники сервера и consumer-а |
| `make docker-up` | Поднять все сервисы в Docker (с миграциями) |
| `make docker-down` | Остановить и удалить контейнеры |
| `make consumer` | Запустить Kafka consumer для обработки событий бронирования |
| `make test` | Запустить юнит-тесты сервисного слоя |
| `make lint` | Запустить golangci-lint |

## API Эндпоинты

### Authentication
| Метод | Путь | Описание |
| :--- | :--- | :--- |
| POST | `/api/auth/register` | Регистрация нового пользователя |
| POST | `/api/auth/login` | Авторизация (получение JWT токена) |

### Concerts
| Метод | Путь | Описание | Авторизация |
| :--- | :--- | :--- | :--- |
| GET | `/api/concerts` | Получить список всех концертов | Нет |
| GET | `/api/concerts/{id}` | Получить информацию о конкретном концерте | Нет |

### Bookings
| Метод | Путь | Описание | Авторизация |
| :--- | :--- | :--- | :--- |
| POST | `/api/bookings` | Создать бронирование (отправляет событие в Kafka) | Да |
| GET | `/api/bookings` | Получить все бронирования пользователя | Да |

## Структура проекта
```
├── cmd
│   ├── consumer         
│   └── server           
├── internal
│   ├── app              
│   ├── cache            
│   ├── config           
│   ├── dto              
│   ├── event            
│   ├── handler          
│   ├── middleware       
│   ├── models           
│   ├── repository       
│   └── service          
├── migrations           
├── .golangci.yml        
└── docker-compose.yml  
```

## Тестирование

Юнит-тесты покрывают сервисный слой. Зависимости (репозитории, кэш, Kafka producer, менеджер транзакций) мокаются через `gomock`.

```bash
make test
```

Покрытые сценарии:
*   **AuthService** — регистрация, логин, парсинг JWT 
*   **ConcertService** — получение из кэша, cache miss с fallback на БД, ошибки
*   **BookingService** — успешная бронь в транзакции, нет мест, дубликат, ошибки репозитория и транзакции

## Примеры использования

### 1. Регистрация пользователя
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123"
  }'
```

### 2. Авторизация
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123"
  }'
```

Ответ:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 3. Получить список концертов
```bash
curl http://localhost:8080/api/concerts
```

### 4. Создать бронирование
```bash
TOKEN="your_jwt_token_here"

curl -X POST http://localhost:8080/api/bookings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "concert_id": "a34d04fb-290b-4485-8150-21a194666fb9",
    "seat_number": 42
  }'
```

### 5. Запуск Kafka Consumer
В отдельном терминале:
```bash
make consumer
```

Вы увидите логи обработки событий:
```
2026/01/29 15:30:42 Received booking event: booking_id=123...
2026/01/29 15:30:42 📧 Sending email notification for booking 123...
```


