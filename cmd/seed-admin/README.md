# seed-admin — bootstrap the first admin account

## Why this exists

`POST /auth/admin-invite` is gated by `RequireRole("admin")` — only an
existing admin can invite new ones. In a fresh deployment there isn't an
admin yet, so this one-shot CLI exists to break the chicken-and-egg loop.

Run it **once per environment** (dev / staging / prod). After that, use
the admin-invite endpoint for every subsequent admin.

## Usage

Point at the same database the API server uses (it reads the same
`DB_*` env vars as `cmd/api`). Then:

```bash
SEED_ADMIN_EMAIL=admin@smartsewa.lk \
SEED_ADMIN_NAME="Platform Admin" \
SEED_ADMIN_PASSWORD='a-strong-password-of-at-least-10-chars' \
SEED_ADMIN_PHONE='+94 77 123 4567' \
  go run ./cmd/seed-admin
```

Mandatory env vars: `SEED_ADMIN_EMAIL`, `SEED_ADMIN_PASSWORD`
(min 10 chars). `SEED_ADMIN_NAME` and `SEED_ADMIN_PHONE` are optional
with sensible defaults.

## Behaviour

- If the email **doesn't exist**, a new user row is created with
  `role=admin`, `is_active=true`, and the given password hashed with
  bcrypt (same algorithm the API uses).
- If the email **exists but role!='admin'**, the row is promoted to
  `role=admin` and `is_active=true`. **The existing password is NOT
  overwritten** — if the user has forgotten their password, use
  `POST /auth/forgot-password` after promotion.
- If the email **exists and is already admin**, the tool is a no-op
  and exits 0.

## Running against prod

In production use the Cloud Run job pattern (don't open a direct SQL
tunnel unless there's no other option):

```bash
# From the backend repo root, build + push an image + run as a Cloud Run Job.
gcloud run jobs deploy seed-admin \
  --project sunrise-508210 \
  --region asia-south1 \
  --image gcr.io/sunrise-508210/scw-backend-seed-admin:latest \
  --command "/app/seed-admin" \
  --set-secrets SEED_ADMIN_PASSWORD=admin-bootstrap-password:latest \
  --set-env-vars \
"SEED_ADMIN_EMAIL=admin@smartsewa.lk,SEED_ADMIN_NAME=Platform Admin,SEED_ADMIN_PHONE=+94 77 123 4567"

gcloud run jobs execute seed-admin --project sunrise-508210 --region asia-south1
```

The job inherits the Cloud SQL / IAM settings from the API service.

### Alternative (one-liner): `gcloud sql connect`

If a Cloud Run Job is overkill:

```sql
-- Hash a password with bcrypt out-of-band (e.g. python -c 'import bcrypt; print(bcrypt.hashpw(b"your-pw", bcrypt.gensalt(12)).decode())').
-- Then:
INSERT INTO users (id, name, email, phone, password_hash, avatar_url, role, is_active, created_at, updated_at)
VALUES (
  gen_random_uuid(),
  'Platform Admin',
  'admin@smartsewa.lk',
  '+94 77 123 4567',
  '$2b$12$...hash...',
  'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=200',
  'admin',
  true,
  NOW(),
  NOW()
);
```

Recommended path: the Cloud Run Job — it reuses the backend's bcrypt
constants and config loader, which closes the "wrong cost factor" trap.
