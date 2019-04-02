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
	Select SelectQuery `json:"select,omitempty"`
	Insert CRUDStmt `json:"insert,omitempty"`
	Delete CRUDStmt `json:"delete,omitempty"`
	Update CRUDStmt `json:"update,omitempty"`
	Where  []WhereCond `json:"where,omitempty"`
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

type Fields struct {
	Table string `json:"table"`
	Fields []string `json:"fields, omitempty"`
}

// CRUDStmt a struct for all CRUD ops
type CRUDStmt struct {
	Table  string `json:"table"`
	// TODO rename to fields?
	Values []Values `json:"values,omitempty"`
}

// Values general purpose struct for linking values with fields
type Values struct {
	Field string `json:"field,omitempty"`
	Value string `json:"value,omitempty"`
}

// SelectQuery to build a SQL statement with ordering, grouping and support for joins
type SelectQuery struct {
	Query   []Fields `json:"query"`
	OrderBy []Order `json:"orderBy,omitempty"`
	GroupBy []string `json:"groupBy,omitempty"`
	Join    JoinStmt `json:"join,omitempty"`
}

// Order results by ascending or descending
type Order struct {
	Field string `json:"field"`
	ASC   bool `json:"asc"`
}

// JoinStmt to set the type of join and the conditions
type JoinStmt struct {
	Type string `json:"type"`
	Cond JoinCondition `json:"cond"`
}

// JoinCondition to set the source and destination table
type JoinCondition struct {
	From Condition `json:"from"`
	To   Condition `json:"to"`
}

// Condition to specify the table and field
type Condition struct {
	Table string `json:"table"`
	Field string `json:"field"`
}

// WhereCond to set the field, operation and whether to AND/OR with other ops
type WhereCond struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Val   string `json:"val"`
	Join  string `json:"join, omitempty"`
	Type  string `json:"type,omitempty"`
}

// ParseObject receives a JTOS struct and builds a SQL statement
func ParseObject(j JTOS) string {
	var stmt strings.Builder
	fmt.Println(reflect.DeepEqual(j.Insert, CRUDStmt{}))
	switch {
	case !reflect.DeepEqual(j.Insert, CRUDStmt{}):
		stmt.WriteString(buildInsert(j.Insert))
		break
	case !reflect.DeepEqual(j.Select, SelectQuery{}):
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

	for idx := range i.Values {
		val := i.Values[idx]
		fieldBuf.WriteString(val.Field)
		valBuf.WriteString("'")
		// TODO handle non string values
		valBuf.WriteString(val.Value)
		valBuf.WriteString("'")
		if idx != len(i.Values)-1 {
			fieldBuf.WriteString(", ")
			valBuf.WriteString(", ")
		}
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s);", i.Table, fieldBuf.String(), valBuf.String())
}

func checkComma(length int, idx int, buf *strings.Builder) strings.Builder {
	if idx != length-1 {
		buf.WriteString(", ")
	}
	return *buf
}

func buildSelect(s SelectQuery) string {
	var selectBuf, fieldBuf, tblBuf strings.Builder
	//selectBuf.WriteString("SELECT ")

	for idx := range s.Query {
		field := s.Query[idx]
		for fidx := range field.Fields {
			fieldBuf.WriteString(fmt.Sprintf("%s.%s", field.Table, field.Fields[fidx]))
			fieldBuf = checkComma(len(field.Fields), fidx, &fieldBuf)
		}
		if len(field.Fields) == 0 {
			fieldBuf.WriteString("*")
		}
		tblBuf.WriteString(field.Table)
		tblBuf = checkComma(len(s.Query), idx, &tblBuf)
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
	var whereBuf strings.Builder

	for idx := range w {
		cond := w[idx]
		whereBuf.WriteString(fmt.Sprintf("%s %s ", cond.Field, mappings[cond.Op]))
		if cond.Type == "number" || cond.Type == "raw" {
			whereBuf.WriteString(cond.Val)
		} else {
			whereBuf.WriteString(fmt.Sprintf("'%s' ", cond.Val))
		}
		whereBuf.WriteString(cond.Join)

	}

	return " WHERE " + whereBuf.String()

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
		ordBuf = checkComma(len(o), idx, &ordBuf)
	}
	return ordBuf.String()
}

func buildDelete(d CRUDStmt) string {
	// TODO handle conditions
	return fmt.Sprintf("DELETE FROM %s", d.Table)
}
