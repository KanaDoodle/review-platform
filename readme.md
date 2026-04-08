# 🚀 Go High-Concurrency Review Platform

A high-performance backend system for a review platform built with Go, featuring Redis-based caching, geo-search, and a reliable seckill (flash sale) system using Redis Stream Consumer Groups.

---

# 📌 Overview

This project is a backend service for a review platform implemented in Go.  
It supports user authentication, shop browsing, nearby search, review publishing, and a high-concurrency voucher seckill system.

The system is designed with scalability and performance in mind, leveraging Redis for caching, geo queries, rate limiting, and asynchronous processing.

---

# 🛠 Tech Stack

- **Language**: Go
- **Framework**: Gin
- **ORM**: GORM
- **Database**: MySQL
- **Cache & MQ**: Redis (Cache, GEO, Stream)
- **Deployment**: Docker

---

# 🧱 Architecture

The project follows a layered architecture:

Handler (API Layer)
↓
Service (Business Logic)
↓
Repository (Data Access)
↓
MySQL / Redis

---

# 📂 Project Structure

review-platform/
├── cmd/server          # Entry point
├── config              # Configuration
├── internal/
│   ├── api             # HTTP handlers
│   ├── service         # Business logic
│   ├── repository      # Data access layer
│   ├── model           # Data models
│   ├── middleware      # JWT middleware
│   └── router          # Route registration
├── pkg/
│   ├── logger
│   └── response
├── scripts/            # SQL initialization
└── docker-compose.yml

---

# ✨ Features

## 🔐 Authentication

- SMS-style verification code login
- Redis-based code storage
- JWT token authentication
- Middleware-based authorization

---

## 🏪 Shop System

- Shop categories and listing
- Pagination support
- Shop detail query

---

## 📝 Review System

- Publish reviews (authenticated)
- Query reviews by shop

---

## ⚡ Redis Cache System

- Cache Aside pattern
- Empty-value caching (prevent cache penetration)
- Random TTL (prevent cache avalanche)
- Double-delete strategy (ensure cache consistency)

---

## 📍 Nearby Search (Redis GEO)

- GEO-based shop location indexing
- Category-based GEO partitioning
- Distance-based sorting
- MySQL fallback + result reordering

---

## 🔥 Seckill System (High-Concurrency Core)

### Version 1
- Redis Lua script
- Atomic stock deduction
- One-user-one-order guarantee

### Version 2 (Upgraded)
- Redis Stream as message queue
- Consumer Group for reliable consumption
- ACK mechanism
- Pending list recovery
- Poison message handling

---

## 🚦 Rate Limiting

- Sliding window algorithm using Redis ZSET
- Per-user request limiting
- Applied to seckill endpoint

---

## 🔄 Cache Consistency

- Delete cache on write
- Double-delete strategy with delay
- Reduces stale data under concurrency

---

# ⚙️ Quick Start

## 1. Clone the repository
```bash
git clone https://github.com/YOUR_USERNAME/review-platform.git
cd review-platform
```

## 2. Start dependencies (MySQL + Redis)

```bash
docker-compose up -d
```

## 3. Initialize database
```bash
docker exec -i review-mysql mysql -uroot -proot < scripts/init.sql
docker exec -i review-mysql mysql -uroot -proot < scripts/seed.sql
```
## 4. Run
```bash
go run ./cmd/server
```
## 5. Test
```bash
curl http://localhost:8080/api/v1/shops/1
```

