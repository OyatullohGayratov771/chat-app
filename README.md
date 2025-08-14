# Chat-App (Microservices Architecture)

Real-time chat application built with **Golang**, **Redis**, **Kafka**, **Docker**, and **WebSockets**.

## Features
- JWT authentication
- Real-time messaging
- Chat rooms (1-to-1 and group chat)
- Notifications service
- Scalable microservices architecture

## Tech Stack
- Golang (Gin, Gorilla WebSocket)
- PostgreSQL
- Redis
- Kafka
- Docker, Docker Compose

## Architecture
![Architecture Diagram](docs/architecture.png)

## Services
- **Auth Service**: Registration, login, JWT tokens
- **Chat Service**: Messaging, WebSocket connection
- **Notification Service**: New message notifications
