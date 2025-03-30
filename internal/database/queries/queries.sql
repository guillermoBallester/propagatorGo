-- name: GetArticle :one
SELECT * FROM articles
WHERE id = $1;

-- name: GetArticleByURL :one
SELECT * FROM articles
WHERE url = $1;

-- name: GetArticleBySymbol :many
SELECT * FROM articles
WHERE symbol = $1;

-- name: GetArticleBySite :many
SELECT * FROM articles
WHERE site_name = $1;

-- name: CreateArticle :one
INSERT INTO articles (
    title, url, text, site_name, scraped_at, symbol
) VALUES (
             $1, $2, $3, $4, $5, $6
         ) ON CONFLICT (url)
    DO UPDATE SET
                  title = EXCLUDED.title,
                  text = EXCLUDED.text,
                  site_name = EXCLUDED.site_name,
                  scraped_at = EXCLUDED.scraped_at,
                  symbol = EXCLUDED.symbol
RETURNING *;