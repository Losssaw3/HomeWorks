package main

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type DbResponse map[string]any

type DbHelper struct {
	db     *sql.DB
	tables map[string]bool
}

type ColMeta struct {
	dataType string
	isNull   bool
}

func NewDbHelper(db *sql.DB, tables []string) (*DbHelper, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	tablesMap := make(map[string]bool, len(tables))
	for _, table := range tables {
		tablesMap[table] = true
	}
	return &DbHelper{db: db, tables: tablesMap}, nil
}

func getTableMetadata(helper *DbHelper, table string) (pk string, columns map[string]ColMeta, err error) {
	columns = make(map[string]ColMeta)
	if ok := helper.tables[table]; !ok {
		return "", nil, fmt.Errorf("unknown table")
	}
	rows, err := helper.db.Query(fmt.Sprintf("SHOW FULL COLUMNS FROM `%s`", table))
	if err != nil {
		return "", nil, err
	}
	defer rows.Close()
	// SHOW FULL COLUMNS возвращает: Field, Type, Collation, Null, Key, Default, Extra, Privileges, Comment
	for rows.Next() {
		var field, colType, collation, null, key, extra, privileges, comment sql.NullString
		var def sql.NullString

		err := rows.Scan(&field, &colType, &collation, &null, &key, &def, &extra, &privileges, &comment)
		if err != nil {
			return "", nil, err
		}

		colName := field.String
		meta := ColMeta{
			isNull: (null.String == "YES"),
		}

		rawType := colType.String
		if strings.Contains(rawType, "int") {
			meta.dataType = "int"
		} else if strings.Contains(rawType, "float") || strings.Contains(rawType, "double") {
			meta.dataType = "float"
		} else {
			meta.dataType = "string"
		}
		columns[colName] = meta
		if key.String == "PRI" {
			pk = colName
		}
	}
	return pk, columns, nil
}

