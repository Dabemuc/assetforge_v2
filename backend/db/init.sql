CREATE TABLE t_etf (
  id VARCHAR(20) not null primary key,
  name TEXT,
  fundVolume VARCHAR(20),
  isDistributing BOOLEAN,
  releaseDate DATE,
  replicationMethod VARCHAR(20),
  shareClassVolume VARCHAR(20),
  totalExpenseRatio DECIMAL
)
