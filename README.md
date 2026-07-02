![Backend Tests](https://img.shields.io/badge/realworld%20hurl%20tests-passing-brightgreen)
![status of ci](https://github.com/sebasukodo/just-another-blog/actions/workflows/ci.yml/badge.svg)

# Just another Blog Website

This project is a self-hostable fullstack blogging platform built with Go and React.

The project is inspired by the excellent [RealWorld](https://github.com/realworld-apps/realworld) specification and aims to provide a clean, modern reference implementation for a real-world web application.

## Project Status
The backend is feature-complete and passes all [RealWorld hurl tests](https://github.com/realworld-apps/realworld/tree/main/specs/api).
The frontend is now fully implemented and functional, though it has not yet undergone thorough testing.

## Quick Start
To run the project locally, see [Run Development Server](#run-development-server).

You will also need a PostgreSQL database running either locally or inside a Docker container.

## Tech Stack

### Backend
- Go
- PostgreSQL
- SQLC
- Goose Migrations

### Frontend
- Typescript
- React
- Vite

---

## Run Development Server

### Backend
For detailed API information see [this link](https://docs.realworld.show/specifications/backend/endpoints/ "realworld api specifications").

#### Install dependencies
```bash
cd backend
go mod tidy
```

#### Start server
```bash
go run .
```

Backend runs on
```txt
http://localhost:7337
```

### Frontend

#### Install dependencies
```bash
cd frontend
npm install
```

#### Start dev server
```bash
npm run dev
```

Frontend runs on
```txt
http://localhost:5173
```
