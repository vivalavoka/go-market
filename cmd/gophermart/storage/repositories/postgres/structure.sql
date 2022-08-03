CREATE TABLE IF NOT EXISTS users (
  user_id SERIAL,
  login VARCHAR UNIQUE,
  password VARCHAR
);

CREATE TABLE IF NOT EXISTS user_balances (
  user_id VARCHAR PRIMARY KEY,
  value INTEGER,
  CONSTRAINT fk_user_balance_user_id FOREIGN KEY(user_id) REFERENCES users(user_id)
);

CREATE TABLE IF NOT EXISTS user_orders (
  user_id VARCHAR PRIMARY KEY,
  order_id INTEGER UNIQUE,
  CONSTRAINT fk_user_orders_user_id FOREIGN KEY(user_id) REFERENCES users(user_id)
);

