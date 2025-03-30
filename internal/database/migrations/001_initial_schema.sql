-- Create articles table
CREATE TABLE IF NOT EXISTS articles (
                                        id SERIAL PRIMARY KEY,
                                        title TEXT NOT NULL,
                                        url TEXT NOT NULL UNIQUE,
                                        text TEXT,
                                        site_name TEXT NOT NULL,
                                        scraped_at TIMESTAMPTZ NOT NULL,
                                        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                        symbol TEXT NOT NULL
    );

-- Create an index for faster URL lookups
CREATE INDEX IF NOT EXISTS articles_url_idx ON articles(url);

-- Create an index for site_name for filtering
CREATE INDEX IF NOT EXISTS articles_site_name_idx ON articles(site_name);

-- Create an index for faster symbol lookups
CREATE INDEX IF NOT EXISTS articles_url_idx ON articles(symbol);