//TODO возможно потыкать k8s и добавить все таки сюда какие минимальные манифесты
//TODO написать более подробные ридми для каждого микросервиса

//TODO написать про CI пайпланы
//TODO возможно добавить отдельный параграф о гарантиях доставки, архитектурных решениях
//TODO посмотреть как делаются доки на эндпоинты и прикрепить сюда
//TODO влепить сюда ER диаграммы обоих постгресов 

# Emergency Notification System

//TODO разобраться с бейджами

[![CI Status](https://img.shields.io/github/actions/workflow/status/SteeperMold/Emergency-Notification-System/e2e.yaml?branch=main)](https://github.com/your-username/notification-system/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/SteeperMold/Emergency-Notification-System)](https://goreportcard.com/report/github.com/your-username/notification-system)
[![Coverage Status](https://coveralls.io/repos/github/SteeperMold/Emergency-Notification-System/badge.svg?branch=main)](https://coveralls.io/github/your-username/notification-system?branch=main)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Содержание

//TODO не забыть в самом конце обновить содержание!!! а может и выпилить его вообще...

* [Обзор](#обзор)
* [Функциональность](#функциональность)
* [Архитектура](#архитектура)
* [Технологический стек](#технологический-стек)
* [Установка и запуск](#установка-и-запуск)
    * [Требования](#требования)
    * [Клонирование и настройка](#клонирование-и-настройка)
    * [Запуск локально](#запуск-локально)
* [Использование](#использование)
* [Структура проекта](#структура-проекта)
* [Тестирование](#тестирование)
* [Участие в развитии](#участие-в-развитии)
* [История изменений](#история-изменений)
* [Лицензия](#лицензия)

## Обзор

**Emergency Notification System** — масштабируемая микросервисная платформа для рассылки SMS-уведомлений большому
количеству контактов с гарантией доставки. Пользователь может загружать контакты из CSV/XLSX-файлов или добавлять
вручную, создавать шаблоны сообщений и отправлять уведомления миллионам получателей одним кликом. Надёжная логика
повторных попыток и сервис балансировки гарантируют, что ни одно сообщение не будет потеряно.

## Архитектура

![Диаграмма архитектуры](docs/architecture.png)

//TODO подробнее написать про кафку и про то у кого какой постгрес

1. **API Service**: Обработка пользовательских запросов, хранение контактов/шаблонов в Postgres, публикация задач в
   Kafka.
2. **Contacts Worker**: Парсинг CSV/XLSX, обработка чанками с помощью goroutines, пакетная запись контактов в БД. //TODO написать умнее
3. **Notification Service**: Чтение задач уведомлений, запись статуса по каждому получателю, создание задач для Sender
   Service, обработка callback’ов от Twilio. //TODO написать умнее
4. **Sender Service**: Отправка запросов к API Twilio, планирование повторных попыток при неудаче.
5. **Rebalancer Service**: Поиск неотправленных нотификаций и повторная публикация в Kafka.

## Технологический стек

//TODO посмотреть как оформляют стек в других проектах, возможно убрать вот эти пояснения

* **Язык**: Go
* **Фреймворки**: net/http, gorilla/mux
* **Брокер-сообщений**: Kafka
* **Базы данных**: PostgreSQL
* **Очереди**: Отдельные топики Kafka для каждого сервиса
* **Тестирование**: Unit-тесты на Go, интеграционные тесты с Testcontainers
* **CI/CD**: GitHub Actions
* **Контейнеризация**: Docker, Docker Compose

## Установка и запуск

### Требования

* Go 1.24.3
* Docker & Docker Compose
* GNU Make
* (Опционально) Kubernetes & kubectl для тестирования манифестов //TODO а будет ли k8s???

### Клонирование и настройка

//TODO переписать вот эти команды с использованием Makefile

```bash
# Клонировать репозиторий
git clone https://github.com/SteeperMold/Emergency-Notification-System
cd Emergency-Notification-System

# Скопировать шаблоны переменных среды\cp .env.example .env
cp services/api-service/.env.example services/api-service/.env
cp services/contacts-worker/.env.example services/contacts-worker/.env
# ... повторить для остальных сервисов
```

### Запуск локально

Запустить инфраструктуру и все микросервисы:

```bash
# Из корня проекта
env $(cat .env | xargs) make dev

# или через Docker Compose
docker-compose up --build
```

Каждый сервис также имеет собственный `docker-compose.yml` для изолированного тестирования.

## Использование

//TODO изменить кёрлы, и посмотреть как оставить коллекцию с постмана

### Загрузка контактов

```bash
curl -X POST \
  -F "file=@contacts.xlsx" \
  http://localhost:8080/api/contacts/upload
```

### Создание шаблона

```bash
curl -X POST http://localhost:8080/api/templates \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alert","body":"Здравствуйте {{.Name}}, это экстренное уведомление."}'
```

### Отправка уведомления

```bash
curl -X POST http://localhost:8080/api/notifications \
  -H 'Content-Type: application/json' \
  -d '{"template_id":1, "contact_list_id":1}'
```

## Структура проекта

//TODO подумать, а нужно ли это вообще??

```text
.
├── .github/workflows      # CI/CD пайплайны
├── services/
│   ├── api-service/      # Код и конфигурация API сервиса
│   ├── contacts-worker/  # Микросервис обработки контактов
│   ├── notification/     # Сервис управления уведомлениями
│   ├── sender/           # Сервис отправки через Twilio
│   └── rebalancer/       # Сервис ребалансировки задач
├── docs/                 # Диаграммы архитектуры, ER-диаграммы, Swagger spec
├── db/                   # SQL-скрипты и миграции
├── docker-compose.yml    # Корневой compose (Kafka, Zookeeper, Postgres + сервисы)
├── .env.example          # Шаблон переменных среды
├── Makefile              # `make dev`, `make test`, `make clean`
├── README.md             # <-- этот файл
└── LICENSE
```

## Тестирование

//TODO и решить нужно ли вот это? может достаточно несколько строчек описания и бейджа с покрытием

* **Unit-тесты**: `make test` или `go test ./services/... -cover`
* **Интеграционные тесты**: Docker Compose + Testcontainers для Postgres

## История изменений

//TODO посмотреть че это вообще за фигня и можно ли это нарулить постфактум

Смотрите [CHANGELOG.md](CHANGELOG.md) для подробного списка версий и изменений.

## Лицензия

//TODO и нужна ли вообще тут лицензия тоже вопрос

Проект лицензирован под MIT License. Подробнее в файле [LICENSE](LICENSE).
