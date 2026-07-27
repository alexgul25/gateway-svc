# Gateway Service (gateway-svc)

Микросервис для проекта **Date Wishlist Hub**, единая точка входа, через которую клиенты взаимодействуют со всеми внутренними сервисами.

Стек: `Go`, `HTTP`, `gRPC`

## Основные возможности

**Gateway Service** предоставляет публичные эндпоинты для всех функций проекта.

Примечания:

- Подробности о структуре Body (`RegisterRequest`, `LoginRequest`, `SubscribeRequest`) и Response (`RegisterResponce` и т.д.) можно найти в файле [dto/user.go](./internal/dto/user.go).

- Столбец `Auth` говорит о том, ожидается ли Auth Header вида `Authorization | Bearer <token>`

| Method | Endpoint                        | Auth | Body | Response | Info                                              |
| :----: | ------------------------------- | :--: | :--: | :------: | ------------------------------------------------- |
| POST   | /api/users                      | ❌   | ✅   | ✅       | Регистрация пользователя                          |
| POST   | /api/auth/login                 | ❌   | ✅   | ✅       | Аутентификация пользователя                       |
| GET    | /api/users/me                   | ✅   | ❌   | ✅       | Получение данных своего профиля                   |
| GET    | /api/users?search_query=name    | ✅   | ❌   | ✅       | Поиск пользователей по имени                      |
| POST   | /api/subscriptions              | ✅   | ✅   | ❌       | Подписка на другого пользователя                  |
| DELETE | /api/subscriptions/{followeeID} | ✅   | ❌   | ❌       | Отписка от другого пользователя                   |
| GET    | /api/users/me/followers         | ✅   | ❌   | ✅       | Получение списка своих подписчиков                |
| GET    | /api/users/{userID}/followers   | ✅   | ❌   | ✅       | Получение списка подписчиков по ID                |

## Архитектура

Основные компоненты:

- команды приложения: [./cmd](./cmd/);

- сборка приложения http сервера: [./internal/app/http](./internal/app/http/);

- определение grpc клиентов для внутренних вызовов: [./internal/clients/grpc](./internal/clients/grpc/);

- слой http хендлеров: [./internal/http](./internal/http/);

- модели для абстракции от protobuf контрактов, возвращаемых grpc методами: [./internal/models](./internal/models/)
