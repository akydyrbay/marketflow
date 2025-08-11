CREATE TABLE AggregatedData(
    Data_id SERIAL PRIMARY KEY,
    Pair_name VARCHAR NOT NULL,
    Exchange VARCHAR(100) NOT NULL,
    StoredTime TimestampTZ DEFAULT NOW(),
    Average_price FLOAT NOT NULL, 
    Min_price FLOAT NOT NULL,
    Max_price FLOAT NOT NULL
);

CREATE TABLE LatestData(
    Exchange VARCHAR(100) NOT NULL,
    Pair_name VARCHAR NOT NULL,
    Price FLOAT NOT NULL,
    StoredTime BIGINT NOT NULL,
    CONSTRAINT unique_exchange_pair UNIQUE (Exchange, Pair_name)
);
