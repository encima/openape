package utils

import (
	"testing"
)

func TestJTOSSelectAll(t *testing.T) {
	expected := "SELECT * FROM users;"
	sel := CRUDStmt{table:"users", values: nil}
	var sels []CRUDStmt
	sels = append(sels, sel)
	q := SelectQuery{sels, nil, nil, JoinStmt{}}

	j := JTOS{q, CRUDStmt{}, CRUDStmt{}, CRUDStmt{}, nil, 0, 0}

	s := ParseObject(j)
	if s != expected {
		t.Errorf("%s does not match %s", s, expected)
	}


}

func TestJTOSSelectFields(t *testing.T) {
	expected := "SELECT users.username FROM users;"
	var vals []Values
	v := Values{"username", ""}
	vals = append(vals, v)
	sel := CRUDStmt{table:"users", values: vals}
	var sels []CRUDStmt
	sels = append(sels, sel)
	q := SelectQuery{sels, nil, nil, JoinStmt{}}

	j := JTOS{q, CRUDStmt{}, CRUDStmt{}, CRUDStmt{}, nil, 0, 0}

	s := ParseObject(j)
	if s != expected {
		t.Errorf("%s does not match %s", s, expected)
	}

}

func TestJTOSParse(t *testing.T) {

}
