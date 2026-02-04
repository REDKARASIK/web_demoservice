CREATE SCHEMA IF NOT EXISTS banks;

create table if not exists banks.banks (
    id BIGSERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);