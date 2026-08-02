SHELL := /bin/bash

# Читаем SERVER_ADDR из .env
SERVER_ADDR := $(shell grep -m1 '^SERVER_ADDR=' .env | cut -d'=' -f2- | xargs)
ifeq ($(shell echo $(SERVER_ADDR) | cut -c1),:)
  SERVER_ADDR := localhost$(SERVER_ADDR)
endif
API := http://$(SERVER_ADDR)

JWT_FILE := .jwt
JWT := $(shell cat $(JWT_FILE) 2>/dev/null)

.PHONY: help register login me search subscribe unsubscribe my-followers followers \
        add-place my-places user-places clean

help:
	@echo "Доступные команды:"
	@echo "  make run-svc        - Запустить Gateway Service (go run)"
	@echo "  make register       - Регистрация нового пользователя"
	@echo "  make login          - Аутентификация и сохранение JWT"
	@echo "  make me             - Показать свой профиль"
	@echo "  make search         - Поиск пользователей по имени"
	@echo "  make subscribe      - Подписаться на пользователя"
	@echo "  make unsubscribe    - Отписаться от пользователя"
	@echo "  make my-followers   - Список своих подписчиков"
	@echo "  make followers      - Список подписчиков пользователя по ID"
	@echo "  make add-place      - Добавить место"
	@echo "  make my-places      - Показать свои места"
	@echo "  make user-places    - Показать места пользователя по ID"
	@echo "  make clean          - Удалить сохранённый JWT"

run-svc:
	@echo "🚀 Запуск Gateway Service..."
	@go run ./cmd/svc-starter/main.go; exit 0

register:
	@echo "Запрос на регистрацию пользователя"; \
	read -p "Email: " email; \
	read -sp "Password: " pass; echo; \
	read -p "Display name: " name; \
	resp=$$(curl -s -X POST $(API)/api/users \
	  -H "Content-Type: application/json" \
	  -d "{\"email\":\"$$email\",\"password\":\"$$pass\",\"display_name\":\"$$name\"}"); \
	echo "Ответ сервера:"; \
	if echo "$$resp" | grep -qE '^[\[\{]'; then \
	    echo "$$resp" | jq .; \
		echo "✅  Пользователь зарегистрирован"; \
	else \
	    echo "$$resp"; \
		echo "❌  Что-то пошло не так..."; \
	fi

login:
	@echo "Запрос на авторизацию пользователя"; \
	read -p "Email: " email; \
	read -sp "Password: " pass; echo; \
	resp=$$(curl -s -X POST $(API)/api/auth/login \
	  -H "Content-Type: application/json" \
	  -d "{\"email\":\"$$email\",\"password\":\"$$pass\"}"); \
	echo "Ответ сервера:"; \
	if echo "$$resp" | grep -qE '^[\[\{]'; then \
	    echo "$$resp" | jq .; \
	    echo "$$resp" | jq -r '.access_token' > $(JWT_FILE); \
	    chmod 600 $(JWT_FILE); \
	    echo "🔐  Токен сохранён в $(JWT_FILE)"; \
		echo "✅  Пользователь авторизирован"; \
	else \
	    echo "$$resp"; \
	    echo "❌  Что-то пошло не так..."; \
	fi

me:
	@test -f $(JWT_FILE) || { echo "❌  Сначала выполните make login"; exit 1; }
	@echo "Запрос данных профиля текущего пользователя"; \
	resp=$$(curl -s -H "Authorization: Bearer $(JWT)" $(API)/api/users/me); \
	echo "Ответ сервера:"; \
	if echo "$$resp" | grep -qE '^[\[\{]'; then \
	    echo "$$resp" | jq .; \
		echo "✅  Данные получены"; \
	else \
	    echo "$$resp"; \
		echo "❌  Что-то пошло не так..."; \
	fi

search:
	@test -f $(JWT_FILE) || { echo "❌  Сначала выполните make login"; exit 1; }
	@echo "Запрос на поиск пользователей по имени"; \
	read -p "Имя для поиска: " q; \
	resp=$$(curl -s -H "Authorization: Bearer $(JWT)" "$(API)/api/users?search_query=$$q"); \
	echo "Ответ сервера:"; \
	if echo "$$resp" | grep -qE '^[\[\{]'; then \
	    echo "$$resp" | jq .; \
		echo "✅  Пользователи найдены"; \
	else \
	    echo "$$resp"; \
	    echo "❌  Что-то пошло не так..."; \
	fi

