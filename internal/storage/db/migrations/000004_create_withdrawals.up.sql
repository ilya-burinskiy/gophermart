CREATE TABLE "withdrawals" (
    "id" bigserial PRIMARY KEY,
    "order_number" varchar(255) UNIQUE NOT NULL,
    "user_id" bigint references "users"("id") NOT NULL,
    "sum" integer NOT NULL DEFAULT 0,
    "processed_at" timestamptz NOT NULL
);
