# Chirpy
I made a HTTP Server that kind of works like Twitter but rather than tweet , you Chirp here! , It was mainly for me to learn how to Build a HTTP Server. 

Built as part of the Boot.dev backend path.


## Features 
- User Registration , Lookup and Account Updates 
- JWT access tokens (1 hour) and refresh tokens (60 days) with revocation 
- Create , Read , Filter , Sort , and Delete Chirps
- Webhook endpoint to upgrade users to chirpy_red which is the premium membership
- Hashed Passwords

## What I used?
- GoLang (majorly)
- PostgreSQL
- sqlc for generation Go code from SQL queries
- goose for migrations
- A bit of HTML (lol by a bit i mean 5 lines)

## Requirements
- Go 1.27+
- PostgreSql running locally
- Goose 
- Sqlc 

## Setup
- Clone the repo
```bash
git clone https://github.com/Rohit-Rajvanshi/chirpy.git
cd chirpy
```

- Create the database 
 
```bash
sudo -u postgres psql
```

```sql
CREATE DATABASE chirpy;
```

- Create your .env file and .gitignore it 

```env
DB_URL="postgres://postgres:postgres@localhost:5432/chirpy?sslmode=disable"
PLATFORM="dev"
SECRET="your-jwt-secret"
POLKA_KEY="your-polka-api-key"
```

- Create your keys using different libraries :3 

- Run the migrations using Goose 

```bash
cd sql/schema
goose postgres <connection_string> up
```

- Build the Server and Run

```bash
go build -o chirpy && ./chirpy
```

The server starts on `http://localhost:8080`

## Documentation 
Documentation briefly covered in [docs/documents.md](docs/documentation.md)

## What I Learned
- Building a HTTP Server fullowing the RESTful COnvention with GoLang and PostgreSQL
- Stateless authentication with JWTs vs. stateful refresh tokens
- Verifying webhooks with API keys
- Writing Unit Tests in Golang




