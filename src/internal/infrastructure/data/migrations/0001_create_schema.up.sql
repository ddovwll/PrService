BEGIN;

CREATE TABLE teams
(
    name TEXT PRIMARY KEY
);

CREATE TABLE users
(
    id        TEXT PRIMARY KEY,
    username  TEXT    NOT NULL,
    team_name TEXT    NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT fk_users_team
        FOREIGN KEY (team_name)
            REFERENCES teams (name)
            ON UPDATE CASCADE
            ON DELETE RESTRICT
);

CREATE INDEX idx_users_team_name ON users (team_name);

CREATE TABLE pull_requests
(
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    author_id  TEXT NOT NULL,
    status     TEXT NOT NULL,
    created_at TIMESTAMPTZ,
    merged_at  TIMESTAMPTZ,

    CONSTRAINT fk_pull_requests_author
        FOREIGN KEY (author_id)
            REFERENCES users (id)
            ON UPDATE CASCADE
            ON DELETE RESTRICT,

    CONSTRAINT chk_pull_request_status
        CHECK (status IN ('OPEN', 'MERGED'))
);

CREATE TABLE pull_request_reviewers
(
    pull_request_id TEXT NOT NULL,
    reviewer_id     TEXT NOT NULL,

    PRIMARY KEY (pull_request_id, reviewer_id),

    CONSTRAINT fk_pr_reviewers_pr
        FOREIGN KEY (pull_request_id)
            REFERENCES pull_requests (id)
            ON UPDATE CASCADE
            ON DELETE CASCADE,

    CONSTRAINT fk_pr_reviewers_user
        FOREIGN KEY (reviewer_id)
            REFERENCES users (id)
            ON UPDATE CASCADE
            ON DELETE CASCADE
);

CREATE INDEX idx_pr_reviewers_reviewer_id ON pull_request_reviewers (reviewer_id);

COMMIT;