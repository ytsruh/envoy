# Authentication API Documentation

This document describes the authentication endpoints for user registration and login using JWT (JSON Web Tokens).

## Overview

The authentication system provides JWT-based authentication with the following features:
- User registration with email and password
- User login with email and password
- JWT tokens with 7-day expiration
- Password hashing using bcrypt
- Token returned directly in JSON response body
- Protected endpoints that require authentication
- Profile endpoint to retrieve JWT token data

## Environment Configuration

Before using the authentication endpoints, ensure the following environment variable is set:

```bash
JWT_SECRET=your-secret-key-change-this-in-production
```

**Security Note:** Use a strong, randomly generated secret key in production environments.

## Endpoints

### 1. Register a New User

Creates a new user account and returns a JWT token.

**Endpoint:** `POST /auth/register`

**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john.doe@example.com",
  "password": "securePassword123"
}
```

**Request Fields:**
- `name` (string, required): User's full name
- `email` (string, required): User's email address (must be unique)
- `password` (string, required): User's password (minimum 8 characters)

**Success Response (201 Created):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Doe",
    "email": "john.doe@example.com",
    "created_at": "2025-01-15T10:30:00Z"
  }
}
```

**Error Responses:**

- **400 Bad Request** - Invalid request body or validation error
```json
{
  "error": "Name, email, and password are required"
}
```

```json
{
  "error": "Password must be at least 8 characters"
}
```

- **409 Conflict** - User already exists
```json
{
  "error": "User with this email already exists"
}
```

- **500 Internal Server Error** - Server error
```json
{
  "error": "Failed to create user"
}
```

**Example cURL:**
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john.doe@example.com",
    "password": "securePassword123"
  }'
```

---

### 2. Login

Authenticates an existing user and returns a JWT token.

**Endpoint:** `POST /auth/login`

**Request Body:**
```json
{
  "email": "john.doe@example.com",
  "password": "securePassword123"
}
```

**Request Fields:**
- `email` (string, required): User's email address
- `password` (string, required): User's password

**Success Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Doe",
    "email": "john.doe@example.com",
    "created_at": "2025-01-15T10:30:00Z"
  }
}
```

**Error Responses:**

- **400 Bad Request** - Invalid request body
```json
{
  "error": "Email and password are required"
}
```

- **401 Unauthorized** - Invalid credentials
```json
{
  "error": "Invalid email or password"
}
```

- **500 Internal Server Error** - Server error
```json
{
  "error": "Failed to fetch user"
}
```

**Example cURL:**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "securePassword123"
  }'
```

---

### 3. Get Profile

Retrieves the authenticated user's profile information from their JWT token.

**Endpoint:** `GET /profile`

**Authentication:** Required - Bearer token in Authorization header

**Request Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Success Response (200 OK):**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "john.doe@example.com",
  "issued_at": 1642243800,
  "expires_at": 1642848600
}
```

**Response Fields:**
- `user_id` (string): The unique identifier for the user
- `email` (string): User's email address
- `issued_at` (integer): Token issuance timestamp (Unix epoch)
- `expires_at` (integer): Token expiration timestamp (Unix epoch)

**Error Responses:**

- **401 Unauthorized** - Missing or invalid token
```json
{
  "error": "Missing authorization header"
}
```

```json
{
  "error": "Invalid token"
}
```

```json
{
  "error": "Token has expired"
}
```

**Example cURL:**
```bash
curl -X GET http://localhost:8080/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

## Protected Endpoints

The following endpoints require JWT authentication. Include the token in the `Authorization` header as a Bearer token.

### GET /hello

A protected greeting endpoint that requires authentication.

**Authentication:** Required

**Request Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Success Response (200 OK):**
```
Hello, World!
```

**Error Response (401 Unauthorized):**
```json
{
  "error": "Missing authorization header"
}
```

**Example cURL:**
```bash
curl -X GET http://localhost:8080/hello \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### POST /goodbye

A protected farewell endpoint that requires authentication.

**Authentication:** Required

**Request Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Success Response (200 OK):**
```
Goodbye, World!
```

**Error Response (401 Unauthorized):**
```json
{
  "error": "Invalid authorization header format. Expected: Bearer <token>"
}
```

**Example cURL:**
```bash
curl -X POST http://localhost:8080/goodbye \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkjXVCJ9..."
```

