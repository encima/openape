package utils

import (
	"encoding/json"
	"testing"
)

const expectedUName = "SELECT users.username FROM users;"
const expectedUNameEmail = "SELECT users.username, users.email FROM users;"

func TestJTOSSelectAll(t *testing.T) {
	expected := "SELECT * FROM users;"
	sel := Fields{Table: "users", Fields: nil}
	var sels []Fields
	sels = append(sels, sel)
	q := SelectQuery{sels, nil, nil, JoinStmt{}}

	j := JTOS{q, CRUDStmt{}, CRUDStmt{}, CRUDStmt{}, nil, 0, 0}

	s := ParseObject(j)
	if s != expected {
		t.Errorf("%s does not match %s", s, expected)
	}

}

func TestJTOSSelectFields(t *testing.T) {
	var vals []string
	v := "username"
	vals = append(vals, v)
	sel := Fields{Table: "users", Fields: vals}
	var sels []Fields
	sels = append(sels, sel)
	q := SelectQuery{sels, nil, nil, JoinStmt{}}

	j := JTOS{q, CRUDStmt{}, CRUDStmt{}, CRUDStmt{}, nil, 0, 0}

	s := ParseObject(j)
	if s != expectedUName {
		t.Errorf("%s does not match %s", s, expectedUName)
	}

}

func TestJTOSParse(t *testing.T) {
	getJson := `{"select": { "query": [ {"table": "users", "fields": ["username", "email"] }]}}`
	var jQuery JTOS
	err := json.Unmarshal([]byte(getJson), &jQuery)
	if err != nil {
		panic(err)
	}

	s := ParseObject(jQuery)
	if s != expectedUNameEmail {
		t.Errorf("%s does not match %s", s, expectedUNameEmail)
	}
}
