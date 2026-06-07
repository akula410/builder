package sqlbuilder_test

import (
	"testing"

	sqlbuilder "github.com/akula410/builder/v2"
)

// ---- DATABASE DDL ----

func TestCreateDatabase(t *testing.T) {
	sql, args, err := sqlbuilder.CreateDatabase("app_db").
		IfNotExists().
		CharacterSet("utf8mb4").
		Collate("utf8mb4_unicode_ci").
		Build()
	assertNoErr(t, err)
	assertEqual(t, "CREATE DATABASE IF NOT EXISTS `app_db` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", sql)
	assertArgs(t, nil, args)
}

func TestDropDatabase(t *testing.T) {
	sql, _, err := sqlbuilder.DropDatabase("app_db").IfExists().Build()
	assertNoErr(t, err)
	assertEqual(t, "DROP DATABASE IF EXISTS `app_db`", sql)
}

func TestUseDatabase(t *testing.T) {
	sql, _, err := sqlbuilder.UseDatabase("app_db").Build()
	assertNoErr(t, err)
	assertEqual(t, "USE `app_db`", sql)
}

func TestCreateDatabase_InvalidCharset_Error(t *testing.T) {
	_, _, err := sqlbuilder.CreateDatabase("db").CharacterSet("malicious").Build()
	assertErr(t, err)
}

// ---- CREATE TABLE ----

func TestCreateTable_Simple(t *testing.T) {
	sql, args, err := sqlbuilder.CreateTable("users").
		IfNotExists().
		Column(sqlbuilder.Column("id", "BIGINT UNSIGNED").NotNull().AutoIncrement()).
		Column(sqlbuilder.Column("email", "VARCHAR(255)").NotNull()).
		Column(sqlbuilder.Column("created_at", "DATETIME").NotNull().DefaultRaw("CURRENT_TIMESTAMP")).
		PrimaryKey("id").
		UniqueIndex("uniq_users_email", "email").
		Engine("InnoDB").
		CharacterSet("utf8mb4").
		Collate("utf8mb4_unicode_ci").
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "CREATE TABLE IF NOT EXISTS `users`")
	assertContains(t, sql, "`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT")
	assertContains(t, sql, "`email` VARCHAR(255) NOT NULL")
	assertContains(t, sql, "PRIMARY KEY (`id`)")
	assertContains(t, sql, "UNIQUE INDEX `uniq_users_email` (`email`)")
	assertContains(t, sql, "ENGINE=InnoDB")
	assertContains(t, sql, "DEFAULT CHARSET=utf8mb4")
	assertContains(t, sql, "COLLATE=utf8mb4_unicode_ci")
	assertArgs(t, nil, args)
}

func TestCreateTable_CompositePrimaryKey(t *testing.T) {
	sql, _, err := sqlbuilder.CreateTable("order_items").
		Column(sqlbuilder.Column("order_id", "BIGINT UNSIGNED").NotNull()).
		Column(sqlbuilder.Column("item_id", "BIGINT UNSIGNED").NotNull()).
		PrimaryKey("order_id", "item_id").
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "PRIMARY KEY (`order_id`, `item_id`)")
}

func TestCreateTable_ForeignKey(t *testing.T) {
	fk := sqlbuilder.ForeignKey("fk_orders_user", []string{"user_id"}).
		References("users", []string{"id"}).
		OnDelete(sqlbuilder.Cascade).
		OnUpdate(sqlbuilder.Restrict)

	sql, _, err := sqlbuilder.CreateTable("orders").
		Column(sqlbuilder.Column("id", "BIGINT UNSIGNED").NotNull().AutoIncrement()).
		Column(sqlbuilder.Column("user_id", "BIGINT UNSIGNED").NotNull()).
		PrimaryKey("id").
		Index("idx_orders_user_id", "user_id").
		ForeignKey(fk).
		Build()

	assertNoErr(t, err)
	assertContains(t, sql, "CONSTRAINT `fk_orders_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)")
	assertContains(t, sql, "ON DELETE CASCADE")
	assertContains(t, sql, "ON UPDATE RESTRICT")
}

