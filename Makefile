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
API := http://$(SERVER_ADDR)

JWT_FILE := .jwt
JWT := $(shell cat $(JWT_FILE) 2>/dev/null | tr -d '\r\n')

# Автоматически выбираем флаг для curl:
# --fail-with-body (curl 7.76+) выводит тело ошибки при 4xx/5xx
# -f (старый вариант) просто падает с кодом 22
CURL_FAIL_FLAG := $(shell curl --help all 2>/dev/null | grep -q "fail-with-body" && echo "--fail-with-body" || echo "-f")

.PHONY: help build run run-only \
		register login me search subscribe unsubscribe my-followers followers \
        add-place my-places user-places \
        clean clean-jwt clean-bin print-config \
        _check_tools _check_jwt

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

register: _check_tools ## Регистрация нового пользователя
	@read -rp "Email: " email; \
	read -rsp "Password: " pass; echo; \
	read -rp "Display name: " name; \
	payload=$$(jq -n --arg email "$$email" --arg pass "$$pass" --arg name "$$name" \
		'{email: $$email, password: $$pass, display_name: $$name}'); \
	echo "📬  Ответ сервера:"; \
	if resp=$$(curl -s $(CURL_FAIL_FLAG) -X POST $(API)/api/users \
		-H "Content-Type: application/json" \
		-d "$$payload" 2>&1); then \
		echo "$$resp" | jq . 2>/dev/null || echo "$$resp"; \
		echo "✅  Пользователь зарегистрирован"; \
	else \
		echo "$$resp"; \
		echo "❌  Не удалось зарегистрироваться"; \
		exit 1; \
	fi

login: _check_tools ## Аутентификация и сохранение JWT
	@read -rp "Email: " email; \
	read -rsp "Password: " pass; echo; \
	payload=$$(jq -n --arg email "$$email" --arg pass "$$pass" \
		'{email: $$email, password: $$pass}'); \
	echo "📬  Ответ сервера:"; \
	if resp=$$(curl -s $(CURL_FAIL_FLAG) -X POST $(API)/api/auth/login \
		-H "Content-Type: application/json" \
		-d "$$payload" 2>&1); then \
		echo "$$resp" | jq . 2>/dev/null || echo "$$resp"; \
		echo "$$resp" | jq -r '.access_token' > $(JWT_FILE); \
		chmod 600 $(JWT_FILE); \
		echo "🔐  Токен сохранён в $(JWT_FILE)"; \
		echo "✅  Пользователь авторизован"; \
	else \
		echo "$$resp"; \
		echo "❌  Не удалось авторизоваться"; \
		exit 1; \
	fi

me: _check_tools _check_jwt ## Данные профиля по JWT
	@echo "📬  Ответ сервера:"; \
	if resp=$$(curl -s $(CURL_FAIL_FLAG) -H "Authorization: Bearer $(JWT)" $(API)/api/users/me 2>&1); then \
		echo "$$resp" | jq . 2>/dev/null || echo "$$resp"; \
		echo "✅  Данные получены"; \
	else \
		echo "$$resp"; \
		echo "❌  Не удалось получить данные"; \
		exit 1; \
	fi

search: _check_tools _check_jwt ## Поиск пользователей по имени
	@read -rp "Имя для поиска: " q; \
	echo "📬  Ответ сервера:"; \
	if resp=$$(curl -s $(CURL_FAIL_FLAG) -H "Authorization: Bearer $(JWT)" \
		"$(API)/api/users?search_query=$$q" 2>&1); then \
		echo "$$resp" | jq . 2>/dev/null || echo "$$resp"; \
		echo "✅  Пользователи найдены"; \
	else \
		echo "$$resp"; \
		echo "❌  Не удалось выполнить поиск"; \
		exit 1; \
	fi

subscribe: _check_tools _check_jwt ## Подписаться на пользователя
	@read -rp "ID пользователя для подписки: " id; \
	payload=$$(jq -n --arg id "$$id" '{followee_id: $$id}'); \
	echo "📬  Ответ сервера:"; \
	if resp=$$(curl -s $(CURL_FAIL_FLAG) -X POST -H "Authorization: Bearer $(JWT)" \
		-H "Content-Type: application/json" \
		-d "$$payload" $(API)/api/subscriptions 2>&1); then \
		echo "$$resp" | jq . 2>/dev/null || echo "$$resp"; \
		echo "✅  Подписка на $$id"; \
	else \
		echo "$$resp"; \
		echo "❌  Не удалось подписаться"; \
		exit 1; \
	fi

