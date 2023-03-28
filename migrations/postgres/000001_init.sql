-- +goose Up

CREATE TABLE "session" (
    "uid" VARCHAR NOT NULL,
    "created_at" TIMESTAMP NOT NULL,
    "user_id" VARCHAR NOT NULL,
    "refresh_token" VARCHAR NOT NULL,
    PRIMARY KEY ("uid")
);

CREATE TABLE "remyx" (
    "uid" VARCHAR NOT NULL,
    "created_at" TIMESTAMP NOT NULL,
    "creator_uid" VARCHAR NOT NULL,
    "head" INT NOT NULL,
    PRIMARY KEY ("uid")
);

CREATE TABLE "source_playlist" (
    "remyx_uid" VARCHAR NOT NULL,
    "playlist_uid" VARCHAR NOT NULL,
    "user_uid" VARCHAR NOT NULL,
    PRIMARY KEY ("remyx_uid", "playlist_uid", "user_uid"),
    FOREIGN KEY ("remyx_uid") 
        REFERENCES "remyx" ("uid")
        ON DELETE CASCADE
);

CREATE TABLE "target_playlist" (
    "remyx_uid" VARCHAR NOT NULL,
    "playlist_uid" VARCHAR NOT NULL,
    "user_uid" VARCHAR NOT NULL,
    PRIMARY KEY ("remyx_uid", "user_uid"),
    FOREIGN KEY ("remyx_uid") 
        REFERENCES "remyx" ("uid")
        ON DELETE CASCADE
);

-- +goose Down

DROP TABLE "source_playlist";
DROP TABLE "target_playlist";
DROP TABLE "remyx";
DROP TABLE "session";