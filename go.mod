module github.com/akula410/builder/v2

go 1.26.4

replace github.com/akula410/connect/v2 => ../connect

require github.com/akula410/connect/v2 v2.0.0-00010101000000-000000000000

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/go-sql-driver/mysql v1.10.0 // indirect
)