---

## JWT Token Details

### Token Structure

The JWT token consists of three parts separated by dots (`.`):
- **Header**: Contains the algorithm (`HS256`) and token type (`JWT`)
- **Payload**: Contains the claims (user_id, email, iat, exp)
- **Signature**: HMAC SHA256 signature

### Token Claims

The JWT payload contains the following claims:

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "john.doe@example.com",
  "iat": 1642243800,
  "exp": 1642848600
}
```

- `user_id` (string): The unique identifier for the user
- `email` (string): The user's email address
- `iat` (integer): Issued at timestamp (Unix epoch)
- `exp` (integer): Expiration timestamp (Unix epoch) - **7 days from issuance**

### Token Expiration

Tokens expire **7 days** (604,800 seconds) after issuance. After expiration, users must login again to obtain a new token.

### Using the Token

Include the JWT token in the `Authorization` header for authenticated requests using the Bearer scheme:

```bash
curl -X GET http://localhost:8080/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

The server will:
1. Extract the token from the Authorization header
2. Validate the token signature
3. Check the token expiration
4. Add user information to the request context
5. Allow access to the protected resource

### Authorization Header Format

The Authorization header must follow this format:
```
Authorization: Bearer <token>
```

Where `<token>` is the JWT string received from the login or register endpoints.

---

## Security Considerations

1. **Password Security:**
   - Passwords are hashed using bcrypt with default cost factor
   - Minimum password length: 8 characters
   - Plain text passwords are never stored

2. **JWT Secret:**
   - Must be kept secure and never exposed
   - Use a strong, randomly generated secret (at least 32 characters)
   - Rotate secrets periodically in production

3. **Token Storage:**
   - Store tokens securely on the client side
   - Consider using secure, httpOnly cookies for web applications
   - Never log or expose tokens in URLs

4. **Email Uniqueness:**
   - Email addresses must be unique across all users
   - Case-sensitive email matching is used

5. **Soft Deletes:**
   - Deleted users (soft delete) cannot login
   - Email addresses from deleted accounts can be reused

---

## Testing

Example test using the complete authentication flow:

```bash
# 1. Register a new user
RESPONSE=$(curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "testPassword123"
  }')

# 2. Extract the token
TOKEN=$(echo $RESPONSE | jq -r '.token')

# 3. View your profile
curl -X GET http://localhost:8080/profile \
  -H "Authorization: Bearer $TOKEN"

# 4. Access protected greeting endpoints
curl -X GET http://localhost:8080/hello \
  -H "Authorization: Bearer $TOKEN"

curl -X POST http://localhost:8080/goodbye \
  -H "Authorization: Bearer $TOKEN"

# 5. Test with expired/invalid token (should return 401)
curl -X GET http://localhost:8080/profile \
  -H "Authorization: Bearer invalid-token-here"
```

### Testing Authentication Failures

```bash
# Missing Authorization header
curl -X GET http://localhost:8080/profile

# Invalid header format
curl -X GET http://localhost:8080/profile \
  -H "Authorization: InvalidFormat token"

# Expired or invalid token
curl -X GET http://localhost:8080/profile \
  -H "Authorization: Bearer expired.or.invalid.token"
```

---

## Error Handling

All error responses follow a consistent format:

```json
{
  "error": "Error message describing what went wrong"
}
```

HTTP status codes indicate the type of error:
- `200` - Success
- `201` - Resource created successfully (registration)
- `400` - Client error (invalid input, validation failure)
- `401` - Authentication error (invalid credentials, missing/invalid token)
- `409` - Conflict (duplicate resource)
- `500` - Server error (internal failure)

### Common Authentication Errors

**401 Unauthorized** responses can occur for several reasons:
- Missing Authorization header
- Invalid header format (not "Bearer <token>")
- Invalid token signature
- Expired token
- Malformed token

Always check the `error` field in the response for specific details.

---

## Implementation Notes

- The authentication system uses the standard library with Echo framework
- Database queries are handled through SQLC-generated code
- JWT implementation is custom, using HMAC SHA256 for signing
- User IDs are UUIDv4
- Timestamps use RFC 3339 format in JSON responses for auth endpoints
- Timestamps use Unix epoch format for profile endpoint
- JWT middleware adds user claims to the request context for use by handlers
- Protected routes automatically validate JWT before calling the handler