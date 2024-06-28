-- Drop schema named "public"
DROP SCHEMA IF EXISTS "public" CASCADE;
-- Add new schema named "notes"
CREATE SCHEMA "notes";
-- Create enum type "access_level"
CREATE TYPE "notes"."access_level" AS ENUM ('owner', 'editor', 'viewer');
-- Create "notes" table
CREATE TABLE "notes"."notes" ("note_id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "created_by" uuid NULL, "updated_at" timestamptz NOT NULL, "updated_by" uuid NULL, "title" text NOT NULL, "body" text NOT NULL, "search_index" tsvector NOT NULL GENERATED ALWAYS AS (to_tsvector('english'::regconfig, ((title || '\n'::text) || body))) STORED, PRIMARY KEY ("note_id"));
-- Create index "idx_note_text_search" to table: "notes"
CREATE INDEX "idx_note_text_search" ON "notes"."notes" USING gin ("search_index");
-- Create "users" table
CREATE TABLE "notes"."users" ("user_id" uuid NOT NULL, "name" text NOT NULL, "created_at" timestamptz NOT NULL, "last_sign_in" timestamptz NOT NULL, "active" boolean NOT NULL, PRIMARY KEY ("user_id"));
-- Create "tags" table
CREATE TABLE "notes"."tags" ("tag_id" uuid NOT NULL, "ordered_id" bigint NOT NULL GENERATED ALWAYS AS IDENTITY, "user_id" uuid NOT NULL, "name" text NOT NULL, PRIMARY KEY ("tag_id"), CONSTRAINT "user_id" FOREIGN KEY ("user_id") REFERENCES "notes"."users" ("user_id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "non empty name" CHECK (length(name) > 0));
-- Create index "idx_unique_ordered_id" to table: "tags"
CREATE UNIQUE INDEX "idx_unique_ordered_id" ON "notes"."tags" ("ordered_id");
-- Create index "idx_unique_user_id_name" to table: "tags"
CREATE UNIQUE INDEX "idx_unique_user_id_name" ON "notes"."tags" ("user_id", "name");
-- Create "note_tags" table
CREATE TABLE "notes"."note_tags" ("note_id" uuid NOT NULL, "tag_id" uuid NOT NULL, PRIMARY KEY ("note_id", "tag_id"), CONSTRAINT "note_id" FOREIGN KEY ("note_id") REFERENCES "notes"."notes" ("note_id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "tag_id" FOREIGN KEY ("tag_id") REFERENCES "notes"."tags" ("tag_id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "user_note_access" table
CREATE TABLE "notes"."user_note_access" ("note_id" uuid NOT NULL, "user_id" uuid NOT NULL, "access" "notes"."access_level" NOT NULL, PRIMARY KEY ("note_id", "user_id"), CONSTRAINT "note_id" FOREIGN KEY ("note_id") REFERENCES "notes"."notes" ("note_id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "user_id" FOREIGN KEY ("user_id") REFERENCES "notes"."users" ("user_id") ON UPDATE NO ACTION ON DELETE CASCADE);
