schema "notes" {
}

table "users" {
  schema = schema.notes

  column "user_id" {
    type = uuid
    null = false
  }

  column "name" {
    type = text
    null = false
  }

  column "created_at" {
    type = timestamptz
    null = false
  }

  column "last_sign_in" {
    type = timestamptz
    null = false
  }

  column "active" {
    type = bool
    null = false
  }

  primary_key {
    columns = [column.user_id]
  }
}

table "notes" {
  schema = schema.notes

  column "note_id" {
    type = uuid
    null = false
  }

  column "created_at" {
    type = timestamptz
    null = false
  }

  column "created_by" {
    type = uuid
    null = true
  }

  column "updated_at" {
    type = timestamptz
    null = false
  }

  column "updated_by" {
    type = uuid
    null = true
  }

  column "title" {
    type = text
    null = false
  }

  column "body" {
    type = text
    null = false
  }

  column "search_index" {
    type = tsvector
    as {
      expr = "to_tsvector('english', title || '\\n' || body)"
      type = STORED
    }
  }

  primary_key {
    columns = [column.note_id]
  }

  index "idx_note_text_search" {
    type = GIN
    columns = [column.search_index]
  }
}

enum "access_level" {
  schema = schema.notes
  values = [
    "owner",
    "editor",
    "viewer",
  ]
}

table "user_note_access" {
  schema = schema.notes

  column "note_id" {
    type = uuid
    null = false
  }

  column "user_id" {
    type = uuid
    null = false
  }

  column "access" {
    type = enum.access_level
    null = false
  }

  primary_key {
    columns = [
      column.note_id,
      column.user_id,
    ]
  }

  foreign_key "note_id" {
    columns     = [column.note_id]
    ref_columns = [table.notes.column.note_id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "user_id" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.user_id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}

table "tags" {
  schema = schema.notes

  column "tag_id" {
    type = uuid
    null = false
  }

  column "user_id" {
    type = uuid
    null = false
  }

  column "name" {
    type = text
    null = false
  }

  check "non empty name" {
    expr = "LENGTH(name) > 0"
  }

  primary_key {
    columns = [column.tag_id]
  }

  foreign_key "user_id" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.user_id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  index "idx_unique_user_id_name" {
    columns = [
      column.user_id,
      column.name,
    ]
    unique = true
  }
}

table "note_tags" {
  schema = schema.notes

  column "note_id" {
    type = uuid
    null = false
  }

  column "tag_id" {
    type = uuid
    null = false
  }

  primary_key {
    columns = [
      column.note_id,
      column.tag_id,
    ]
  }

  foreign_key "note_id" {
    columns     = [column.note_id]
    ref_columns = [table.notes.column.note_id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "tag_id" {
    columns     = [column.tag_id]
    ref_columns = [table.tags.column.tag_id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}
