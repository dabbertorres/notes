env "local" {
  src = [
    "file://schema.hcl",
    "file://notes.hcl",
    "file://tags.hcl",
    "file://users.hcl",
  ]

  url = "postgres://postgres:postgres@localhost:5437/postgres?sslmode=disable"

  dev = "docker://postgres/16/postgres"

  schemas = [
    "notes",
  ]
}

lint {
  concurrent_index {
    check_create = true
    check_drop   = true
    check_txmode = true

    error = true
  }

  data_depend {
    error = true
  }

  destructive {
    error = true
  }

  incompatible {
    error = true
  }

  naming {
    error   = true
    match   = "^[a-z_]+$"
    message = "must be lower snake case"

    index {
      match   = "^idx_[a-z_]+$"
      message = "must be lower snake case and start with 'idx_'"
    }
  }

  non_linear {
    error = true
  }
}