func getTablesList(db *sql.DB) (DbResponse, error) {
	query := `SHOW TABLES`
	rows, err := db.Query(query)
	if err != nil {
		return DbResponse{"error": "Internal database error"}, err
	}
	var tables []string
	var table string
	for rows.Next() {
		err = rows.Scan(&table)
		if err != nil {
			return DbResponse{"error": "Internal database error"}, err
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return DbResponse{"error": "Internal database error"}, err
	}
	response := DbResponse{
		"tables": tables,
	}
	return DbResponse{"response": response}, nil

}

func getDataFromTable(helper *DbHelper, table string, limit, offset int) (DbResponse, error) {
	if ok := helper.tables[table]; !ok {
		return DbResponse{"error": "unknown table"}, fmt.Errorf("no table with name %s", table)
	}
	query := fmt.Sprintf("SELECT * FROM `%s` LIMIT ? OFFSET ?", table)
	rows, err := helper.db.Query(query, limit, offset)
	if err != nil {
		return DbResponse{"error": "Internal database error"}, err
	}
	defer rows.Close()
	var records []map[string]any
	cols, _ := rows.Columns()

	for rows.Next() {
		columms := make([]any, len(cols))
		ptr := make([]any, len(cols))
		for i := range columms {
			ptr[i] = &columms[i]
		}

		err := rows.Scan(ptr...)
		if err != nil {
			return DbResponse{"error": "Internal database error"}, err
		}

		res := make(map[string]any)
		for i, colName := range cols {
			val := columms[i]
			if b, ok := val.([]byte); ok {
				res[colName] = string(b)
			} else {
				res[colName] = val
			}
		}
		records = append(records, res)
	}
	response := DbResponse{
		"records": records,
	}
	return DbResponse{"response": response}, nil
}

func getDataFromTableById(helper *DbHelper, table string, id int) (DbResponse, error) {
	if ok := helper.tables[table]; !ok {
		return DbResponse{"error": "unknown table"}, fmt.Errorf("no table found with name %s", table)
	}
	pkName, _, err := getTableMetadata(helper, table)
	query := fmt.Sprintf("SELECT * FROM `%s` WHERE %s = ? LIMIT 1", table, pkName)
	rows, err := helper.db.Query(query, id)
	if err != nil {
		return DbResponse{"error": "Internal database error"}, err
	}
	defer rows.Close()
	res := make(map[string]any)
	if !rows.Next() {
		return DbResponse{"error": "record not found"}, fmt.Errorf("row with id %d not found", id)
	} else {

		cols, _ := rows.Columns()
		columms := make([]any, len(cols))
		ptr := make([]any, len(cols))
		for i := range columms {
			ptr[i] = &columms[i]
		}

		err := rows.Scan(ptr...)
		if err != nil {
			return DbResponse{"status": "fail"}, err
		}

		for i, colName := range cols {
			val := columms[i]
			if b, ok := val.([]byte); ok {
				res[colName] = string(b)
			} else {
				res[colName] = val
			}
		}
	}
	if err := rows.Err(); err != nil {
		return DbResponse{"error": "Internal database error"}, err
	}
	response := DbResponse{
		"record": res,
	}
	return DbResponse{"response": response}, nil
}

func PutRecordIntoTable(helper *DbHelper, table string, payload map[string]any) (DbResponse, error) {
	if ok := helper.tables[table]; !ok {
		return DbResponse{"status": "not found"}, fmt.Errorf("no table found with name %s", table)
	}
	pkName, allowedCols, err := getTableMetadata(helper, table)
	if err != nil {
		return DbResponse{"error": "Internal database error"}, err
	}

	cols := []string{}
	vals := []any{}
	placeholders := []string{}

	for col, meta := range allowedCols {
		val, ok := payload[col]
		if col == pkName {
			continue
		}
		if !ok {
			if meta.isNull {
				vals = append(vals, nil)
			} else {
				switch meta.dataType {
				case "int":
					vals = append(vals, 0)
				case "float":
					vals = append(vals, 0.0)
				default:
					vals = append(vals, "")
				}
			}
		} else {
			vals = append(vals, val)
		}

		cols = append(cols, fmt.Sprintf("`%s`", col))
		placeholders = append(placeholders, "?")

	}
	query := fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES (%s)",
		table,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)
	result, err := helper.db.Exec(query, vals...)
	if err != nil {
		return DbResponse{"error": "Internal database error"}, fmt.Errorf("insert error: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return DbResponse{"error": "Internal database error"}, fmt.Errorf("error getting last inserted id: %w", err) //TODO cast to dbresponse
	}
	response := DbResponse{
		pkName: id,
	}
	return DbResponse{"response": response}, nil
}

func UpdateRecordInTable(helper *DbHelper, table string, Id int, payload map[string]any) (DbResponse, error) {
	if ok := helper.tables[table]; !ok {
		return DbResponse{"error": "unknown table"}, fmt.Errorf("no table found with name %s", table)
	}
	var setParts []string
	var vals []any

	pkName, allowedCols, err := getTableMetadata(helper, table)
	if err != nil {
		return DbResponse{"error": "Internal database error"}, err
	}
	for col, val := range payload {
		meta, ok := allowedCols[col]
		if !ok {
			continue
		}
		if col == pkName {
			return DbResponse{"error": fmt.Sprintf("field %s have invalid type", col)}, fmt.Errorf("invalid field type")
		}
		if val == nil {
			if !meta.isNull {
				return DbResponse{"error": fmt.Sprintf("field %s have invalid type", col)}, fmt.Errorf("invalid type")
			}
		} else {
			switch meta.dataType {
			case "int":
				if _, ok := val.(float64); !ok {
					return DbResponse{"error": fmt.Sprintf("field %s have invalid type", col)}, fmt.Errorf("invalid type")
				}
			case "float":
				if _, ok := val.(float64); !ok {
					return DbResponse{"error": fmt.Sprintf("field %s have invalid type", col)}, fmt.Errorf("invalid type")
				}
			case "string":
				if _, ok := val.(string); !ok {
					return DbResponse{"error": fmt.Sprintf("field %s have invalid type", col)}, fmt.Errorf("invalid type")
				}
			}
		}
		setParts = append(setParts, fmt.Sprintf("`%s` = ?", col))
		vals = append(vals, val)
	}
	vals = append(vals, Id)

	query := fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s` = ?",
		table,
		strings.Join(setParts, ", "),
		pkName,
	)

	result, err := helper.db.Exec(query, vals...)
	if err != nil {
		return DbResponse{"error": "Internal database error"}, fmt.Errorf("insert error: %w", err) //TODO cast to dbresponse
	}
	num, err := result.RowsAffected()
	if err != nil {
		return DbResponse{"error": "Internal database error"}, fmt.Errorf("error getting affected rows num: %w", err) //TODO cast to dbresponse
	}
	response := DbResponse{
		"updated": num,
	}
	return DbResponse{"response": response}, nil
}

func DeleteRecordFromTable(helper *DbHelper, table string, Id int) (DbResponse, error) {
	if ok := helper.tables[table]; !ok {
		return DbResponse{"error": "unknown table"}, fmt.Errorf("no table found with name %s", table)
	}
	pkName, _, err := getTableMetadata(helper, table)
	if err != nil {
		return DbResponse{"error": "Internal database error"}, err
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", table, pkName)
	result, err := helper.db.Exec(query, Id)
	if err != nil {
		return DbResponse{"error": "Internal database error"}, fmt.Errorf("delete error: %w", err) //TODO cast to dbresponse
	}
	num, err := result.RowsAffected()
	if err != nil {
		return DbResponse{"error": "Internal database error"}, fmt.Errorf("error getting affected rows num: %w", err) //TODO cast to dbresponse
	}
	response := DbResponse{
		"deleted": num,
	}
	return DbResponse{"response": response}, nil
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	tables, err := getTablesList(db)
	if err != nil {
		return nil, err
	}
	innerResponse := tables["response"].(DbResponse)
	tablesList := innerResponse["tables"].([]string)
	helper, err := NewDbHelper(db, tablesList)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		tables, err := getTablesList(helper.db)
		if err != nil {
			if status := tables["error"]; status == "Internal database error" {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			} else if status == "unknown table" {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(tables)
				return
			}

		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tables)
	})

	mux.HandleFunc("GET /{table}", func(w http.ResponseWriter, r *http.Request) {
		tableName := r.PathValue("table")
		query := r.URL.Query()

		limitStr := query.Get("limit")
		offsetStr := query.Get("offset")

		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			limit = 5
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			offset = 0
		}
		data, err := getDataFromTable(helper, tableName, limit, offset)

		if err != nil {
			if status := data["error"]; status == "Internal database error" {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			} else if status == "unknown table" {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(data)
				return
			}

		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	mux.HandleFunc("GET /{table}/{id}", func(w http.ResponseWriter, r *http.Request) {
		tableName := r.PathValue("table")
		IdString := r.PathValue("id")
		Id, err := strconv.Atoi(IdString)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(err.Error())
			return
		}

		data, err := getDataFromTableById(helper, tableName, Id)
		if err != nil {
			if status := data["error"]; status == "Internal database error" {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			} else if status == "unknown table" {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(data)
				return
			} else if status == "record not found" {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(data)
				return
			}

		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	mux.HandleFunc("PUT /{table}/", func(w http.ResponseWriter, r *http.Request) {
		tableName := r.PathValue("table")
		var payload map[string]any
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(err.Error())
			return
		}
		id, err := PutRecordIntoTable(helper, tableName, payload)
		if err != nil {
			if status := id["error"]; status == "Internal database error" {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(id)
			} else if status == "unknown table" {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(id)
			} else {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(id)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(id)
	})

	mux.HandleFunc("POST /{table}/{id}", func(w http.ResponseWriter, r *http.Request) {
		tableName := r.PathValue("table")
		IdString := r.PathValue("id")
		Id, err := strconv.Atoi(IdString)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(err.Error())
			return
		}
		var payload map[string]any
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(err.Error())
			return
		}
		num, err := UpdateRecordInTable(helper, tableName, Id, payload)
		if err != nil {
			if status := num["error"]; status == "Internal database error" {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(num)
			} else if status == "unknown table" {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(num)
			} else {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(num)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(num)
	})

	mux.HandleFunc("DELETE /{table}/{id}", func(w http.ResponseWriter, r *http.Request) {
		tableName := r.PathValue("table")
		IdString := r.PathValue("id")
		Id, err := strconv.Atoi(IdString)
		if err != nil {
			http.Error(w, "Ошибка", http.StatusInternalServerError)
			return
		}
		num, err := DeleteRecordFromTable(helper, tableName, Id)
		if err != nil {
			http.Error(w, "Ошибка", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(num)
	})

	return mux, nil
}
