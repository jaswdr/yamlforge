package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/yamlforge/yamlforge/internal/parser"
)

type Database interface {
	Connect() error
	Close() error
	CreateSchema(schema *parser.Schema) error
	Query(model string, params parser.QueryParams) ([]map[string]any, error)
	Get(model string, id any) (map[string]any, error)
	Create(model string, data map[string]any) (any, error)
	Update(model string, id any, data map[string]any) error
	Delete(model string, id any) error
	Count(model string, filters []parser.Filter) (int64, error)
	BeginTx() (*sql.Tx, error)
}

type DB struct {
	config *parser.DatabaseConfig
	conn   *sql.DB
	dbType parser.DatabaseType
	schema *parser.Schema
}

func New(config *parser.DatabaseConfig) (Database, error) {
	return &DB{
		config: config,
		dbType: parser.DatabaseType(config.Type),
	}, nil
}

func (db *DB) getDriverName() string {
	switch db.dbType {
	case parser.DatabaseSQLite:
		return "sqlite3"
	default:
		return ""
	}
}

func (db *DB) getConnectionString() string {
	switch db.dbType {
	case parser.DatabaseSQLite:
		return db.config.Path
	default:
		return ""
	}
}

func (db *DB) BeginTx() (*sql.Tx, error) {
	return db.conn.Begin()
}

func (db *DB) buildSelectQuery(model string, params parser.QueryParams) (string, []any) {
	m, ok := db.schema.GetModel(model)
	if !ok {
		return "", nil
	}

	var parts []string
	var args []any

	parts = append(parts, "SELECT * FROM "+db.quote(model))

	if len(params.Filters) > 0 {
		whereClauses := []string{}
		for _, filter := range params.Filters {
			clause, arg := db.buildWhereClause(filter)
			whereClauses = append(whereClauses, clause)
			args = append(args, arg)
		}
		parts = append(parts, "WHERE "+strings.Join(whereClauses, " AND "))
	}

	if params.Search != "" && len(m.UI.List.Searchable) > 0 {
		searchClauses := []string{}
		for _, field := range m.UI.List.Searchable {
			searchClauses = append(searchClauses, db.quote(field)+" LIKE ?")
			args = append(args, "%"+params.Search+"%")
		}
		if len(params.Filters) > 0 {
			parts = append(parts, "AND ("+strings.Join(searchClauses, " OR ")+")")
		} else {
			parts = append(parts, "WHERE "+strings.Join(searchClauses, " OR "))
		}
	}

	if len(params.Sort) > 0 {
		orderClauses := []string{}
		for _, sort := range params.Sort {
			order := "ASC"
			if sort.Desc {
				order = "DESC"
			}
			orderClauses = append(orderClauses, db.quote(sort.Field)+" "+order)
		}
		parts = append(parts, "ORDER BY "+strings.Join(orderClauses, ", "))
	}

	if params.PageSize > 0 {
		limit := params.PageSize
		offset := (params.Page - 1) * params.PageSize
		parts = append(parts, fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset))
	}

	return strings.Join(parts, " "), args
}

func (db *DB) buildWhereClause(filter parser.Filter) (string, any) {
	operator := filter.Operator
	if operator == "" {
		operator = "="
	}

	switch operator {
	case "like":
		return db.quote(filter.Field) + " LIKE ?", "%" + fmt.Sprint(filter.Value) + "%"
	case "in":
		values := filter.Value.([]any)
		placeholders := make([]string, len(values))
		for i := range values {
			placeholders[i] = "?"
		}
		return db.quote(filter.Field) + " IN (" + strings.Join(placeholders, ",") + ")", values
	default:
		return db.quote(filter.Field) + " " + operator + " ?", filter.Value
	}
}

func (db *DB) quote(name string) string {
	switch db.dbType {
	case parser.DatabaseSQLite:
		return "\"" + name + "\""
	default:
		return name
	}
}

func (db *DB) buildInsertQuery(model string, data map[string]any) (string, []any) {
	var columns []string
	var placeholders []string
	var args []any

	for col, val := range data {
		columns = append(columns, db.quote(col))
		placeholders = append(placeholders, "?")
		args = append(args, val)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		db.quote(model),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	return query, args
}

func (db *DB) buildUpdateQuery(model string, id any, data map[string]any) (string, []any) {
	var setClauses []string
	var args []any

	for col, val := range data {
		setClauses = append(setClauses, db.quote(col)+" = ?")
		args = append(args, val)
	}

	args = append(args, id)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = ?",
		db.quote(model),
		strings.Join(setClauses, ", "),
	)

	return query, args
}

func (db *DB) buildDeleteQuery(model string, id any) (string, []any) {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", db.quote(model))
	return query, []any{id}
}

func (db *DB) scanRow(rows *sql.Rows) (map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	result := make(map[string]any)
	for i, col := range columns {
		val := values[i]
		if b, ok := val.([]byte); ok {
			result[col] = string(b)
		} else {
			result[col] = val
		}
	}

	return result, nil
}

func (db *DB) executeQuery(query string, args []any) ([]map[string]any, error) {
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		row, err := db.scanRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, row)
	}

	return results, rows.Err()
}

func (db *DB) executeQueryRow(query string, args []any) (map[string]any, error) {
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	return db.scanRow(rows)
}

func escapeSQL(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func (db *DB) Connect() error {
	return fmt.Errorf("Connect not implemented for base DB type")
}

func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

func (db *DB) CreateSchema(schema *parser.Schema) error {
	return fmt.Errorf("CreateSchema not implemented for base DB type")
}

func (db *DB) Query(model string, params parser.QueryParams) ([]map[string]any, error) {
	return nil, fmt.Errorf("Query not implemented for base DB type")
}

func (db *DB) Get(model string, id any) (map[string]any, error) {
	return nil, fmt.Errorf("Get not implemented for base DB type")
}

func (db *DB) Create(model string, data map[string]any) (any, error) {
	return nil, fmt.Errorf("Create not implemented for base DB type")
}

func (db *DB) Update(model string, id any, data map[string]any) error {
	return fmt.Errorf("Update not implemented for base DB type")
}

func (db *DB) Delete(model string, id any) error {
	return fmt.Errorf("Delete not implemented for base DB type")
}

func (db *DB) Count(model string, filters []parser.Filter) (int64, error) {
	return 0, fmt.Errorf("Count not implemented for base DB type")
}

