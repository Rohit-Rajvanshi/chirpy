# Chirpy API Documentation

Base URL (local): `http://localhost:8080`

All request and response bodies are JSON unless stated otherwise. Timestamps are RFC 3339 (e.g. `2026-10-28T10:15:00Z`). IDs are UUIDs.

---

## Authentication

Chirpy uses two kinds of tokens, plus an API key for the payment webhook.

| Token | How you get it | Lifetime | Where it goes |
|---|---|---|---|
| Access token (JWT, HS256) | `POST /api/login` or `POST /api/refresh` | 1 hour | `Authorization: Bearer <token>` |
| Refresh token (random 256-bit hex string, stored in DB) | `POST /api/login` | 60 days, or until revoked | `Authorization: Bearer <token>` (only on `/api/refresh` and `/api/revoke`) |
| Polka API key | Set via `POLKA_KEY` env var | Doesn't expire | `Authorization: ApiKey <key>` |

Access tokens are stateless: the server checks the signature and expiry, not the database. Refresh tokens are checked against the database, so they can be revoked.

---

## Errors

Every error response has the same shape:

```json
{
  "error": "description of what went wrong"
}
```

| Status | Meaning in Chirpy |
|---|---|
| `400 Bad Request` | Body couldn't be parsed, or failed validation (e.g. chirp too long) |
| `401 Unauthorized` | Missing, malformed, expired, or invalid token/key; wrong email or password |
| `403 Forbidden` | Authenticated, but not allowed (e.g. deleting someone else's chirp, reset outside dev) |
| `404 Not Found` | Resource doesn't exist |
| `500 Internal Server Error` | Something broke on the server (database error, etc.) |

---

## Resources

### User object

```json
{
  "id": "5a1a8b2c-3d4e-4f60-8a9b-0c1d2e3f4a5b",
  "created_at": "2026-10-28T10:15:00Z",
  "updated_at": "2026-10-28T10:15:00Z",
  "email": "rohit@example.com",
  "is_chirpy_red": false
}
```

Passwords are hashed and never returned.

### Chirp object

```json
{
  "id": "9f8e7d6c-5b4a-4321-9abc-def012345678",
  "created_at": "2026-10-28T10:20:00Z",
  "updated_at": "2026-10-28T10:20:00Z",
  "body": "Learning Go is going well",
  "user_id": "5a1a8b2c-3d4e-4f60-8a9b-0c1d2e3f4a5b"
}
```

---

## Users

### Create a user

`POST /api/users`

No auth required.

**Request**

```json
{
  "email": "rohit@example.com",
  "password": "hunter2"
}
```

**Response** `201 Created` — a [User object](#user-object).

**Errors**
- `400` — invalid JSON
- `500` — couldn't hash password or create user (e.g. email already exists)

---

### Update your email and password

`PUT /api/users`

**Auth:** `Authorization: Bearer <access_token>`

Updates the user the access token belongs to. Both fields are required.

**Request**

```json
{
  "email": "new@example.com",
  "password": "newpassword"
}
```

**Response** `200 OK` — the updated [User object](#user-object).

**Errors**
- `400` — invalid JSON
- `401` — missing or invalid access token
- `500` — database error

---

## Auth

### Log in

`POST /api/login`

No auth required.

**Request**

```json
{
  "email": "rohit@example.com",
  "password": "hunter2"
}
```

**Response** `200 OK` — the User object plus both tokens:

```json
{
  "id": "5a1a8b2c-3d4e-4f60-8a9b-0c1d2e3f4a5b",
  "created_at": "2026-10-28T10:15:00Z",
  "updated_at": "2026-10-28T10:15:00Z",
  "email": "rohit@example.com",
  "is_chirpy_red": false,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "56aa826d22baab4b5ec2cea41a59ecbba03e542aedbb31d9b80326ac8ffcfa2a"
}
```

**Errors**
- `400` — invalid JSON
- `401` — incorrect email or password (the same message for both, so the API doesn't reveal which emails exist)
- `500` — couldn't create tokens

---

### Get a new access token

`POST /api/refresh`

**Auth:** `Authorization: Bearer <refresh_token>`

No request body.

**Response** `200 OK`

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Errors**
- `401` — refresh token missing, doesn't exist, expired, or revoked

---

### Revoke a refresh token

`POST /api/revoke`

**Auth:** `Authorization: Bearer <refresh_token>`

No request body. After this, the refresh token can't be used again. Access tokens already issued stay valid until they expire.

**Response** `204 No Content`

**Errors**
- `401` — refresh token missing
- `500` — database error

---

## Chirps

### Create a chirp

`POST /api/chirps`

**Auth:** `Authorization: Bearer <access_token>`

The author is taken from the token, not the body.

**Request**

```json
{
  "body": "Learning Go is going well"
}
```

**Rules**
- Max 140 characters.
- The words `kerfuffle`, `sharbert`, and `fornax` are replaced with `****` (case-insensitive, whole words only).

**Response** `201 Created` — a [Chirp object](#chirp-object) with the cleaned body.

**Errors**
- `400` — invalid JSON, or chirp is longer than 140 characters
- `401` — missing or invalid access token
- `500` — database error

---

### List chirps

`GET /api/chirps`

No auth required.

**Query parameters** (both optional)

| Param | Values | Default | Effect |
|---|---|---|---|
| `author_id` | a user UUID | — | Only return chirps by this user |
| `sort` | `asc`, `desc` | `asc` | Sort by `created_at` |

**Example**

```
GET /api/chirps?author_id=5a1a8b2c-3d4e-4f60-8a9b-0c1d2e3f4a5b&sort=desc
```

**Response** `200 OK` — an array of [Chirp objects](#chirp-object). Empty array if there are none.

**Errors**
- `400` — `author_id` isn't a valid UUID
- `500` — database error

---

### Get a chirp

`GET /api/chirps/{chirpID}`

No auth required.

**Response** `200 OK` — a [Chirp object](#chirp-object).

**Errors**
- `400` — `chirpID` isn't a valid UUID
- `404` — chirp doesn't exist

---

### Delete a chirp

`DELETE /api/chirps/{chirpID}`

**Auth:** `Authorization: Bearer <access_token>`

Only the chirp's author can delete it.

**Response** `204 No Content`

**Errors**
- `400` — `chirpID` isn't a valid UUID
- `401` — missing or invalid access token
- `403` — you're not the author
- `404` — chirp doesn't exist
- `500` — database error

---

## Webhooks

### Polka payment webhook

`POST /api/polka/webhooks`

**Auth:** `Authorization: ApiKey <POLKA_KEY>`

Called by Polka (the payment provider) when a user pays. Upgrades the user to Chirpy Red.

**Request**

```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "5a1a8b2c-3d4e-4f60-8a9b-0c1d2e3f4a5b"
  }
}
```

**Response** `204 No Content` — for a successful upgrade, and also for any event other than `user.upgraded` (it's acknowledged and ignored, so Polka doesn't keep retrying).

**Errors**
- `400` — invalid JSON or invalid `user_id`
- `401` — API key missing or wrong
- `404` — user doesn't exist

---

## Admin & health

### Health check

`GET /api/healthz`

**Response** `200 OK`, `Content-Type: text/plain`

```
OK
```

---

### Metrics

`GET /admin/metrics/`

Returns an **HTML** page showing how many times the file server (`/app/`) has been hit since the server started. The count lives in memory and resets on restart.

---

### Reset

`POST /admin/reset`

**Dev only.** Deletes all users (and their chirps and refresh tokens, via cascade) and resets the hit counter.

**Response** `200 OK`

**Errors**
- `403` — `PLATFORM` is not `dev`

---

## Static files

`GET /app/`

Serves files from the project root (`index.html`, `/app/assets/logo.png`). Every request here increments the metrics counter.