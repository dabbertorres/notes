table "tags" {
  schema = schema.notes

  column "tag_id" {
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

  index "idx_tags_name" {
    columns = [column.name]
    unique  = true
  }
}

table "user_tag_access" {
  schema = schema.notes

  column "tag_id" {
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
      column.tag_id,
      column.user_id,
    ]
  }

  foreign_key "tag_id" {
    columns     = [column.tag_id]
    ref_columns = [table.tags.column.tag_id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "user_id" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.user_id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  index "idx_fk_user_tag_access_tag_id" {
    columns = [column.tag_id]
    unique  = false
  }

  index "idx_fk_user_tag_access_user_id" {
    columns = [column.user_id]
    unique  = false
  }

  unique "uniq_user_tag_access_tag_id_user_id" {
    columns = [
      column.tag_id,
      column.user_id,
    ]
  }
}
