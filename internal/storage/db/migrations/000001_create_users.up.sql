CREATE TABLE "users" (
    "id" bigserial PRIMARY KEY,
    "login" varchar(255) UNIQUE NOT NULL,
    "encrypted_password" varchar(500) NOT NULL
);
