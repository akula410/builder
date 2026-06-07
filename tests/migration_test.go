package sqlbuilder_test

import (
	"testing"

	sqlbuilder "github.com/akula410/builder/v2"
)

func TestMigration_BuildUp(t *testing.T) {
	m := sqlbuilder.NewMigration("20260607_create_users").
		Up(
			sqlbuilder.CreateTable("users").
				IfNotExists().
				Column(sqlbuilder.Column("id", "BIGINT UNSIGNED").NotNull().AutoIncrement()).
				Column(sqlbuilder.Column("email", "VARCHAR(255)").NotNull()).
				PrimaryKey("id").
				UniqueIndex("uniq_users_email", "email"),
		).
		Down(
			sqlbuilder.DropTable("users").IfExists(),
		)

	steps, err := m.BuildUp()
	assertNoErr(t, err)
	if len(steps) != 1 {
		t.Fatalf("expected 1 Up step, got %d", len(steps))
	}
	assertContains(t, steps[0].SQL, "CREATE TABLE IF NOT EXISTS `users`")
}

func TestMigration_BuildDown(t *testing.T) {
	m := sqlbuilder.NewMigration("20260607_create_users").
		Up(
			sqlbuilder.CreateTable("users").
				Column(sqlbuilder.Column("id", "INT").NotNull()),
		).
		Down(
			sqlbuilder.DropTable("users").IfExists(),
		)

	steps, err := m.BuildDown()
	assertNoErr(t, err)
	if len(steps) != 1 {
		t.Fatalf("expected 1 Down step, got %d", len(steps))
	}
	assertEqual(t, "DROP TABLE IF EXISTS `users`", steps[0].SQL)
}

func TestMigration_MultipleSteps(t *testing.T) {
	m := sqlbuilder.NewMigration("20260607_two_tables").
		Up(
			sqlbuilder.CreateTable("a").Column(sqlbuilder.Column("id", "INT").NotNull()),
			sqlbuilder.CreateTable("b").Column(sqlbuilder.Column("id", "INT").NotNull()),
		).
		Down(
			sqlbuilder.DropTable("b"),
			sqlbuilder.DropTable("a"),
		)

	ups, err := m.BuildUp()
	assertNoErr(t, err)
	if len(ups) != 2 {
		t.Fatalf("expected 2 Up steps, got %d", len(ups))
	}

	downs, err := m.BuildDown()
	assertNoErr(t, err)
	if len(downs) != 2 {
		t.Fatalf("expected 2 Down steps, got %d", len(downs))
	}
}

func TestMigration_Name(t *testing.T) {
	m := sqlbuilder.NewMigration("my_migration")
	if m.Name() != "my_migration" {
		t.Fatalf("expected name %q, got %q", "my_migration", m.Name())
	}
}

func TestMigration_StepError_ReportsIndex(t *testing.T) {
	m := sqlbuilder.NewMigration("bad_migration").
		Up(
			sqlbuilder.DropTable("ok"),
			sqlbuilder.DropTable(""), // empty name → build error
		)

	_, err := m.BuildUp()
	assertErr(t, err)
	assertContains(t, err.Error(), "step 2")
}
