CREATE TABLE `text` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `imageid` TEXT NOT NULL,
    `text` TEXT NOT NULL,
    `language` TEXT NOT NULL,
    `ip` TEXT NOT NULL,
    `flagged` INTEGER NOT NULL,
    `created` DATE DEFAULT (datetime('now'))
);
