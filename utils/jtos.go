package utils

import (
	"fmt"
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

type JTOS struct {
	Select SelectQuery
	Insert CRUDStmt
	Delete CRUDStmt
	Update CRUDStmt
	Limit  int
	Offset int
}

type CRUDStmt struct {
	table  string
	values []Values
}

type Values struct {
	field string
	value string
}

type SelectQuery struct {
	Query   []CRUDStmt
	OrderBy []Order
	GroupBy []string
}

type Order struct {
	Field string
	ASC   bool
}

type JoinStmt struct {
	Type string
}

type JoinConditions struct {
	from Condition
	to   Condition
}

type Condition struct {
	table string
	field string
}

type WhereCond struct {
	Field string
	Op    string
	Val   string
	Join  string
}

func ParseObject() {

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
	if s.OrderBy != nil {
		selectBuf.WriteString(buildOrder(s.OrderBy))
	}
	if s.GroupBy != nil {
		selectBuf.WriteString(buildOrder(s.OrderBy))
	}

	return selectBuf.String()
}

func buildWhere(w WhereCond) string {
	return ""
}

func buildJoin(j JoinConditions) string {
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
