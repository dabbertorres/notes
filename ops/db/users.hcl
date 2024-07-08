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

enum "access_level" {
  schema = schema.notes
  values = [
    "owner",
    "editor",
    "viewer",
  ]
}
