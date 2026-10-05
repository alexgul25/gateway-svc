# :globe_with_meridians: Gateway Service

Микросервис для проекта **Date Wishlist Hub**.

Ссылка на центральный репозиторий проекта: **[Date Wishlist Hub Deploy](https://github.com/alexgul25/date-wishlist-hub-deploy)**

Ссылка на канбан-доску проекта: **[Date Wishlist Hub - Development](https://github.com/users/alexgul25/projects/2)**

*Стек технологий сервиса:* `Go`  `HTTP`  `gRPC`

## :bulb: Описание сервиса

**Gateway Service** - внешний HTTP-сервер, предоставляющий публичные эндпоинты для всех функций проекта (единая точка входа для пользователей).

- Получает запрос от пользователя и через gRPC-клиент вызывает нужный метод **[User Service](https://github.com/alexgul25/user-svc)** или **[Place Service](https://github.com/alexgul25/place-svc)** (см. [архитектуру проекта](https://github.com/alexgul25/date-wishlist-hub-deploy#building_construction-архитектура-проекта)).

- Protobuf-контракты определены публично в **[Protos](https://github.com/alexgul25/protos)**.

- Не использует БД, за хранение данных отвечают gRPC-серверы.

***Таблица REST API***

| Method | Endpoint                        | Auth | Request Body | Response Body | Info                                              |
| :----: | ------------------------------- | :--: | :----------: | :-----------: | ------------------------------------------------- |
| POST   | /api/users                      | ❌   | ✅           | ✅            | Регистрация пользователя                          |
| POST   | /api/auth/login                 | ❌   | ✅           | ✅            | Аутентификация пользователя                       |
| GET    | /api/users/me                   | ✅   | ❌           | ✅            | Получение данных своего профиля                   |
| GET    | /api/users?search_query=name    | ✅   | ❌           | ✅            | Поиск пользователей по имени                      |
| POST   | /api/subscriptions              | ✅   | ✅           | ❌            | Подписка на другого пользователя                  |
| DELETE | /api/subscriptions/{followeeID} | ✅   | ❌           | ❌            | Отписка от другого пользователя                   |
| GET    | /api/users/me/followers         | ✅   | ❌           | ✅            | Получение списка своих подписчиков                |
| GET    | /api/users/{userID}/followers   | ✅   | ❌           | ✅            | Получение списка подписчиков по ID                |
| POST   | /api/places                     | ✅   | ✅           | ✅            | Добавление нового места в свой список мест        |
| GET    | /api/users/me/places            | ✅   | ❌           | ✅            | Получение своего списка мест                      |
| GET    | /api/users/{userID}/places      | ✅   | ❌           | ✅            | Получения списка мест пользователя по ID          |

<!-- markdownlint-disable MD033 -->
<details>
<summary>Примечания к таблице</summary>

- Столбец `Auth` указывает: ✅ или ❌ - соответственно нужен или нет Auth Header вида `Authorization: Bearer <token>` (токен выдаётся сервером при успешной аутентификации)

- Подробности о структурах `Request Body` и `Response Body` для каждого HTTP-метода можно найти в папке [./internal/dto/](./internal/dto/).

</details>
<!-- markdownlint-enable MD033 -->

## :gear: Структура сервиса

:open_file_folder: **[./cmd](./cmd/)** - команды для запуска приложения и проверки состояния.

:open_file_folder: **[./internal/app/http](./internal/app/http/)** - сборка HTTP-сервера.

:open_file_folder: **[./internal/clients/grpc](./internal/clients/grpc/)** - абстракции для gRPC-клиентов, определённых в публичных Protobuf-контрактах.

:open_file_folder: **[./internal/config](./internal/config/)** - загрузка файлов конфигурации.

:open_file_folder: **[./internal/dto](./internal/dto/)** - определение объектов передачи данных.

:open_file_folder: **[./internal/http/handlers](./internal/http/handlers/)** - определение **HTTP-хендлеров**.

:open_file_folder: **[./internal/http/handlerutil](./internal/http/handlerutil/)** - общие функции для HTTP-хендлеров.

:open_file_folder: **[./internal/http/routing](./internal/http/routing/)** - константы для определения эндпоинтов.

:open_file_folder: **[./internal/lib](./internal/lib/)** - общие вспомогательные утилиты и функции.

:open_file_folder: **[./internal/models](./internal/models/)** - модели для абстракции от данных, определённых в Protobuf-контрактах и возвращаемых gRPC-методами.

## :desktop_computer: Локальный запуск и работа через терминал

В данном разделе приведена инструкция по запуску одного **Gateway Service**. Сервис можно запустить двумя способами: собрать и запустить бинарник локально или запустить в Docker-контейнере.

Инструкция по запуску всего проекта целиком доступна по **[ссылке](https://github.com/alexgul25/date-wishlist-hub-deploy#desktop_computer-локальный-запуск-и-работа-через-терминал)**.

### 1. Подготовка окружения

В вашем дистрибутиве должны быть установлены и готовы к работе:

- актуальная для проекта версия Go (см. [go.mod](./go.mod)) - для локального запуска;
- Docker Engine с плагином `buildx` - для запуска в контейнере;
- утилита `curl`;
- утилита `jq`;
- утилита `make`.

### 2. Клонирование репозитория

Клонируйте репозиторий c помощью HTTP или SSH

```bash
git clone https://github.com/alexgul25/gateway-svc.git
```

```bash
git clone git@github.com:alexgul25/gateway-svc.git
```

***ВАЖНО!*** Создайте в корневой папке репозитория файл `.env` для переменных окружения и заполните его (см [.env.example](.env.example)).

<!-- markdownlint-disable MD033 -->
<details>
<summary>Особенности .env при запуске в Docker</summary>

- `SERVER_ADDR` указывайте в формате `:<порт>` (например, `:8082`). Адрес вида `localhost:<порт>` внутри контейнера означает сам контейнер, и сервер будет недоступен снаружи.
- Значения указывайте без кавычек: Docker передаёт их в контейнер как есть, вместе с кавычками.
- `localhost` в `USER_SERVICE_ADDR` и `PLACE_SERVICE_ADDR` внутри контейнера тоже означает сам контейнер, а не вашу машину (см. подсказки к варианту запуска в Docker в [следующем шаге](#3-запуск-и-работа)).

</details>
<!-- markdownlint-enable MD033 -->

### 3. Запуск и работа

<!-- markdownlint-disable MD033 -->
<details>
<summary>Примечание</summary>

HTTP-сервер является шлюзом и ожидает, что gRPC-сервисы (User Service, Place Service) уже запущены. Без них запросы будут возвращать `internal server error`. Инструкция для запуска всей системы приведена в [начале раздела](#desktop_computer-локальный-запуск-и-работа-через-терминал).

</details>
<!-- markdownlint-enable MD033 -->
#### Вариант 1. Локальный запуск

Для удобства локальной работы определён Makefile.

1. `make help` - узнайте о доступных командах.
2. `make run` - соберите бинарник и запустите HTTP-сервер.
3. `CTRL + C` - пошлите серверу сигнал завершения, когда закончите работу.

#### Вариант 2. Запуск в Docker

В корне репозитория определены **[Dockerfile](./Dockerfile)** и **[.dockerignore](./.dockerignore)**. Итоговый образ содержит только собранный бинарник, переменные окружения передаются в контейнер при запуске.

1. `docker build -t gateway-svc .` - соберите образ.
2. `docker run --rm --name gateway-svc -p <порт>:<порт> --env-file .env gateway-svc` - запустите контейнер. Оба порта должны совпадать с портом из `SERVER_ADDR`, тогда команды Makefile для работы с API будут работать без изменений.
3. `CTRL + C` или `docker stop gateway-svc` из другого терминала - пошлите серверу сигнал завершения, когда закончите работу.

<!-- markdownlint-disable MD033 -->
<details>
<summary>Подсказки</summary>

- Если gRPC-сервисы запущены локально на вашей машине (не в Docker), запустите контейнер в сети хоста: `docker run --rm --name gateway-svc --network host --env-file .env gateway-svc`. В этом режиме `localhost` внутри контейнера означает вашу машину, а флаг `-p` не нужен. Режим работает в Docker Engine на Linux (в том числе в WSL2).
- Если при сборке не удаётся скачать Go-модули (например, `proxy.golang.org` недоступен), передайте другой прокси через аргумент сборки: `docker build --build-arg GOPROXY=https://goproxy.io,direct -t gateway-svc .`
- Чтобы запустить контейнер в фоне, замените `--rm` на `-d`. Логи сервиса можно посмотреть командой `docker logs -f gateway-svc`, остановить и удалить контейнер - командами `docker stop gateway-svc` и `docker rm gateway-svc`.
- Проверить состояние запущенного сервиса можно командой `docker exec gateway-svc /app/healthcheck` (код завершения 0 означает, что сервис отвечает) или запросом `curl -i localhost:<порт>/healthz`.

</details>
<!-- markdownlint-enable MD033 -->
#### Работа с API

В отдельном терминале перейдите в корневую папку проекта и посылайте запросы на сервер.

- `make register` - зарегистрируйте пользователя.
- `make login` - авторизуйтесь (для удобства полученный JWT-токен будет сохранен в файл `.jwt` и вычитываться оттуда для запросов, требующих авторизации).
- `make search`, `make subscribe` и т.д. - работайте с API.
- `make clean` - выполните, чтобы удалить сохранённые JWT-токен и бинарник.
