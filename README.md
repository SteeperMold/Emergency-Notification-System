# Emergency Notification System

[![CI Status](https://img.shields.io/github/actions/workflow/status/SteeperMold/Emergency-Notification-System/ci.yaml?branch=main)](https://github.com/SteeperMold/Emergency-Notification-System/actions)
[![API Service](https://goreportcard.com/badge/github.com/SteeperMold/Emergency-Notification-System/services/apiservice)](https://goreportcard.com/report/github.com/SteeperMold/Emergency-Notification-System/services/apiservice)
[![Coverage Status](https://codecov.io/gh/SteeperMold/Emergency-Notification-System/branch/main/graph/badge.svg?style=flat-square)](https://codecov.io/gh/SteeperMold/Emergency-Notification-System)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Обзор

**Emergency Notification System** - это платформа для рассылки SMS-уведомлений большому количеству контактов,
гарантирующая доставку каждого сообщения. Пользователь может загружать контакты из CSV или XLSX файлов, а также
добавлять их вручную, создавать шаблоны сообщений и отправлять уведомления миллионам получателей в один клик.

## Архитектура

![Схема архитектуры](docs/architecture.jpg)

1. **Frontend (React)**: Пользовательский интерфейс для работы с системой: управление контактами, шаблонами и
   отправка нотификаций.

2. **API Service**: Обрабатывает пользовательские запросы, хранит контакты/шаблоны в Postgres, публикует задачи в
   Kafka, сохраняет CSV/XLSX файлы с контактами в S3.

3. **Contacts Worker**: Получает CSV/XLSX файлы из S3, параллельно обрабатывает и валидирует, и записывает контакты
   в Postgres.

4. **Notification Service**: Читает задачи на отправку нотификаций, записывает статус доставки сообщения каждому
   получателю в Postgres, создает и публикует задачи для Sender Service, обрабатывает callback’и от Twilio для
   подтверждения доставки пользователю.

5. **Sender Service**: Отправляет запросы к API Twilio, планирует повторные попытки при неудаче.

6. **Rebalancer Service**: Ищет неотправленные нотификации и повторно публикует их в Kafka.

Backend-сервисы находятся по пути [./services](./services)

* [ER-диаграмма API Service DB](docs/er/er_api_service.jpeg)
* [ER-диаграмма Notification Service DB](docs/er/er_notification_service.jpeg)

## Надёжность и масштабируемость

- **Гарантии доставки**:  
  Для обеспечения надежной доставки используется Kafka и PostgreSQL для записи статусов доставки.
  Каждое сообщение Kafka обрабатывается с гарантиями (`at-least-once delivery`), из-за чего сообщения не будут
  потеряны при передаче между микросервисами. Notification Service отслеживает статусы доставки и обрабатывает
  callback’и от Twilio, чтобы точно убедиться, что нотификации были доставлены, либо запланировать повторную отправку.


- **Масштабируемость**:  
  Архитектура построена на микросервисах, которые можно горизонтально масштабировать.  
  Contacts Worker способен параллельно обрабатывать большие CSV/XLSX-файлы с миллионами контактов.  
  Система протестирована на нагрузке **1,000,000 получателей** в рамках одной нотификации. Подробнее об этом можно
  прочитать в разделе [тестирование](#тестирование)

## Технологии

* Go
* Gorilla Mux
* Kafka
* PostgreSQL
* GitHub Actions for CI
* Docker & Docker Compose

## Установка и запуск

### Требования

* Docker & Docker Compose
* GNU Make

### Клонирование и настройка

Клонирование репозитория:

```bash
git clone https://github.com/SteeperMold/Emergency-Notification-System
cd Emergency-Notification-System
```

Создание .env файлов со значениями по умолчанию:

```bash
make prepare-env
```

### Запуск локально

Запустить инфраструктуру и все микросервисы с использованием Docker Compose:

```bash
make run-dev
```

При этом запустится [React-фронтенд](http://localhost:3000), который можно использовать для удобного взаимодействия
с системой. Также, вы можете посмотреть метрики и логи в [Grafana](http://localhost:3001). По умолчанию фронтенд
запускается на 3000 порту, Grafana на 3001.

## Использование API

👉 [Скачать коллекцию Postman со всеми эндпоинтами и их описаниями](docs/ENS.postman_collection.json)

### Примеры curl

#### Аутентификация

```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"123456789admin"}'
```

Для работы с API необходимо получить JWT-токен.  
После успешной аутентификации сервер отправит ответ в таком формате:

```json
{
  "user": {
    "id": 1,
    "email": "test@test.com",
    "creationTime": "2025-08-23T11:05:56.102572Z"
  },
  "accessToken": "<токен>",
  "refreshToken": "<токен>"
}
```

Сохраните `accessToken` - он понадобится для авторизации в последующих запросах.

#### Создать контакт

Учтите, что номер телефона должен быть в международном формате.

```bash
curl -X POST http://localhost:8080/contacts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"name":"Test Contact Name","phone":"89123456789"}'
```

#### Создать шаблон нотификации

```bash
curl -X POST http://localhost:8080/templates \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"name":"Test Template Name","body":"Это тестовое уведомление."}'
```

#### Отправить нотификацию всем контактам

```bash
curl -X POST http://localhost:8080/send-notification/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>"
```

При запуске в `development` режиме, в папке [./services/sender-service/tmp/sms-dev](./services/sender-service/tmp/sms-dev) 
(если её нет, она создастся автоматически) появятся текстовые файлы со всеми нотификациями. При запуске в `production`
режиме, будут сделаны запросы к Twilio API.

## Тестирование

Система покрыта unit-, integration- и E2E-тестами для обеспечения надежности и корректной работы. Их можно запустить с
помощью команды `make all`, а также они автоматически срабатывают в рамках CI-пайплайна Github Actions.

### Нагрузочные E2E-тесты

Для демонстрации гарантии доставки система поддерживает нагрузочный сценарий с нотификацией на миллион контактов.
Пример запуска теста: 

```bash
# Создание .env файлов с настройками по умолчанию
make prepare-env

# Запуск E2E-теста с большим файлом контактов
make e2e-test-load
```

При этом режим работы сервисов поменяется на `test` автоматически. В этом режиме нотификации не будут отправлены 
в Twilio (так как это слишком дорого для нагрузочных тестов) и не будут записаны в файлы (так как файловая система 
сильно замедляется при таком объеме). В остальном всё будет работать так же, как и с реальным Twilio: статусы доставки 
отслеживаются, callback'и от Twilio симулируются, при попытке отправки нотификации симулированные сетевые ошибки 
срабатывают с вероятностью 5%, а ошибки доставки Twilio — с вероятностью 20%. Это необходимо для гарантирования
доставки даже при ошибках инфраструктуры.

> Во время выполнения нагрузочного теста проверяется корректная обработка CSV/XLSX, запись контактов в базу, публикация
> задач в Kafka и доставка уведомлений через Sender Service.

## Лицензия

Проект лицензирован под MIT License. Подробнее в файле [LICENSE](LICENSE).
