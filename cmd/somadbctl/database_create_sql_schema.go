package main

import (
	"fmt"
)

func createSQLSchema(printOnly bool, verbose bool) {
	idx := 0
	// map for storing the SQL statements by name
	queryMap := make(map[string]string)
	// slice storing the required statement order so foreign keys can
	// resolve successfully
	queries := make([]string, 100)

	queryMap["createSchemaSoma"] = `create schema if not exists soma;`
	queries[idx] = "createSchemaSoma"
	idx++

	queryMap["createSchemaInventory"] = `create schema if not exists inventory;`
	queries[idx] = "createSchemaInventory"
	idx++

	queryMap["createSchemaAuth"] = `create schema if not exists auth;`
	queries[idx] = "createSchemaAuth"
	idx++

	queryMap["createSchemaRoot"] = `create schema if not exists root;`
	queries[idx] = "createSchemaRoot"
	idx++

	queryMap["alterDatabaseDefaultSearchPath"] = fmt.Sprintf("alter database %s set search_path TO soma,inventory,auth;", Cfg.Database.Name)
	queries[idx] = "alterDatabaseDefaultSearchPath"
	idx++

	queryMap["alterDatabaseSearchPath"] = `set search_path to soma,inventory,auth;`
	queries[idx] = "alterDatabaseSearchPath"
	idx++

	queryMap["grantUsageSomaSvcSoma"] = `grant usage on schema soma to soma_svc;`
	queries[idx] = "grantUsageSomaSvcSoma"
	idx++

	queryMap["grantUsageSomaSvcInventory"] = `grant usage on schema inventory to soma_svc;`
	queries[idx] = "grantUsageSomaSvcInventory"
	idx++

	queryMap["grantUsageSomaSvcRoot"] = `grant usage on schema root to soma_svc;`
	queries[idx] = "grantUsageSomaSvcRoot"
	idx++

	queryMap["grantUsageSomaSvcAuth"] = `grant usage on schema auth to soma_svc;`
	queries[idx] = "grantUsageSomaSvcAuth"
	idx++

	queryMap["grantUsageSomaInvSoma"] = `grant usage on schema soma to soma_inv;`
	queries[idx] = "grantUsageSomaInvSoma"
	idx++

	queryMap["grantUsageSomaInvInventory"] = `grant usage on schema inventory to soma_inv;`
	queries[idx] = "grantUsageSomaInvInventory"
	idx++

	queryMap["grantUsageSomaInvAuth"] = `grant usage on schema auth to soma_inv;`
	queries[idx] = "grantUsageSomaInvAuth"
	idx++

	performDatabaseTask(printOnly, verbose, queries[:idx], queryMap)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
