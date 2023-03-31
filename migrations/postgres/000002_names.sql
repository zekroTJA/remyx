-- +goose Up

ALTER TABLE "remyx"
ADD COLUMN "name" VARCHAR;

-- +goose Down

ALTER TABLE "remyx"
DROP COLUMN "name";