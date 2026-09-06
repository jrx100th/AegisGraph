FROM node:22-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

FROM golang:1.22-alpine AS backend
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
COPY --from=frontend /src/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/aegisgraph ./cmd/aegisgraph

FROM alpine:3.20
RUN adduser -D -H -u 10001 aegisgraph
WORKDIR /app
COPY --from=backend /out/aegisgraph /app/aegisgraph
COPY --from=frontend /src/frontend/dist /app/frontend/dist
RUN chown -R aegisgraph:aegisgraph /app
USER aegisgraph
EXPOSE 8080
ENTRYPOINT ["/app/aegisgraph","-addr",":8080","-static","/app/frontend/dist","-db","/data/aegisgraph.db"]
