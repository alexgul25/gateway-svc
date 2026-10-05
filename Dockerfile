# syntax=docker/dockerfile:1

# ------------------------------------------------------------
# Версия Go вынесена в аргумент, чтобы её было легко поменять
# без правки остального файла: docker build --build-arg GO_VERSION=1.27 
# ------------------------------------------------------------
ARG GO_VERSION=1.26

# ============================================================
# Стадия 1: build — компилируем бинарник
# ============================================================
FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

# Откуда Go скачивает модули. По умолчанию официальный прокси,
# но его можно переопределить при сборке:
#   docker build --build-arg GOPROXY=https://goproxy.io,direct .
ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

# Статическая сборка без C-зависимостей
ENV CGO_ENABLED=0

# Сначала копируем только файлы зависимостей и скачиваем модули.
# Этот слой пересоберётся, только если изменятся go.mod / go.sum.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Теперь копируем исходники и собираем.
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/gateway-svc ./cmd/svc-starter && \
    go build -trimpath -ldflags="-s -w" -o /out/healthcheck ./cmd/healthcheck

# ============================================================
# Стадия 2: runtime — минимальный образ только с бинарником
# ============================================================
FROM gcr.io/distroless/static:nonroot AS runtime

WORKDIR /app

COPY --from=build /out/gateway-svc /out/healthcheck /app/

# Запуск от непривилегированного пользователя
USER nonroot:nonroot

# Документация: сервис слушает этот порт внутри контейнера.
# Сам порт задаётся через SERVER_ADDR (например, SERVER_ADDR=:8080)
EXPOSE 8082

ENTRYPOINT ["/app/gateway-svc"]