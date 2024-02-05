CREATE TABLE "balances" (
    "id" bigserial PRIMARY KEY,
    "user_id" bigint references "users"("id") UNIQUE NOT NULL,
    "current_amount" integer NOT NULL DEFAULT 0,
    "withdrawn_amount" integer NOT NULL DEFAULT 0
);
