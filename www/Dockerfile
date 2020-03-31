# for build
FROM node:13.12.0-alpine3.11 AS builder

RUN apk update && \
    npm install -g npm && \
    npm install -g @vue/cli
RUN mkdir /app
WORKDIR /app
COPY frontend/ ./frontend
WORKDIR /app/frontend

RUN npm run build

# production
FROM nginx:alpine
WORKDIR /frontend
COPY --from=builder /app/frontend/dist .
COPY ./nginx_config/nginx.conf /etc/nginx/conf.d/default.conf
