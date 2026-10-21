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
- `/` or `/user`: authenticated customer portal
- `/tenant`: organization/venue operations for `owner` and `staff` roles
- `/admin`: platform operations for `admin` role

The frontend has three portal boundaries in the router: User, Tenant, and
Admin. Tenant and Admin routes fail closed until the User Gateway includes a
`roles` array in the authenticated user response. The current public contract
does not expose roles yet, so this split is source-level and route-level; live
permission acceptance still requires the auth contract and Admin Gateway.

## Localization

Vietnamese (`vi`) and English (`en`) are supported through the shared catalog
under `src/lib/i18n.ts`. Vietnamese remains the default product language unless
the browser already has a saved choice. The language switcher persists the
selection in `localStorage` and updates the document language for accessibility.

Access and refresh tokens are currently stored in local storage because the API returns both tokens in JSON. For production, prefer an HTTP-only secure cookie for the refresh token.
