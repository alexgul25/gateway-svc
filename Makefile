SHELL := /bin/bash
.SHELLFLAGS := -o pipefail -c
.DEFAULT_GOAL := help
.DELETE_ON_ERROR:

SERVICE_NAME := gateway-svc
BIN_DIR := bin
BINARY := $(BIN_DIR)/$(SERVICE_NAME)

SERVER_CMD := ./cmd/svc-starter

# Читаем SERVER_ADDR из .env
SERVER_ADDR := $(shell grep -m1 '^SERVER_ADDR=' .env 2>/dev/null | cut -d'=' -f2- | tr -d '\r' | xargs)
ifeq ($(shell echo $(SERVER_ADDR) | cut -c1),:)
  SERVER_ADDR := localhost$(SERVER_ADDR)
endif
ifneq ($(SERVER_ADDR),)
  API := http://$(SERVER_ADDR)
endif

JWT_FILE := .jwt
JWT := $(shell cat $(JWT_FILE) 2>/dev/null | tr -d '\r\n')

# ------------------------------------------------------------------------------
# Отправка HTTP-запроса к API
#
#   $(call request,МЕТОД,ПУТЬ[,JSON-ТЕЛО])
#
# Печатает ответ сервера целиком: статус, заголовки и тело (JSON форматируется
# через jq). После вызова в shell доступны переменные:
#   $$code - HTTP-код ответа (200, 404, 500...)
#   $$body - тело ответа
#
# Цель завершается с ошибкой, только если запрос не удалось выполнить
# (сервер недоступен, неверный адрес и т.п.). Ответ 4xx/5xx - это корректный
# ответ сервера, поэтому make завершается успешно.
#
# Заголовок Authorization добавляется, если есть сохранённый JWT.
# ------------------------------------------------------------------------------
request = \
	tmp=$$(mktemp -d) && trap 'rm -rf "$$tmp"' EXIT; \
	code=$$(curl -sS -X $(1) "$(API)$(2)" \
		-o "$$tmp/body" -D "$$tmp/headers" -w '%{http_code}' \
		$(if $(JWT),-H "Authorization: Bearer $(JWT)") \
		$(if $(3),-H "Content-Type: application/json" -d "$(3)")) \
		|| { echo "❌  Не удалось выполнить запрос $(1) $(API)$(2)"; exit 1; }; \
	body=$$(cat "$$tmp/body"); \
	echo "📬  Ответ сервера:"; \
	tr -d '\r' < "$$tmp/headers"; \
	if grep -qi '^content-type:.*json' "$$tmp/headers"; then \
		printf '%s\n' "$$body" | jq . 2>/dev/null || printf '%s\n' "$$body"; \
	elif [ -n "$$body" ]; then \
		printf '%s\n' "$$body"; \
	fi

# Итоговая строка по коду ответа: $(call report,Сообщение при успехе)
report = \
	if [[ $$code == 2* ]]; then \
		echo "✅  $(1)"; \
	else \
		echo "⚠️   Сервер вернул ошибку: HTTP $$code"; \
	fi

.PHONY: help build run run-only \
		register login me search subscribe unsubscribe my-followers followers \
        add-place my-places user-places \
        clean clean-jwt clean-bin print-config \
        _check_tools _check_api _check_jwt

help: ## Показать список доступных команд
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Собрать бинарник сервиса
	@echo "🔨  Сборка $(SERVICE_NAME)..."
	@mkdir -p "$(BIN_DIR)"
	@go build -o "$(BINARY)" $(SERVER_CMD)
	@echo "✅  Собран $(BINARY)"

run: build ## Собрать бинарник и запустить сервис
	@echo "🚀  Запуск $(SERVICE_NAME)..."
	@exec "$(BINARY)"

run-only: ## Запустить сервис без сборки (требуется собранный бинарник)
	@test -f "$(BINARY)" || { echo "❌  $(BINARY) не найден. Выполните make build"; exit 1; }
	@echo "🚀  Запуск $(SERVICE_NAME)..."
	@exec "$(BINARY)"

register: _check_tools _check_api ## Регистрация нового пользователя
	@read -rp "Email: " email; \
	read -rsp "Password: " pass; echo; \
	read -rp "Display name: " name; \
	payload=$$(jq -n --arg email "$$email" --arg pass "$$pass" --arg name "$$name" \
		'{email: $$email, password: $$pass, display_name: $$name}'); \
	$(call request,POST,/api/users,$$payload); \
	$(call report,Пользователь зарегистрирован)

