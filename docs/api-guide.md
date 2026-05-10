# API Guide

The Booking Management API is a RESTful service that uses JSON for data exchange and JWT for authentication.

## 🔗 Base URL
`http://localhost:8088/v1`

## 📖 Live Documentation
The API includes an interactive reference powered by **Scalar**:
- **URL**: `http://localhost:8088/v1/doc`
- **Source**: Swagger 2.0 (generated from code annotations)

## 🔑 Authentication

Most endpoints require a Bearer token in the `Authorization` header:
`Authorization: Bearer <your_jwt_token>`

### Auth Flow
1. **Register**: `POST /auth/register`
2. **Login**: `POST /auth/login` -> Returns `access_token` and `refresh_token`.
3. **Refresh**: `POST /auth/refresh` -> Use refresh token to get a new access token.
4. **Logout**: `POST /auth/logout` -> Revokes the token JTI.

## 📡 Endpoints Summary

### Public
| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | System health check |
| GET | `/rooms` | List all rooms (filtered) |
| GET | `/rooms/:id` | Get specific room details |
| GET | `/availability` | Search available rooms by date |

### Customer (Authenticated)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/bookings` | Create a new booking |
| GET | `/bookings/me` | List my current and past bookings |
| POST | `/bookings/:id/cancel` | Cancel my booking |

### Admin & Staff
| Method | Path | Description | Roles |
|--------|------|-------------|-------|
| GET | `/bookings` | List all bookings in the system | Admin, Officer |
| PATCH | `/rooms/:id/status` | Update room status (cleaning, etc) | Admin, Officer |
| GET | `/admin/users` | List all registered users | Admin |
| POST | `/admin/rooms` | Create a new room | Admin |
| PUT | `/admin/rooms/:id` | Update room info | Admin |
| DELETE | `/admin/rooms/:id` | Remove a room | Admin |

## 🛡 Security Measures
- **Rate Limiting**: 50 req/s per client (burst of 100).
- **Security Headers**: Standard headers like `X-Content-Type-Options`, `Referrer-Policy` are enforced.
- **CORS**: Configurable origins for frontend integration.
