ALTER TABLE exchange_rates
CHANGE COLUMN data fetched_at DATETIME(3) NULL;

ALTER TABLE exchange_rates
ADD COLUMN rate_date DATE NULL AFTER rate;

UPDATE exchange_rates
SET rate_date = DATE(fetched_at)
WHERE rate_date IS NULL
  AND fetched_at IS NOT NULL;

ALTER TABLE exchange_rates
MODIFY COLUMN from_currency VARCHAR(3) NOT NULL,
MODIFY COLUMN to_currency VARCHAR(3) NOT NULL,
MODIFY COLUMN rate DECIMAL(20,10) NOT NULL,
MODIFY COLUMN rate_date DATE NOT NULL,
MODIFY COLUMN fetched_at DATETIME(3) NOT NULL;

CREATE UNIQUE INDEX idx_rate_pair_date
ON exchange_rates (from_currency, to_currency, rate_date);