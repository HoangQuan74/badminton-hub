# BadmintonHub Frontend

Customer-facing web application for finding and booking badminton courts.

## Stack

- React 19, TypeScript, Vite
- React Router
- TanStack Query and Axios
- Zustand
- React Hook Form and Zod

## Local development

```bash
cp .env.example .env
npm install
npm run dev
```

The API defaults to `http://localhost:8080`. Override it with `VITE_API_URL`.

## Current routes

- `/login`: customer login
- `/register`: customer registration
- `/`: authenticated customer home

Access and refresh tokens are currently stored in local storage because the API returns both tokens in JSON. For production, prefer an HTTP-only secure cookie for the refresh token.