subscribe:
	@test -f $(JWT_FILE) || { echo "❌  Сначала выполните make login"; exit 1; }
	@echo "Запрос подписки на другого пользователя"; \
	read -p "ID пользователя для подписки: " id; \
	resp=$$(curl -s -X POST -H "Authorization: Bearer $(JWT)" \
	  -H "Content-Type: application/json" \
	  -d "{\"followee_id\":\"$$id\"}" $(API)/api/subscriptions); \
	echo "Ответ сервера:"; \
	if [ -z "$$resp" ]; then \
	    echo "(пустой)"; \
		echo "✅  Подписка на $$id"; \
	else \
	    echo "$$resp"; \
		echo "❌  Что-то пошло не так..."; \
	fi

unsubscribe:
	@test -f $(JWT_FILE) || { echo "❌  Сначала выполните make login"; exit 1; }
	@echo "Запрос отписки от другого пользователя"; \
	read -p "ID пользователя для отписки: " id; \
	resp=$$(curl -s -X DELETE -H "Authorization: Bearer $(JWT)" \
	  "$(API)/api/subscriptions/$$id"); \
	echo "Ответ сервера:"; \
	if [ -z "$$resp" ]; then \
	    echo "(пустой)"; \
		echo "✅  Отписка от $$id"; \
	else \
	    echo "$$resp"; \
		echo "❌  Что-то пошло не так..."; \
	fi

my-followers:
	@test -f $(JWT_FILE) || { echo "❌  Сначала выполните make login"; exit 1; }
	@echo "Запрос на получение списка своих подписчиков"; \
	resp=$$(curl -s -H "Authorization: Bearer $(JWT)" $(API)/api/users/me/followers); \
	echo "Ответ сервера:"; \
	if echo "$$resp" | grep -qE '^[\[\{]'; then \
	    echo "$$resp" | jq .; \
		echo "✅  Подписчики получены"; \
	else \
	    echo "$$resp"; \
	    echo "❌  Что-то пошло не так..."; \
	fi

followers:
	@test -f $(JWT_FILE) || { echo "❌  Сначала выполните make login"; exit 1; }
	@echo "Запрос на получение списка подписчиков пользователя"; \
	read -p "ID пользователя: " id; \
	resp=$$(curl -s -H "Authorization: Bearer $(JWT)" "$(API)/api/users/$$id/followers"); \
	echo "Ответ сервера:"; \
	if echo "$$resp" | grep -qE '^[\[\{]'; then \
	    echo "$$resp" | jq .; \
		echo "✅  Подписчики получены"; \
	else \
	    echo "$$resp"; \
	    echo "❌  Что-то пошло не так..."; \
	fi

add-place:
	@test -f $(JWT_FILE) || { echo "❌  Сначала выполните make login"; exit 1; }
	@echo "Запрос на добавление места"; \
	read -p "Название места: " name; \
	read -p "Описание места: " info; \
	resp=$$(curl -s -X POST -H "Authorization: Bearer $(JWT)" \
	  -H "Content-Type: application/json" \
	  -d "{\"name\":\"$$name\",\"info\":\"$$info\"}" $(API)/api/places); \
	echo "Ответ сервера:"; \
	if echo "$$resp" | grep -qE '^[\[\{]'; then \
	    echo "$$resp" | jq .; \
	    echo "✅  Место добавлено"; \
	else \
	    echo "$$resp"; \
	    echo "❌  Что-то пошло не так..."; \
	fi

my-places:
	@test -f $(JWT_FILE) || { echo "❌  Сначала выполните make login"; exit 1; }
	@echo "Запрос на получение своих мест"; \
	resp=$$(curl -s -H "Authorization: Bearer $(JWT)" $(API)/api/users/me/places); \
	echo "Ответ сервера:"; \
	if echo "$$resp" | grep -qE '^[\[\{]'; then \
	    echo "$$resp" | jq '.places[] | {name, info, created_at}'; \
	    echo "✅  Места получены"; \
	else \
	    echo "$$resp"; \
	    echo "❌  Что-то пошло не так..."; \
	fi

user-places:
	@test -f $(JWT_FILE) || { echo "❌  Сначала выполните make login"; exit 1; }
	@echo "Запрос на получение мест пользователя"; \
	read -p "ID пользователя: " id; \
	resp=$$(curl -s -H "Authorization: Bearer $(JWT)" "$(API)/api/users/$$id/places"); \
	echo "Ответ сервера:"; \
	if echo "$$resp" | grep -qE '^[\[\{]'; then \
	    echo "$$resp" | jq '.places[] | {name, info, created_at}'; \
	    echo "✅  Места получены"; \
	else \
	    echo "$$resp"; \
	    echo "❌  Что-то пошло не так..."; \
	fi

clean:
	@rm -f $(JWT_FILE)
	@echo "🗑️   Файл $(JWT_FILE) удалён"