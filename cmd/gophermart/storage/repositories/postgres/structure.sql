CREATE TABLE IF NOT EXISTS users (
  user_id SERIAL,
  login VARCHAR UNIQUE,
  password VARCHAR
);

CREATE TABLE IF NOT EXISTS user_balances (
  user_id VARCHAR PRIMARY KEY,
  value INTEGER,
);

CREATE TABLE IF NOT EXISTS user_orders (
  user_id VARCHAR PRIMARY KEY,
  order_id VARCHAR UNIQUE,
);

