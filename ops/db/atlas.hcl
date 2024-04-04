env "local" {
  src = [
    "file://notes.hcl",
  ]

  url = "postgres://postgres:postgres@localhost:5437/postgres?sslmode=disable"

  dev = "docker://postgres/16/postgres"

  schemas = [
    "notes",
  ]
}
