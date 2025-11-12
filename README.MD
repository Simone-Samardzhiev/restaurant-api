# restaurant api

Work in progress restfull API for restaurant.

## Table of content

- [Installation](#installation)
- [Docs](#docs)
- [Development tools](#development-tools)
---

## Installation

Follow this steps to run the API.

### Prerequisites

- Go (1.25+)
- PostgreSQL
- Git

1. **Clone the repository**

   ```bash
   git clone https://github.com/Simone-Samardzhiev/shop-api-go
   cd shop-go-api
   ```

2. **Add environment variables (either via `.env` file or export them directly)**
   #### `.env` example:
    ```ini
    PORT=:8080
    ENVIRONMENT=development
    DB_URL=postgres://user:password@host:port/database
    DB_MAX_IDLE_CONNECTIONS=10
    DB_MAX_OPEN_CONNECTIONS=10
    IMAGES_API_KEY=YOUR_KEY
    USERNAME=adminUsername
    PASSWORD=adminPassword
    ```
   
3. **Run database migrations**

   ```bash
   task migrate-up
   ```

4. **Run the API**

   ```bash
   go run ./cmd/http/main.go
   ```
   
---

## Docs
- Apidog documentation:  
  👉 [View it here](https://cffasd1kb3.apidog.io)

---

## Development Tools

- [Task](https://taskfile.dev) — for running project tasks easily
- [golang-migrate](https://github.com/golang-migrate/migrate) — for database migrations