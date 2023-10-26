CREATE EXTENSION IF NOT EXISTS citext;
Create TYPE   USERTYPE as ENUM ('rider', 'customer', 'admin');
CREATE TABLE IF NOT EXISTS users (
                                     id bigserial PRIMARY KEY,
                                     created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
                                     userName text NOT NULL,
                                     type USERTYPE NOT NULL,
                                     email citext UNIQUE NOT NULL,
                                     password_hash bytea NOT NULL,
                                     activated bool NOT NULL,
                                     version integer NOT NULL DEFAULT 1
);