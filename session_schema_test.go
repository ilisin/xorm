// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStoreEngine(t *testing.T) {
	assert.NoError(t, prepareEngine())

	assert.NoError(t, testEngine.DropTables("user_store_engine"))

	type UserinfoStoreEngine struct {
		Id   int64
		Name string
	}

	assert.NoError(t, testEngine.StoreEngine("InnoDB").Table("user_store_engine").CreateTable(&UserinfoStoreEngine{}))
}

func TestCreateTable(t *testing.T) {
	assert.NoError(t, prepareEngine())

	assert.NoError(t, testEngine.DropTables("user_user"))

	type UserinfoCreateTable struct {
		Id   int64
		Name string
	}

	assert.NoError(t, testEngine.Table("user_user").CreateTable(&UserinfoCreateTable{}))
}

func TestCreateMultiTables(t *testing.T) {
	assert.NoError(t, prepareEngine())

	session := testEngine.NewSession()
	defer session.Close()

	type UserinfoMultiTable struct {
		Id   int64
		Name string
	}

	user := &UserinfoMultiTable{}
	assert.NoError(t, session.Begin())

	for i := 0; i < 10; i++ {
		tableName := fmt.Sprintf("user_%v", i)

		assert.NoError(t, session.DropTable(tableName))

		assert.NoError(t, session.Table(tableName).CreateTable(user))
	}

	assert.NoError(t, session.Commit())
}

func TestIsTableEmpty(t *testing.T) {
	assert.NoError(t, prepareEngine())

	type NumericEmpty struct {
		Numeric float64 `xorm:"numeric(26,2)"`
	}

	type PictureEmpty struct {
		Id          int64
		Url         string `xorm:"unique"` //image's url
		Title       string
		Description string
		Created     time.Time `xorm:"created"`
		ILike       int
		PageView    int
		From_url    string
		Pre_url     string `xorm:"unique"` //pre view image's url
		Uid         int64
	}

	assert.NoError(t, testEngine.DropTables(&PictureEmpty{}, &NumericEmpty{}))

	assert.NoError(t, testEngine.Sync(new(PictureEmpty), new(NumericEmpty)))

	isEmpty, err := testEngine.IsTableEmpty(&PictureEmpty{})
	assert.NoError(t, err)
	assert.True(t, isEmpty)

	tbName := testEngine.TableMapper.Obj2Table("PictureEmpty")
	isEmpty, err = testEngine.IsTableEmpty(tbName)
	assert.NoError(t, err)
	assert.True(t, isEmpty)
}

type CustomTableName struct {
	Id   int64
	Name string
}

func (c *CustomTableName) TableName() string {
	return "customtablename"
}

func TestCustomTableName(t *testing.T) {
	assert.NoError(t, prepareEngine())

	c := new(CustomTableName)
	assert.NoError(t, testEngine.DropTables(c))

	assert.NoError(t, testEngine.CreateTables(c))
}

func TestDump(t *testing.T) {
	assert.NoError(t, prepareEngine())

	fp := testEngine.Dialect().URI().DbName + ".sql"
	os.Remove(fp)
	assert.NoError(t, testEngine.DumpAllToFile(fp))
}

type IndexOrUnique struct {
	Id        int64
	Index     int `xorm:"index"`
	Unique    int `xorm:"unique"`
	Group1    int `xorm:"index(ttt)"`
	Group2    int `xorm:"index(ttt)"`
	UniGroup1 int `xorm:"unique(lll)"`
	UniGroup2 int `xorm:"unique(lll)"`
}

func TestIndexAndUnique(t *testing.T) {
	assert.NoError(t, prepareEngine())

	assert.NoError(t, testEngine.CreateTables(&IndexOrUnique{}))

	assert.NoError(t, testEngine.DropTables(&IndexOrUnique{}))

	assert.NoError(t, testEngine.CreateTables(&IndexOrUnique{}))

	assert.NoError(t, testEngine.CreateIndexes(&IndexOrUnique{}))

	assert.NoError(t, testEngine.CreateUniques(&IndexOrUnique{}))
}

func TestMetaInfo(t *testing.T) {
	assert.NoError(t, prepareEngine())
	assert.NoError(t, testEngine.Sync2(new(CustomTableName), new(IndexOrUnique)))

	tables, err := testEngine.DBMetas()
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(tables))
	assert.EqualValues(t, "customtablename", tables[0].Name)
	assert.EqualValues(t, "index_or_unique", tables[1].Name)
}
