CREATE TABLE corporation_messages (
    corporation_id INTEGER PRIMARY KEY NOT NULL,
    message TEXT NOT NULL,
    source_url TEXT,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by INTEGER,
    FOREIGN KEY (corporation_id) REFERENCES corporations (id) ON DELETE CASCADE,
    FOREIGN KEY (updated_by) REFERENCES characters (id) ON DELETE SET NULL
);