func TestCreateTable_NoColumns_Error(t *testing.T) {
	_, _, err := sqlbuilder.CreateTable("empty").Build()
	assertErr(t, err)
}

func TestCreateTable_InvalidEngine_Error(t *testing.T) {
	_, _, err := sqlbuilder.CreateTable("t").
		Column(sqlbuilder.Column("id", "INT")).
		Engine("NotAnEngine").
		Build()
	assertErr(t, err)
}

// ---- DROP TABLE ----

func TestDropTable_Single(t *testing.T) {
	sql, _, err := sqlbuilder.DropTable("users").IfExists().Build()
	assertNoErr(t, err)
	assertEqual(t, "DROP TABLE IF EXISTS `users`", sql)
}

func TestDropTable_Multiple(t *testing.T) {
	sql, _, err := sqlbuilder.DropTable("old_users", "old_orders").IfExists().Build()
	assertNoErr(t, err)
	assertEqual(t, "DROP TABLE IF EXISTS `old_users`, `old_orders`", sql)
}

// ---- TRUNCATE TABLE ----

func TestTruncateTable(t *testing.T) {
	sql, _, err := sqlbuilder.TruncateTable("logs").Build()
	assertNoErr(t, err)
	assertEqual(t, "TRUNCATE TABLE `logs`", sql)
}

// ---- RENAME TABLE ----

func TestRenameTable_Single(t *testing.T) {
	sql, _, err := sqlbuilder.RenameTable("users_old", "users_archive").Build()
	assertNoErr(t, err)
	assertEqual(t, "RENAME TABLE `users_old` TO `users_archive`", sql)
}

func TestRenameTable_Multiple(t *testing.T) {
	sql, _, err := sqlbuilder.RenameTables(
		sqlbuilder.TableRename("users_old", "users_archive"),
		sqlbuilder.TableRename("orders_old", "orders_archive"),
	).Build()
	assertNoErr(t, err)
	assertEqual(t, "RENAME TABLE `users_old` TO `users_archive`, `orders_old` TO `orders_archive`", sql)
}

// ---- ALTER TABLE: COLUMN operations ----

func TestAlterTable_AddColumn_Nullable(t *testing.T) {
	sql, _, err := sqlbuilder.AlterTable("users").
		AddColumn(sqlbuilder.Column("phone", "VARCHAR(32)").Nullable()).
		Build()
	assertNoErr(t, err)
	assertEqual(t, "ALTER TABLE `users`\nADD COLUMN `phone` VARCHAR(32) NULL", sql)
}

func TestAlterTable_AddColumn_NotNull_Default(t *testing.T) {
	sql, _, err := sqlbuilder.AlterTable("users").
		AddColumn(sqlbuilder.Column("status", "VARCHAR(32)").NotNull().Default("active")).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "ADD COLUMN `status` VARCHAR(32) NOT NULL DEFAULT 'active'")
}

func TestAlterTable_AddColumn_DefaultRaw(t *testing.T) {
	sql, _, err := sqlbuilder.AlterTable("users").
		AddColumn(sqlbuilder.Column("created_at", "DATETIME").NotNull().DefaultRaw("CURRENT_TIMESTAMP")).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "ADD COLUMN `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP")
}

func TestAlterTable_AddColumn_After(t *testing.T) {
	sql, _, err := sqlbuilder.AlterTable("users").
		AddColumn(sqlbuilder.Column("middle_name", "VARCHAR(255)").Nullable().After("first_name")).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "AFTER `first_name`")
}

func TestAlterTable_AddColumn_First(t *testing.T) {
	sql, _, err := sqlbuilder.AlterTable("users").
		AddColumn(sqlbuilder.Column("uuid", "CHAR(36)").NotNull().First()).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "ADD COLUMN `uuid` CHAR(36) NOT NULL FIRST")
}

func TestAlterTable_DropColumn(t *testing.T) {
	sql, _, err := sqlbuilder.AlterTable("users").
		DropColumn("phone").
		Build()
	assertNoErr(t, err)
	assertEqual(t, "ALTER TABLE `users`\nDROP COLUMN `phone`", sql)
}

