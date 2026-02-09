# server builder

FROM golang:1.24 AS server_builder

ENV APP_HOME=/code/bbs-go/server
WORKDIR "$APP_HOME"

COPY ./server ./
ENV http_proxy=http://10.4.212.21:8123 https_proxy=http://10.4.212.21:8123
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download
ENV http_proxy= https_proxy=
RUN CGO_ENABLED=0 go build -v -o bbs-go main.go && chmod +x bbs-go


# site builder
FROM node:20-alpine AS site_builder

ENV APP_HOME=/code/bbs-go/site
WORKDIR "$APP_HOME"

COPY ./site ./
ENV http_proxy=http://10.4.212.21:8123 https_proxy=http://10.4.212.21:8123
RUN npm install -g pnpm --registry=https://registry.npmmirror.com
RUN pnpm install --registry=https://registry.npmmirror.com
RUN npm install -g pnpm
RUN pnpm install
RUN pnpm build
ENV http_proxy= https_proxy=


# admin builder
FROM node:20-alpine AS admin_builder

ENV APP_HOME=/code/bbs-go/admin
WORKDIR "$APP_HOME"

COPY ./admin ./
ENV http_proxy=http://10.4.212.21:8123 https_proxy=http://10.4.212.21:8123
RUN npm install -g pnpm --registry=https://registry.npmmirror.com
RUN pnpm install --registry=https://registry.npmmirror.com
RUN npm install -g pnpm
RUN pnpm install
RUN pnpm build
ENV http_proxy= https_proxy=

# run
FROM node:20-alpine

ENV APP_HOME=/app/bbs-go
WORKDIR "$APP_HOME"

COPY --from=server_builder /code/bbs-go/server/bbs-go ./server/bbs-go
COPY --from=server_builder /code/bbs-go/server/migrations ./server/migrations
COPY --from=server_builder /code/bbs-go/server/locales ./server/locales
COPY --from=site_builder /code/bbs-go/site/.output ./site/.output
COPY --from=site_builder /code/bbs-go/site/node_modules ./site/node_modules
COPY --from=admin_builder /code/bbs-go/admin/dist ./server/admin

COPY ./start.sh ${APP_HOME}/start.sh
RUN chmod +x ${APP_HOME}/start.sh

EXPOSE 8082 3000

CMD ["./start.sh"]
