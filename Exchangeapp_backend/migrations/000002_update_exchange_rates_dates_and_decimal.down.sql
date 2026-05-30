DROP INDEX idx_rate_pair_date ON exchange_rates;

ALTER TABLE exchange_rates
MODIFY COLUMN from_currency LONGTEXT NULL,
MODIFY COLUMN to_currency LONGTEXT NULL,
MODIFY COLUMN rate DOUBLE NULL,
MODIFY COLUMN rate_date DATE NULL,
MODIFY COLUMN fetched_at DATETIME(3) NULL;

ALTER TABLE exchange_rates
DROP COLUMN rate_date;

ALTER TABLE exchange_rates
CHANGE COLUMN fetched_at data DATETIME(3) NULL;