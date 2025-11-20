-- name: GetCorporationMessage :one
SELECT
    cm.corporation_id,
    cm.message,
    cm.source_url,
    cm.updated_at,
    cm.updated_by,
    ec.name AS updated_by_name
FROM
    corporation_messages cm
    LEFT JOIN eve_characters ec ON cm.updated_by = ec.id
WHERE
    cm.corporation_id = ?;

-- name: UpdateCorporationMessage :one
INSERT INTO corporation_messages (
    corporation_id,
    message,
    source_url,
    updated_at,
    updated_by
)
VALUES (?1, ?2, ?3, ?4, ?5)
ON CONFLICT (corporation_id) DO
UPDATE SET
    message = excluded.message,
    source_url = excluded.source_url,
    updated_at = excluded.updated_at,
    updated_by = excluded.updated_by
RETURNING corporation_id, message, source_url, updated_at, updated_by;
