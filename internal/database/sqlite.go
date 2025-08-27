package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/yamlforge/yamlforge/internal/parser"
)

type SQLiteDB struct {
	*DB
}

func NewSQLite(config *parser.DatabaseConfig) (Database, error) {
	db := &SQLiteDB{
		DB: &DB{
			config: config,
			dbType: parser.DatabaseSQLite,
		},
	}
	return db, nil
}

func (db *SQLiteDB) Connect() error {
	conn, err := sql.Open("sqlite3", db.config.Path)
	if err != nil {
		return fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	db.conn = conn

	if _, err := db.conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return nil
}

func (db *SQLiteDB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

func (db *SQLiteDB) CreateSchema(schema *parser.Schema) error {
	db.schema = schema

	for modelName, model := range schema.Models {
		if err := db.createTable(modelName, model); err != nil {
			return fmt.Errorf("failed to create table %s: %w", modelName, err)
		}
	}

	for modelName, model := range schema.Models {
		if err := db.createIndexes(modelName, model); err != nil {
			return fmt.Errorf("failed to create indexes for %s: %w", modelName, err)
		}
	}

	return nil
}

func (db *SQLiteDB) createTable(name string, model *parser.Model) error {
	var columns []string
	var constraints []string

	for _, field := range model.Fields {
		column := db.buildColumnDefinition(field)
		columns = append(columns, column)

		if field.Type == parser.FieldTypeRelation && field.RelatedTo != "" {
			constraint := fmt.Sprintf(
				"FOREIGN KEY (%s) REFERENCES %s(id)",
				db.quote(field.Name),
				db.quote(field.RelatedTo),
			)

			if field.OnDelete != "" {
				switch field.OnDelete {
				case "cascade":
					constraint += " ON DELETE CASCADE"
				case "restrict":
					constraint += " ON DELETE RESTRICT"
				case "set_null":
					constraint += " ON DELETE SET NULL"
				}
			}

			constraints = append(constraints, constraint)
		}
	}

	parts := append(columns, constraints...)

	query := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (\n  %s\n)",
		db.quote(name),
		strings.Join(parts, ",\n  "),
	)

	_, err := db.conn.Exec(query)
	return err
}

func (db *SQLiteDB) buildColumnDefinition(field parser.Field) string {
	parts := []string{
		db.quote(field.Name),
		db.getSQLiteType(field),
	}

	if field.Primary {
		if field.Type == parser.FieldTypeID {
			parts = append(parts, "PRIMARY KEY AUTOINCREMENT")
		} else {
			parts = append(parts, "PRIMARY KEY")
		}
	}

	if field.Required && !field.Primary {
		parts = append(parts, "NOT NULL")
	}

	if field.Unique && !field.Primary {
		parts = append(parts, "UNIQUE")
	}

	if field.Default != nil {
		defaultValue := db.formatDefaultValue(field.Default, field.Type)
		parts = append(parts, "DEFAULT "+defaultValue)
	}

	return strings.Join(parts, " ")
}

func (db *SQLiteDB) getSQLiteType(field parser.Field) string {
	switch field.Type {
	case parser.FieldTypeID:
		return "INTEGER"
	case parser.FieldTypeNumber:
		return "INTEGER"
	case parser.FieldTypeText, parser.FieldTypeEmail, parser.FieldTypePassword,
		parser.FieldTypePhone, parser.FieldTypeURL, parser.FieldTypeSlug,
		parser.FieldTypeEnum, parser.FieldTypeColor, parser.FieldTypeMarkdown,
		parser.FieldTypeJSON, parser.FieldTypeCurrency, parser.FieldTypeIP,
		parser.FieldTypeUUID, parser.FieldTypeDuration:
		if field.Max != nil && *field.Max < 255 {
			return fmt.Sprintf("VARCHAR(%d)", *field.Max)
		}
		return "TEXT"
	case parser.FieldTypeBoolean:
		return "BOOLEAN"
	case parser.FieldTypeDatetime, parser.FieldTypeDate, parser.FieldTypeTime:
		return "DATETIME"
	case parser.FieldTypeFile, parser.FieldTypeImage:
		return "TEXT"
	case parser.FieldTypeArray:
		return "TEXT"
	case parser.FieldTypeRelation:
		return "INTEGER"
	case parser.FieldTypeLocation:
		return "TEXT"
	default:
		return "TEXT"
	}
}

func (db *SQLiteDB) formatDefaultValue(value any, fieldType parser.FieldType) string {
	if value == "CURRENT_TIMESTAMP" {
		return "CURRENT_TIMESTAMP"
	}

	switch v := value.(type) {
	case string:
		return "'" + escapeSQL(v) + "'"
	case int, int64, float64:
		return fmt.Sprint(v)
	case bool:
		if v {
			return "1"
		}
		return "0"
	default:
		return "NULL"
	}
}

func (db *SQLiteDB) createIndexes(modelName string, model *parser.Model) error {
	for _, field := range model.Fields {
		if field.Index && !field.Primary && !field.Unique {
			indexName := fmt.Sprintf("idx_%s_%s", modelName, field.Name)
			query := fmt.Sprintf(
				"CREATE INDEX IF NOT EXISTS %s ON %s (%s)",
				db.quote(indexName),
				db.quote(modelName),
				db.quote(field.Name),
			)

			if _, err := db.conn.Exec(query); err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *SQLiteDB) Query(model string, params parser.QueryParams) ([]map[string]any, error) {
	query, args := db.buildSelectQuery(model, params)

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

	if results == nil {
		results = []map[string]any{}
	}

	return results, rows.Err()
}

func (db *SQLiteDB) Get(model string, id any) (map[string]any, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", db.quote(model))
	return db.executeQueryRow(query, []any{id})
}

func (db *SQLiteDB) Create(model string, data map[string]any) (any, error) {
	query, args := db.buildInsertQuery(model, data)

	result, err := db.conn.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return lastID, nil
}

func (db *SQLiteDB) Update(model string, id any, data map[string]any) error {
	query, args := db.buildUpdateQuery(model, id, data)

	_, err := db.conn.Exec(query, args...)
	return err
}

func (db *SQLiteDB) Delete(model string, id any) error {
	query, args := db.buildDeleteQuery(model, id)

	_, err := db.conn.Exec(query, args...)
	return err
}

func (db *SQLiteDB) Count(model string, filters []parser.Filter) (int64, error) {
	var parts []string
	var args []any

	tableName := strings.ToLower(model)
	parts = append(parts, "SELECT COUNT(*) FROM "+db.quote(tableName))

	if len(filters) > 0 {
		whereClauses := []string{}
		for _, filter := range filters {
			clause, arg := db.buildWhereClause(filter)
			whereClauses = append(whereClauses, clause)
			args = append(args, arg)
		}
		parts = append(parts, "WHERE "+strings.Join(whereClauses, " AND "))
	}

	query := strings.Join(parts, " ")

	var count int64
	row := db.conn.QueryRow(query, args...)
	err := row.Scan(&count)

	return count, err
}

func (db *SQLiteDB) buildSelectQuery(model string, params parser.QueryParams) (string, []any) {
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

	if params.Search != "" && db.schema != nil {
		if m, ok := db.schema.GetModel(model); ok && len(m.UI.List.Searchable) > 0 {
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
	} else {
		parts = append(parts, "ORDER BY "+db.quote("id")+" DESC")
	}

	if params.PageSize > 0 {
		limit := params.PageSize
		offset := (params.Page - 1) * params.PageSize
		parts = append(parts, fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset))
	}

	return strings.Join(parts, " "), args
}

func (db *SQLiteDB) buildWhereClause(filter parser.Filter) (string, any) {
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

func (db *SQLiteDB) scanRow(rows *sql.Rows) (map[string]any, error) {
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

func (db *SQLiteDB) executeQueryRow(query string, args []any) (map[string]any, error) {
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

func (db *SQLiteDB) GetConnection() *sql.DB {
	return db.conn
}