login: _check_tools _check_api ## Аутентификация и сохранение JWT
	@read -rp "Email: " email; \
	read -rsp "Password: " pass; echo; \
	payload=$$(jq -n --arg email "$$email" --arg pass "$$pass" \
		'{email: $$email, password: $$pass}'); \
	$(call request,POST,/api/auth/login,$$payload); \
	$(call report,Пользователь авторизован); \
	if [[ $$code == 2* ]]; then \
		token=$$(printf '%s' "$$body" | jq -r '.access_token // empty' 2>/dev/null); \
		if [ -n "$$token" ]; then \
			(umask 077; printf '%s\n' "$$token" > $(JWT_FILE)); \
			echo "🔐  Токен сохранён в $(JWT_FILE)"; \
		else \
			echo "⚠️   В ответе нет поля access_token, токен не сохранён"; \
		fi; \
	fi

me: _check_tools _check_api _check_jwt ## Данные профиля по JWT
	@$(call request,GET,/api/users/me); \
	$(call report,Данные получены)

search: _check_tools _check_api _check_jwt ## Поиск пользователей по имени
	@read -rp "Имя для поиска: " q; \
	q=$$(jq -rn --arg q "$$q" '$$q | @uri'); \
	$(call request,GET,/api/users?search_query=$$q); \
	$(call report,Поиск выполнен)

subscribe: _check_tools _check_api _check_jwt ## Подписаться на пользователя
	@read -rp "ID пользователя для подписки: " id; \
	payload=$$(jq -n --arg id "$$id" '{followee_id: $$id}'); \
	$(call request,POST,/api/subscriptions,$$payload); \
	$(call report,Подписка на $$id)

unsubscribe: _check_tools _check_api _check_jwt ## Отписаться от пользователя
	@read -rp "ID пользователя для отписки: " id; \
	$(call request,DELETE,/api/subscriptions/$$id); \
	$(call report,Отписка от $$id)

my-followers: _check_tools _check_api _check_jwt ## Список своих подписчиков
	@$(call request,GET,/api/users/me/followers); \
	$(call report,Подписчики получены)

followers: _check_tools _check_api _check_jwt ## Список подписчиков пользователя по ID
	@read -rp "ID пользователя: " id; \
	$(call request,GET,/api/users/$$id/followers); \
	$(call report,Подписчики получены)

add-place: _check_tools _check_api _check_jwt ## Добавить место в профиль
	@read -rp "Название места: " name; \
	read -rp "Описание места: " info; \
	payload=$$(jq -n --arg name "$$name" --arg info "$$info" \
		'{name: $$name, info: $$info}'); \
	$(call request,POST,/api/places,$$payload); \
	$(call report,Место добавлено)

my-places: _check_tools _check_api _check_jwt ## Показать свои места
	@$(call request,GET,/api/users/me/places); \
	$(call report,Места получены)

user-places: _check_tools _check_api _check_jwt ## Показать места пользователя по ID
	@read -rp "ID пользователя: " id; \
	$(call request,GET,/api/users/$$id/places); \
	$(call report,Места получены)

clean: clean-jwt clean-bin ## Очистить артефакты
clean-jwt: ## Удалить сохранённый JWT
	@rm -f $(JWT_FILE)
	@echo "🧹  Файл $(JWT_FILE) удалён"
clean-bin: ## Удалить собранные бинарники
	@rm -rf $(BIN_DIR)
	@echo "🧹  Директория $(BIN_DIR) удалена"

print-config: ## Показать текущую конфигурацию
	@echo "SERVICE_NAME = $(SERVICE_NAME)"
	@echo "BINARY       = $(BINARY)"
	@echo "SERVER_ADDR  = $(SERVER_ADDR)"
	@echo "API          = $(API)"
	@echo "JWT_FILE     = $(JWT_FILE)"

_check_tools:
	@command -v curl >/dev/null 2>&1 || { echo "❌  curl не найден. Установите curl."; exit 1; }
	@command -v jq >/dev/null 2>&1 || { echo "❌  jq не найден. Установите jq."; exit 1; }

_check_api:
	@if [ -z "$(API)" ]; then \
		echo "❌  Адрес сервера не задан. Укажите SERVER_ADDR в .env или передайте API явно:"; \
		echo "    make <цель> API=http://localhost:8082"; \
		exit 1; \
	fi

_check_jwt:
	@if [ -z "$(JWT)" ]; then \
		echo "❌  JWT не найден. Выполните make login."; \
		exit 1; \
	fi