DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS accounts;

CREATE TABLE accounts
(
    id            TEXT    NOT NULL,
    balance_cents INTEGER NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE transactions
(
    account      TEXT    NOT NULL,
    id           TEXT    NOT NULL,
    amount_cents INTEGER NOT NULL,
    description  TEXT,
    PRIMARY KEY (account, id),
    FOREIGN KEY(account) REFERENCES accounts(id)
);

INSERT INTO accounts (id, balance_cents)
VALUES ('90903a90-d8f0-45eb-a4aa-dea4d24b2f54', 100000); /* 1,000.00€ */