func TestAlterTable_MultipleOps(t *testing.T) {
	sql, _, err := sqlbuilder.AlterTable("users").
		DropColumn("phone").
		DropColumn("middle_name").
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "DROP COLUMN `phone`")
	assertContains(t, sql, "DROP COLUMN `middle_name`")
}

func TestAlterTable_ModifyColumn(t *testing.T) {
	sql, _, err := sqlbuilder.AlterTable("users").
		ModifyColumn(sqlbuilder.Column("phone", "VARCHAR(64)").Nullable()).
		Build()
	assertNoErr(t, err)
	assertEqual(t, "ALTER TABLE `users`\nMODIFY COLUMN `phone` VARCHAR(64) NULL", sql)
}

func TestAlterTable_ChangeColumn(t *testing.T) {
	sql, _, err := sqlbuilder.AlterTable("users").
		ChangeColumn("old_phone", sqlbuilder.Column("phone", "VARCHAR(64)").Nullable()).
		Build()
	assertNoErr(t, err)
	assertEqual(t, "ALTER TABLE `users`\nCHANGE COLUMN `old_phone` `phone` VARCHAR(64) NULL", sql)
}

// ---- ALTER TABLE: INDEX operations ----

func TestAlterTable_AddIndex(t *testing.T) {
	sql, _, err := sqlbuilder.AlterTable("users").
		AddIndex("idx_users_email", "email").
		Build()
	assertNoErr(t, err)
	assertEqual(t, "ALTER TABLE `users`\nADD INDEX `idx_users_email` (`email`)", sql)
}

func TestAlterTable_AddUniqueIndex(t *testing.T) {
	sql, _, err := sqlbuilder.AlterTable("users").
		AddUniqueIndex("uniq_users_email", "email").
		Build()
	assertNoErr(t, err)
	assertEqual(t, "ALTER TABLE `users`\nADD UNIQUE INDEX `uniq_users_email` (`email`)", sql)
}

func TestAlterTable_AddCompositeIndex(t *testing.T) {
	sql, _, err := sqlbuilder.AlterTable("orders").
		AddIndex("idx_orders_user_status", "user_id", "status").
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "ADD INDEX `idx_orders_user_status` (`user_id`, `status`)")
}

func TestAlterTable_AddFullTextIndex(t *testing.T) {
	sql, _, err := sqlbuilder.AlterTable("products").
		AddFullTextIndex("ft_products_name", "name", "description").
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "ADD FULLTEXT INDEX `ft_products_name` (`name`, `description`)")
}

func TestAlterTable_AddSpatialIndex(t *testing.T) {
	sql, _, err := sqlbuilder.AlterTable("locations").
		AddSpatialIndex("sp_locations_point", "point").
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "ADD SPATIAL INDEX `sp_locations_point` (`point`)")
}

func TestAlterTable_DropIndex(t *testing.T) {
	sql, _, err := sqlbuilder.AlterTable("users").
		DropIndex("idx_users_email").
		Build()
	assertNoErr(t, err)
	assertEqual(t, "ALTER TABLE `users`\nDROP INDEX `idx_users_email`", sql)
}

func TestAlterTable_AddIndexWithPrefixLength(t *testing.T) {
	sql, _, err := sqlbuilder.AlterTable("products").
		AddIndexColumns("idx_products_name", sqlbuilder.IndexColumn("name").Length(100)).
		Build()
	assertNoErr(t, err)
	assertContains(t, sql, "ADD INDEX `idx_products_name` (`name`(100))")
}

func TestAlterTable_EmptyOps_Error(t *testing.T) {
	_, _, err := sqlbuilder.AlterTable("users").Build()
	assertErr(t, err)
}

// ---- Column type validation ----

func TestColumn_InvalidType_Error(t *testing.T) {
	col := sqlbuilder.Column("x", "NOTATYPE")
	_, err := col.BuildSQL()
	assertErr(t, err)
}

func TestColumn_RawType_Bypasses(t *testing.T) {
	col := sqlbuilder.Column("x", sqlbuilder.RawType("CUSTOM_TYPE"))
	_, err := col.BuildSQL()
	assertNoErr(t, err)
}
