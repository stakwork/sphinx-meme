
-- media table stores references to media, and terms to buy 
CREATE TABLE media (
  id TEXT NOT NULL PRIMARY KEY,
  owner_pub_key TEXT NOT NULL,
  name TEXT,
  description TEXT,
  price BIGINT,
  tags TEXT[] not null default '{}',
  ttl BIGINT,
  filename TEXT,
  size BIGINT,
  mime TEXT,
  nonce TEXT,
  created timestamptz,
  updated timestamptz,
  expiry timestamptz, -- optional permanent deletion of file
  total_sats BIGINT,
  total_buys BIGINT,
  template boolean,
  width INT,
  height INT
);

-- for searching 

ALTER TABLE media ADD COLUMN tsv tsvector;

UPDATE media SET tsv =
  setweight(to_tsvector(name), 'A') ||
	setweight(to_tsvector(description), 'B') ||
	setweight(array_to_tsvector(tags), 'C');

CREATE INDEX media_tsv ON media USING GIN(tsv);

-- select

SELECT name, description, tags
FROM media, to_tsquery('foo') q
WHERE tsv @@ q;

-- rank

SELECT name, id, description, ts_rank(tsv, q) as rank
FROM media, to_tsquery('anothe') q
WHERE tsv @@ q
ORDER BY rank DESC
LIMIT 12;

-- plainto_tsquery is another way
