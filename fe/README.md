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

Run the three frontend portals as independent Vite processes in three
terminals:

```bash
npm run dev:user    # http://localhost:5173
npm run dev:tenant  # http://localhost:5174
npm run dev:admin   # http://localhost:5175
npm run dev:all     # start all three processes together
```

User and Tenant use the User Gateway (`18080`); Admin uses the Admin Gateway
(`18082`).

## Current routes

- User process `/login`: customer login
- Tenant process `/login`: tenant operator login
- Admin process `/login`: platform admin login

The three local login links are `http://localhost:5173/login`,
`http://localhost:5174/login`, and `http://localhost:5175/login`.

The codebase shares components, i18n, and auth utilities, but each process has
its own root page, token namespace, API base URL, and authentication path.

## Localization

Vietnamese (`vi`) and English (`en`) are supported through the shared catalog
under `src/lib/i18n.ts`. Vietnamese remains the default product language unless
the browser already has a saved choice. The language switcher persists the
selection in `localStorage` and updates the document language for accessibility.

Access and refresh tokens are currently stored in local storage because the API returns both tokens in JSON. For production, prefer an HTTP-only secure cookie for the refresh token.
