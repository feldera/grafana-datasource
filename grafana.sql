CREATE TABLE times (
    count INT NOT NULL,
    ts TIMESTAMP NOT NULL
) with (
  'connectors' = '[{
    "transport": {
      "name": "datagen",
      "config": {
        "plan": [{
            "rate": 100,
            "fields": {
                "ts": { "range": ["2025-01-22T00:00:00Z", "2025-01-31T00:00:02Z"], "scale": 1000 }
            }
        }]
      }
    }
  }]'
);

CREATE MATERIALIZED VIEW v0 AS SELECT * FROM times;
