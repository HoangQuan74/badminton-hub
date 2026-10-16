FROM node:22-alpine AS builder

WORKDIR /src/fe
COPY fe/package.json fe/package-lock.json ./
RUN npm ci
COPY fe/ ./
ARG VITE_API_URL=/
ENV VITE_API_URL=${VITE_API_URL}
RUN npm run build

FROM nginx:1.29-alpine
COPY devops/docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /src/fe/dist /usr/share/nginx/html
EXPOSE 80
