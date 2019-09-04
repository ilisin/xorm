// Copyright 2018 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	sql2 "database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const placeholderConverterSQL = "SELECT a, b FROM table_a WHERE b_id=(SELECT id FROM table_b WHERE b=?) AND id=? AND c=? AND d=? AND e=? AND f=?"
const placeholderConvertedSQL = "SELECT a, b FROM table_a WHERE b_id=(SELECT id FROM table_b WHERE b=$1) AND id=$2 AND c=$3 AND d=$4 AND e=$5 AND f=$6"
const placeholderBoundSQL = "SELECT a, b FROM table_a WHERE b_id=(SELECT id FROM table_b WHERE b=1) AND id=2.1 AND c='3' AND d=4 AND e='5' AND f=true"

func TestPlaceholderConverter(t *testing.T) {
	newSQL, err := ConvertPlaceholder(placeholderConverterSQL, "$")
	assert.NoError(t, err)
	assert.EqualValues(t, placeholderConvertedSQL, newSQL)
}

func BenchmarkPlaceholderConverter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ConvertPlaceholder(placeholderConverterSQL, "$")
	}
}

func TestBoundSQLConverter(t *testing.T) {
	newSQL, err := ConvertToBoundSQL(placeholderConverterSQL, []interface{}{1, 2.1, "3", uint(4), "5", true})
	assert.NoError(t, err)
	assert.EqualValues(t, placeholderBoundSQL, newSQL)

	newSQL, err = ConvertToBoundSQL(placeholderConverterSQL, []interface{}{1, 2.1, sql2.Named("any", "3"), uint(4), "5", true})
	assert.NoError(t, err)
	assert.EqualValues(t, placeholderBoundSQL, newSQL)

	newSQL, err = ConvertToBoundSQL(placeholderConverterSQL, []interface{}{1, 2.1, "3", 4, "5"})
	assert.Error(t, err)
	assert.EqualValues(t, ErrNeedMoreArguments, err)

	newSQL, err = ToBoundSQL(Select("id").From("table").Where(In("a", 1, 2)))
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT id FROM table WHERE a IN (1,2)", newSQL)

	newSQL, err = ToBoundSQL(Eq{"a": 1})
	assert.NoError(t, err)
	assert.EqualValues(t, "a=1", newSQL)

	newSQL, err = ToBoundSQL(1)
	assert.Error(t, err)
	assert.EqualValues(t, ErrNotSupportType, err)
}

func TestSQL(t *testing.T) {
	newSQL, args, err := ToSQL(In("a", 1, 2))
	assert.NoError(t, err)
	assert.EqualValues(t, "a IN (?,?)", newSQL)
	assert.EqualValues(t, []interface{}{1, 2}, args)

	newSQL, args, err = ToSQL(Select("id").From("table").Where(In("a", 1, 2)))
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT id FROM table WHERE a IN (?,?)", newSQL)
	assert.EqualValues(t, []interface{}{1, 2}, args)

	newSQL, args, err = ToSQL(1)
	assert.Error(t, err)
	assert.EqualValues(t, ErrNotSupportType, err)
}

type fiddler struct {
	sessionCode string
	dbType      int
}

func readPreparationSQLFromFile(path string) (string, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func TestReadPreparationSQLFromFile(t *testing.T) {
	sqlFromFile, err := readPreparationSQLFromFile("testdata/mysql_fiddle_data.sql")
	assert.NoError(t, err)
	fmt.Println(sqlFromFile)
}

/*
func TestNewFiddler(t *testing.T) {
	sqlFromFile, err := readPreparationSQLFromFile("testdata/mysql_fiddle_data.sql")
	assert.NoError(t, err)
	f, err := newFiddler("", MYSQL, sqlFromFile)
	assert.NoError(t, err)
	assert.NotEmpty(t, f.sessionCode)
}

func TestExecutableCheck(t *testing.T) {
	sqlFromFile, err := readPreparationSQLFromFile("testdata/mysql_fiddle_data.sql")
	assert.NoError(t, err)
	f, err := newFiddler("", MYSQL, sqlFromFile)
	assert.NoError(t, err)
	assert.NotEmpty(t, f.sessionCode)

	assert.NoError(t, f.executableCheck("SELECT * FROM table1"))

	err = f.executableCheck("SELECT * FROM table3")
	assert.Error(t, err)
}*/

func TestToSQLInDifferentDialects(t *testing.T) {
	sql, args, err := Postgres().Select().From("table1").Where(Eq{"a": "1"}.And(Neq{"b": "100"})).ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT * FROM table1 WHERE a=$1 AND b<>$2", sql)
	assert.EqualValues(t, []interface{}{"1", "100"}, args)

	sql, args, err = MySQL().Select().From("table1").Where(Eq{"a": "1"}.And(Neq{"b": "100"})).ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT * FROM table1 WHERE a=? AND b<>?", sql)
	assert.EqualValues(t, []interface{}{"1", "100"}, args)

	sql, args, err = MsSQL().Select().From("table1").Where(Eq{"a": "1"}.And(Neq{"b": "100"})).ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT * FROM table1 WHERE a=@p1 AND b<>@p2", sql)
	assert.EqualValues(t, []interface{}{sql2.Named("p1", "1"), sql2.Named("p2", "100")}, args)

	// test sql.NamedArg in cond
	sql, args, err = MsSQL().Select().From("table1").Where(Eq{"a": sql2.NamedArg{Name: "param", Value: "1"}}.And(Neq{"b": "100"})).ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT * FROM table1 WHERE a=@p1 AND b<>@p2", sql)
	assert.EqualValues(t, []interface{}{sql2.Named("p1", "1"), sql2.Named("p2", "100")}, args)

	sql, args, err = Oracle().Select().From("table1").Where(Eq{"a": "1"}.And(Neq{"b": "100"})).ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT * FROM table1 WHERE a=:p1 AND b<>:p2", sql)
	assert.EqualValues(t, []interface{}{sql2.Named("p1", "1"), sql2.Named("p2", "100")}, args)

	// test sql.NamedArg in cond
	sql, args, err = Oracle().Select().From("table1").Where(Eq{"a": sql2.Named("a", "1")}.And(Neq{"b": "100"})).ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT * FROM table1 WHERE a=:p1 AND b<>:p2", sql)
	assert.EqualValues(t, []interface{}{sql2.Named("p1", "1"), sql2.Named("p2", "100")}, args)

	sql, args, err = SQLite().Select().From("table1").Where(Eq{"a": "1"}.And(Neq{"b": "100"})).ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT * FROM table1 WHERE a=? AND b<>?", sql)
	assert.EqualValues(t, []interface{}{"1", "100"}, args)
}

func TestToSQLInjectionHarmlessDisposal(t *testing.T) {
	sql, err := MySQL().Select("*").From("table1").Where(Cond(Eq{"name": "cat';truncate table table1;"})).ToBoundSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT * FROM table1 WHERE name='cat'';truncate table table1;'", sql)
}
