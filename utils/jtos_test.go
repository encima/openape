package utils

import (
	"fmt"
	"testing"
)

func TestJTOSSelectAll(t *testing.T) {
	sel := CRUDStmt{table:"users", values: nil}
	var sels []CRUDStmt
	sels = append(sels, sel)
	q := SelectQuery{sels, nil, nil, JoinStmt{}}

	j := JTOS{q, CRUDStmt{}, CRUDStmt{}, CRUDStmt{}, nil, 0, 0}

	s := ParseObject(j)
	fmt.Println(s)

}

func TestJTOSSelectFields(t *testing.T) {
	var vals []Values
	v := Values{"username", ""}
	vals = append(vals, v)
	sel := CRUDStmt{table:"users", values: vals}
	var sels []CRUDStmt
	sels = append(sels, sel)
	q := SelectQuery{sels, nil, nil, JoinStmt{}}

	j := JTOS{q, CRUDStmt{}, CRUDStmt{}, CRUDStmt{}, nil, 0, 0}

	s := ParseObject(j)
	fmt.Println(s)

}