unsubscribe: _check_tools _check_jwt ## Отписаться от пользователя
	@read -rp "ID пользователя для отписки: " id; \
	echo "📬  Ответ сервера:"; \
	if resp=$$(curl -s $(CURL_FAIL_FLAG) -X DELETE -H "Authorization: Bearer $(JWT)" \
		"$(API)/api/subscriptions/$$id" 2>&1); then \
		echo "$$resp" | jq . 2>/dev/null || echo "$$resp"; \
		echo "✅  Отписка от $$id"; \
	else \
		echo "$$resp"; \
		echo "❌  Не удалось отписаться"; \
		exit 1; \
	fi

my-followers: _check_tools _check_jwt ## Список своих подписчиков
	@echo "📬  Ответ сервера:"; \
	if resp=$$(curl -s $(CURL_FAIL_FLAG) -H "Authorization: Bearer $(JWT)" \
		$(API)/api/users/me/followers 2>&1); then \
		echo "$$resp" | jq . 2>/dev/null || echo "$$resp"; \
		echo "✅  Подписчики получены"; \
	else \
		echo "$$resp"; \
		echo "❌  Не удалось получить подписчиков"; \
		exit 1; \
	fi

followers: _check_tools _check_jwt ## Список подписчиков пользователя по ID
	@read -rp "ID пользователя: " id; \
	echo "📬  Ответ сервера:"; \
	if resp=$$(curl -s $(CURL_FAIL_FLAG) -H "Authorization: Bearer $(JWT)" \
		"$(API)/api/users/$$id/followers" 2>&1); then \
		echo "$$resp" | jq . 2>/dev/null || echo "$$resp"; \
		echo "✅  Подписчики получены"; \
	else \
		echo "$$resp"; \
		echo "❌  Не удалось получить подписчиков"; \
		exit 1; \
	fi

add-place: _check_tools _check_jwt ## Добавить место в профиль
	@read -rp "Название места: " name; \
	read -rp "Описание места: " info; \
	payload=$$(jq -n --arg name "$$name" --arg info "$$info" \
		'{name: $$name, info: $$info}'); \
	echo "📬  Ответ сервера:"; \
	if resp=$$(curl -s $(CURL_FAIL_FLAG) -X POST -H "Authorization: Bearer $(JWT)" \
		-H "Content-Type: application/json" \
		-d "$$payload" $(API)/api/places 2>&1); then \
		echo "$$resp" | jq . 2>/dev/null || echo "$$resp"; \
		echo "✅  Место добавлено"; \
	else \
		echo "$$resp"; \
		echo "❌  Не удалось добавить место"; \
		exit 1; \
	fi

my-places: _check_tools _check_jwt ## Показать свои места
	@echo "📬  Ответ сервера:"; \
	if resp=$$(curl -s $(CURL_FAIL_FLAG) -H "Authorization: Bearer $(JWT)" \
		$(API)/api/users/me/places 2>&1); then \
		echo "$$resp" | jq '.places[] | {name, info, created_at}' 2>/dev/null || echo "$$resp"; \
		echo "✅  Места получены"; \
	else \
		echo "$$resp"; \
		echo "❌  Не удалось получить места"; \
		exit 1; \
	fi

user-places: _check_tools _check_jwt ## Показать места пользователя по ID
	@read -rp "ID пользователя: " id; \
	echo "📬  Ответ сервера:"; \
	if resp=$$(curl -s $(CURL_FAIL_FLAG) -H "Authorization: Bearer $(JWT)" \
		"$(API)/api/users/$$id/places" 2>&1); then \
		echo "$$resp" | jq '.places[] | {name, info, created_at}' 2>/dev/null || echo "$$resp"; \
		echo "✅  Места получены"; \
	else \
		echo "$$resp"; \
		echo "❌  Не удалось получить места"; \
		exit 1; \
	fi

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

_check_jwt:
	@if [ -z "$(JWT)" ]; then \
		echo "❌  JWT не найден. Выполните make login."; \
		exit 1; \
	fi