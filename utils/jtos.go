package utils

import (
	"fmt"
	"reflect"
	"strings"
)

var mappings = map[string]string{
	"gt":  ">",
	"lt":  "<",
	"gte": ">=",
	"lte": "<=",
	"e":   "=",
	"ne":  "!=",
	"l":   "LIKE",
	"nl":  "NOT LIKE",
	"a":   "AND",
	"o":   "OR",
	"i":   "IN",
	"ni":  "NOT IN",
}

// JTOS wrapping struct to handle CRUD ops
type JTOS struct {
	Select SelectQuery
	Insert CRUDStmt
	Delete CRUDStmt
	Update CRUDStmt
	Where  []WhereCond
	Limit  int
	Offset int
}

// CRUDStmt a struct for all CRUD ops
type CRUDStmt struct {
	table  string
	values []Values
}

// Values general purpose struct for linking values with fields
type Values struct {
	field string
	value string
}

// SelectQuery to build a SQL statement with ordering, grouping and support for joins
type SelectQuery struct {
	Query   []CRUDStmt
	OrderBy []Order
	GroupBy []string
	Join    JoinStmt
}

// Order results by ascending or descending
type Order struct {
	Field string
	ASC   bool
}

// JoinStmt to set the type of join and the conditions
type JoinStmt struct {
	Type string
	Cond JoinCondition
}

// JoinCondition to set the source and destination table
type JoinCondition struct {
	from Condition
	to   Condition
}

// Condition to specify the table and field
type Condition struct {
	table string
	field string
}

// WhereCond to set the field, operation and whether to AND/OR with other ops
type WhereCond struct {
	Field string
	Op    string
	Val   string
	Join  string
}

// ParseObject receives a JTOS struct and builds a SQL statement
func ParseObject(j JTOS) string {
	var stmt strings.Builder

	switch {
	case !reflect.DeepEqual(j.Insert, CRUDStmt{}):
		stmt.WriteString(buildInsert(j.Insert))
		break
	case reflect.DeepEqual(j.Select, SelectQuery{}):
		stmt.WriteString(buildSelect(j.Select))
		break
	case !reflect.DeepEqual(j.Update, CRUDStmt{}):
		stmt.WriteString(buildUpdate(j.Update))
		break
	case !reflect.DeepEqual(j.Delete, CRUDStmt{}):
		stmt.WriteString(buildDelete(j.Update))
		break
	}

	if len(j.Where) > 0 {
		stmt.WriteString(buildWhere(j.Where))
	}

	if j.Limit > 0 {
		stmt.WriteString(fmt.Sprintf(" LIMIT %d", j.Limit))
	}

	if j.Offset > 0 {
		stmt.WriteString(fmt.Sprintf(" OFFSET %d", j.Offset))
	}
	stmt.WriteString(";")

	return stmt.String()
}

func buildInsert(i CRUDStmt) string {
	var valBuf, fieldBuf strings.Builder

	for idx := range i.values {
		val := i.values[idx]
		fieldBuf.WriteString(val.field)
		valBuf.WriteString("'")
		// TODO handle non string values
		valBuf.WriteString(val.value)
		valBuf.WriteString("'")
		if idx != len(i.values)-1 {
			fieldBuf.WriteString(", ")
			valBuf.WriteString(", ")
		}
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s);", i.table, fieldBuf.String(), valBuf.String())
}

func checkComma(length int, idx int, buf strings.Builder) strings.Builder {
	if idx != length-1 {
		buf.WriteString(", ")
	}
	return buf
}

func buildSelect(s SelectQuery) string {
	var selectBuf, fieldBuf, tblBuf strings.Builder
	selectBuf.WriteString("SELECT ")

	for idx := range s.Query {
		field := s.Query[idx]
		for fidx := range field.values {
			fieldBuf.WriteString(fmt.Sprintf("%s.%s", field.table, field.values[fidx]))
			fieldBuf = checkComma(len(field.values), fidx, fieldBuf)
		}
		tblBuf.WriteString(field.table)
		tblBuf = checkComma(len(s.Query), idx, tblBuf)
	}
	selectBuf.WriteString(fmt.Sprintf("SELECT %s FROM %s", fieldBuf.String(), tblBuf.String()))
	if !reflect.DeepEqual(s.Join, JoinStmt{}) {
		selectBuf.WriteString(buildJoin(s.Join))
	}
	if s.OrderBy != nil {
		selectBuf.WriteString(buildOrder(s.OrderBy))
	}
	if s.GroupBy != nil {
		selectBuf.WriteString(buildOrder(s.OrderBy))
	}

	return selectBuf.String()
}

func buildUpdate(u CRUDStmt) string {
	return ""
}

func buildWhere(w []WhereCond) string {
	// TODO handle conditions
	return ""
}

func buildJoin(j JoinStmt) string {
	return ""
}

func buildOrder(o []Order) string {
	var ordBuf strings.Builder
	ordBuf.WriteString(" ORDER BY ")
	for idx := range o {
		ordBuf.WriteString(o[idx].Field)
		switch o[idx].ASC {
		case true:
			ordBuf.WriteString(" ASC")
		default:
			ordBuf.WriteString(" DESC")
		}
		ordBuf = checkComma(len(o), idx, ordBuf)
	}
	return ordBuf.String()
}

func buildDelete(d CRUDStmt) string {
	// TODO handle conditions
	return fmt.Sprintf("DELETE FROM %s", d.table)
}
