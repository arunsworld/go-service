package query

import "database/sql"

// ColumnHandler is a function that handles a callback for the columns (Names & Types)
type ColumnHandler func([]*sql.ColumnType)

// RowHandler is a function that handles a callback for a row in the table
type RowHandler func([]string)

// GenericQueryHandlers holds the handlers required by the generic query during callback
type GenericQueryHandlers struct {
	ColHandler ColumnHandler
	RowHandler RowHandler
}

// GenericQuery performs the given query on the given DB; and calls the callbacks
func GenericQuery(db *sql.DB, query string, handlers GenericQueryHandlers) error {
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.ColumnTypes()
	if err != nil {
		return err
	}
	if handlers.ColHandler != nil {
		handlers.ColHandler(cols)
	}

	if handlers.RowHandler == nil {
		return nil
	}

	vals := make([]interface{}, len(cols))
	for i := range cols {
		vals[i] = new(sql.RawBytes)
	}
	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			return err
		}
		row := make([]string, len(cols))
		for i, v := range vals {
			row[i] = string(*(v.(*sql.RawBytes)))
		}
		handlers.RowHandler(row)
	}
	return nil
}
