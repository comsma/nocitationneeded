FROM node:22-alpine AS css
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm install
COPY ui/ ./ui/
RUN npm run css

FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=css /app/ui/static/css/output.css ./ui/static/css/output.css
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /nocitationneeded ./cmd/web/

FROM gcr.io/distroless/static-debian13
WORKDIR /app
COPY --from=builder /nocitationneeded .
COPY --from=builder /app/ui ./ui
EXPOSE 8080
CMD ["./nocitationneeded"]
