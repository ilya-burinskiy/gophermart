CREATE TABLE "orders" (
    "id" bigserial PRIMARY KEY,
    "user_id" bigint references "users"("id") NOT NULL,
    "status" integer NOT NULL DEFAULT 0,
    "accrual" integer NOT NULL DEFAULT 0,
    "number" varchar(255) UNIQUE NOT NULL,
    "created_at" timestamptz NOT NULL
);
