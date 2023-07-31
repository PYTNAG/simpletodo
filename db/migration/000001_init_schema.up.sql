CREATE TABLE "users" (
  "id" serial PRIMARY KEY,
  "username" text UNIQUE NOT NULL,
  "hash" bytea NOT NULL
);

CREATE TABLE "lists" (
  "id" serial PRIMARY KEY,
  "author" int NOT NULL,
  "header" text NOT NULL
);

CREATE TABLE "tasks" (
  "id" serial PRIMARY KEY,
  "list_id" int NOT NULL,
  "parent_task" int,
  "task" text NOT NULL,
  "complete" bool NOT NULL DEFAULT FALSE
);

CREATE INDEX ON "lists" ("id");

CREATE INDEX ON "lists" ("author");

CREATE INDEX ON "lists" ("author", "header");

ALTER TABLE "lists" ADD FOREIGN KEY ("author") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "tasks" ADD FOREIGN KEY ("list_id") REFERENCES "lists" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "tasks" ADD FOREIGN KEY ("parent_task") REFERENCES "tasks" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
