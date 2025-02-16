/*
 * Licensed to the AcmeStack under one or more contributor license
 * agreements. See the NOTICE file distributed with this work for
 * additional information regarding copyright ownership.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tests

import (
	"errors"
	"fmt"
	"github.com/acmestack/gorm-plus/gplus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"testing"
)

var gormDb *gorm.DB
var gormDbConnName = "test1"
var dbAddress = "127.0.0.1:3306"
var dbUser = "root"
var dbPassword = "123456"

func init() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/test?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbAddress)
	var err error
	gormDb, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		fmt.Println(err)
	}
	var u User
	gormDb.AutoMigrate(u)
	gplus.Init(gormDb)
	initDb()
}

func initDb() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/test1?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbAddress)
	var err error
	gormDb1, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		fmt.Println(err)
	}
	var u User
	gormDb1.AutoMigrate(u)
	gplus.InitDb(gormDb1, gormDbConnName)
}

func TestInsert(t *testing.T) {
	deleteOldData()

	user := &User{Username: "afumu", Password: "123456", Age: 18, Score: 100, Dept: "开发部门"}
	resultDb := gplus.Insert(user)

	if resultDb.Error != nil {
		t.Fatalf("errors happened when insert: %v", resultDb.Error)
	} else if resultDb.RowsAffected != 1 {
		t.Fatalf("rows affected expects: %v, got %v", 1, resultDb.RowsAffected)
	}

	newUser, db := gplus.SelectById[User](user.ID)
	if db.Error != nil {
		t.Fatalf("errors happened when SelectById: %v", db.Error)
	}
	AssertObjEqual(t, newUser, user, "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")
}

func TestInsertBatch(t *testing.T) {
	deleteOldData()
	users := getUsers()
	resultDb := gplus.InsertBatch[User](users)
	if resultDb.RowsAffected != int64(len(users)) {
		t.Errorf("affected rows should be %v, but got %v", len(users), resultDb.RowsAffected)
	}

	for _, user := range users {
		newUser, db := gplus.SelectById[User](user.ID)
		if db.Error != nil {
			t.Fatalf("errors happened when SelectById: %v", db.Error)
		}
		AssertObjEqual(t, newUser, user, "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")
	}
}

func TestInsertBatchSize(t *testing.T) {
	deleteOldData()
	users := getUsers()
	resultDb := gplus.InsertBatchSize[User](users, 2)
	if resultDb.RowsAffected != int64(len(users)) {
		t.Errorf("affected rows should be %v, but got %v", len(users), resultDb.RowsAffected)
	}

	for _, user := range users {
		newUser, db := gplus.SelectById[User](user.ID)
		if db.Error != nil {
			t.Fatalf("errors happened when SelectById: %v", db.Error)
		}
		AssertObjEqual(t, newUser, user, "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")
	}
}

func TestDeleteById(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatchSize[User](users, 2)

	if res := gplus.DeleteById[User](users[1].ID); res.Error != nil || res.RowsAffected != 1 {
		t.Errorf("errors happened when deleteById: %v, affected: %v", res.Error, res.RowsAffected)
	}

	_, resultDb := gplus.SelectById[User](users[1].ID)
	if !errors.Is(resultDb.Error, gorm.ErrRecordNotFound) {
		t.Errorf("should returns record not found error, but got %v", resultDb.Error)
	}
}

func TestDeleteByIds(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)

	userIds := []int64{users[0].ID, users[1].ID, users[2].ID}
	if res := gplus.DeleteByIds[User](userIds); res.Error != nil || res.RowsAffected != 3 {
		t.Errorf("errors happened when deleteByIds: %v, affected: %v", res.Error, res.RowsAffected)
	}

	_, resultDb := gplus.SelectById[User](users[0].ID)
	if !errors.Is(resultDb.Error, gorm.ErrRecordNotFound) {
		t.Errorf("should returns record not found error, but got %v", resultDb.Error)
	}

	_, resultDb = gplus.SelectById[User](users[1].ID)
	if !errors.Is(resultDb.Error, gorm.ErrRecordNotFound) {
		t.Errorf("should returns record not found error, but got %v", resultDb.Error)
	}

	_, resultDb = gplus.SelectById[User](users[2].ID)
	if !errors.Is(resultDb.Error, gorm.ErrRecordNotFound) {
		t.Errorf("should returns record not found error, but got %v", resultDb.Error)
	}
}

func TestDelete(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)

	query, u := gplus.NewQuery[User]()
	query.Eq(&u.Username, "afumu1")
	if res := gplus.Delete[User](query); res.Error != nil || res.RowsAffected != 1 {
		t.Errorf("errors happened when Delete: %v, affected: %v", res.Error, res.RowsAffected)
	}

	_, resultDb := gplus.SelectOne[User](query)
	if !errors.Is(resultDb.Error, gorm.ErrRecordNotFound) {
		t.Errorf("should returns record not found error, but got %v", resultDb.Error)
	}
}

func TestUpdateById(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)

	user := users[0]
	user.Score = 100
	user.Age = 25

	if res := gplus.UpdateById[User](user); res.Error != nil || res.RowsAffected != 1 {
		t.Errorf("errors happened when deleteByIds: %v, affected: %v", res.Error, res.RowsAffected)
	}

	newUser, db := gplus.SelectById[User](user.ID)
	if db.Error != nil {
		t.Fatalf("errors happened when SelectById: %v", db.Error)
	}
	AssertObjEqual(t, newUser, user, "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")

}

func TestUpdateZeroById(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)

	updateUser := &User{ID: users[0].ID, Score: 100, Age: 25, CreatedAt: users[0].CreatedAt}

	if res := gplus.UpdateZeroById[User](updateUser); res.Error != nil || res.RowsAffected != 1 {
		t.Errorf("errors happened when deleteByIds: %v, affected: %v", res.Error, res.RowsAffected)
	}

	newUser, db := gplus.SelectById[User](updateUser.ID)
	if db.Error != nil {
		t.Fatalf("errors happened when SelectById: %v", db.Error)
	}
	AssertObjEqual(t, newUser, updateUser, "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")

}

func TestUpdate(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)

	user := users[0]
	q, model := gplus.NewQuery[User]()
	q.Eq(&model.ID, user.ID).Set(&model.Score, 60)
	if err := gplus.Update(q).Error; err != nil {
		t.Errorf("errors happened when update: %v", err)
	}
	newUser, db := gplus.SelectById[User](user.ID)
	if db.Error != nil {
		t.Fatalf("errors happened when SelectById: %v", db.Error)
	}
	user.Score = 60
	AssertObjEqual(t, newUser, user, "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")
}

func TestSelectById(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)
	user := users[0]
	resultUser, db := gplus.SelectById[User](user.ID)
	if db.Error != nil {
		t.Errorf("errors happened when selectById : %v", db.Error)
	} else {
		AssertObjEqual(t, resultUser, user, "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")
	}
}

func TestSelectByIds(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)

	userIds := []int64{users[0].ID, users[1].ID}
	resultUsers, db := gplus.SelectByIds[User](userIds)
	if db.Error != nil {
		t.Errorf("errors happened when selectByIds : %v", db.Error)
	} else {
		AssertObjEqual(t, resultUsers[0], users[0], "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")
		AssertObjEqual(t, resultUsers[1], users[1], "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")
	}
}

func TestSelectOne(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)

	query, model := gplus.NewQuery[User]()
	query.Eq(&model.Username, users[0].Username)
	resultUser, db := gplus.SelectOne[User](query)
	if db.Error != nil {
		t.Errorf("errors happened when selectByOne : %v", db.Error)
	} else {
		AssertObjEqual(t, resultUser, users[0], "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")
	}
}

func TestSelectList(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)

	query, model := gplus.NewQuery[User]()
	query.Eq(&model.Username, users[0].Username).Or().Eq(&model.Username, users[5].Username)
	resultUsers, db := gplus.SelectList(query)
	if db.Error != nil {
		t.Errorf("errors happened when SelectList : %v", db.Error)
	} else {
		AssertObjEqual(t, resultUsers[0], users[0], "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")
		AssertObjEqual(t, resultUsers[1], users[5], "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")
	}
}

func TestSelectPage(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)

	query, model := gplus.NewQuery[User]()
	page := gplus.NewPage[User](1, 10)
	query.Eq(&model.Username, users[0].Username).Or().Eq(&model.Username, users[5].Username)
	resultPage, db := gplus.SelectPage(page, query)
	if db.Error != nil {
		t.Errorf("errors happened when selectByIds : %v", db.Error)
	}
	if resultPage.Total != 2 {
		t.Errorf("page total expects: %v, got %v", 2, resultPage.Total)
	}

	AssertObjEqual(t, resultPage.Records[0], users[0], "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")
	AssertObjEqual(t, resultPage.Records[1], users[5], "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")

}

func TestSelectPageDefaultPageParam(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)

	query, model := gplus.NewQuery[User]()
	page := gplus.NewPage[User](0, 0)
	query.Eq(&model.Username, users[0].Username).Or().Eq(&model.Username, users[5].Username)
	resultPage, db := gplus.SelectPage(page, query)
	if db.Error != nil {
		t.Errorf("errors happened when selectByIds : %v", db.Error)
	}
	if resultPage.Total != 2 {
		t.Errorf("page total expects: %v, got %v", 2, resultPage.Total)
	}

	AssertObjEqual(t, page.Current, 1)
	AssertObjEqual(t, page.Size, 10)
}

func TestSelectPageGeneric2(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)

	type UserVo struct {
		ID       int64
		Username string
		Password string
	}

	query, model := gplus.NewQuery[User]()
	page := gplus.NewPage[UserVo](1, 10)
	query.Eq(&model.Username, users[0].Username).Or().Eq(&model.Username, users[5].Username)

	resultPage, db := gplus.SelectPageGeneric[User, UserVo](page, query)
	if db.Error != nil {
		t.Errorf("errors happened when selectByIds : %v", db.Error)
	}
	if resultPage.Total != 2 {
		t.Errorf("page total expects: %v, got %v", 2, resultPage.Total)
	}

	AssertObjEqual(t, resultPage.Records[0], users[0], "ID", "Username", "Password")
	AssertObjEqual(t, resultPage.Records[1], users[5], "ID", "Username", "Password")

}

func TestSelectPageGeneric3(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)

	query, model := gplus.NewQuery[User]()
	page := gplus.NewPage[map[string]any](1, 10)
	query.Eq(&model.Username, users[0].Username).Or().Eq(&model.Username, users[5].Username)

	resultPage, db := gplus.SelectPageGeneric[User, map[string]any](page, query)
	if db.Error != nil {
		t.Errorf("errors happened when selectByIds : %v", db.Error)
	}
	if resultPage.Total != 2 {
		t.Errorf("page total expects: %v, got %v", 2, resultPage.Total)
	}

	var userResult []*User
	for _, userMap := range resultPage.RecordsMap {
		user := &User{Username: userMap["username"].(string), Password: userMap["password"].(string)}
		userResult = append(userResult, user)
	}

	AssertObjEqual(t, userResult[0], users[0], "Username", "Password")
	AssertObjEqual(t, userResult[1], users[5], "Username", "Password")

}

func TestSelectCount(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)

	query, model := gplus.NewQuery[User]()
	query.Eq(&model.Username, users[0].Username).Or().Eq(&model.Username, users[5].Username)
	count, db := gplus.SelectCount(query)
	if db.Error != nil {
		t.Errorf("errors happened when SelectCount : %v", db.Error)
	}
	if count != 2 {
		t.Errorf("count expects: %v, got %v", 2, count)
	}
}

func TestExists(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)

	query, model := gplus.NewQuery[User]()
	query.Eq(&model.Username, users[0].Username)
	exists, db := gplus.Exists[User](query)
	if db.Error != nil {
		t.Errorf("errors happened when SelectCount : %v", db.Error)
	}
	if !exists {
		t.Errorf("errors happened when SelectCount : %v", db.Error)
	}
}

func TestSelectGeneric1(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)
	type UserVo struct {
		Username string
		Age      int
	}
	query, u := gplus.NewQuery[User]()
	query.Eq(&u.Username, users[0].Username)
	userVo, resultDb := gplus.SelectGeneric[User, UserVo](query)

	if resultDb.Error != nil {
		t.Errorf("errors happened when resultDb : %v", resultDb.Error)
	}

	if userVo.Username != users[0].Username || userVo.Age != users[0].Age {
		t.Errorf("errors happened when SelectGeneric")
	}
}

func TestSelectGeneric2(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)
	type UserVo struct {
		Name string
		Age  int
	}
	query, u := gplus.NewQuery[User]()
	uvo := gplus.GetModel[UserVo]()
	query.Eq(&u.Username, users[0].Username).Select(gplus.As(&u.Username, &uvo.Name), &u.Age)
	userVo, resultDb := gplus.SelectGeneric[User, UserVo](query)

	if resultDb.Error != nil {
		t.Errorf("errors happened when resultDb : %v", resultDb.Error)
	}

	if userVo.Name != users[0].Username || userVo.Age != users[0].Age {
		t.Errorf("errors happened when SelectGeneric")
	}
}

func TestSelectGeneric3(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)
	type UserVo struct {
		TotalAge int
	}
	var totalAge int
	for _, user := range users {
		totalAge += user.Age
	}
	query, u := gplus.NewQuery[User]()
	uvo := gplus.GetModel[UserVo]()
	model := gplus.GetModel[UserVo]()
	fmt.Println(model)
	query.Select(gplus.Sum(&u.Age).As(&uvo.TotalAge))
	userVo, resultDb := gplus.SelectGeneric[User, UserVo](query)

	if resultDb.Error != nil {
		t.Errorf("errors happened when resultDb : %v", resultDb.Error)
	}

	if userVo.TotalAge != totalAge {
		t.Errorf("errors happened when SelectGeneric")
	}
}

func TestSelectGeneric4(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)
	var totalAge int
	for _, user := range users {
		totalAge += user.Age
	}
	query, u := gplus.NewQuery[User]()
	query.Select(gplus.Sum(&u.Age))
	total, resultDb := gplus.SelectGeneric[User, int](query)

	if resultDb.Error != nil {
		t.Errorf("errors happened when resultDb : %v", resultDb.Error)
	}

	if total != totalAge {
		t.Errorf("errors happened when SelectGeneric")
	}
}

func TestSelectGeneric5(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)

	var ages []int
	var agesMap = make(map[int]struct{})
	for _, user := range users {
		agesMap[user.Age] = struct{}{}
	}
	for key, _ := range agesMap {
		ages = append(ages, key)
	}
	sort.Ints(ages)
	query, u := gplus.NewQuery[User]()
	query.Select(&u.Age).Distinct(&u.Age)
	allAges, _ := gplus.SelectGeneric[User, []int](query)
	sort.Ints(allAges)
	if !reflect.DeepEqual(allAges, ages) {
		t.Errorf("errors happened when SelectGeneric")
	}
}

func TestSelectGeneric6(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)
	type UserVo struct {
		Dept  string
		Score int
	}
	var userMap = make(map[string]int)
	for _, user := range users {
		userMap[user.Dept] += user.Score
	}
	query, u := gplus.NewQuery[User]()
	uvo := gplus.GetModel[UserVo]()
	query.Select(&u.Dept, gplus.Sum(&u.Score).As(&uvo.Score)).Group(&u.Dept)
	UserVos, resultDb := gplus.SelectGeneric[User, []UserVo](query)

	if resultDb.Error != nil {
		t.Errorf("errors happened when resultDb : %v", resultDb.Error)
	}

	for _, userVo := range UserVos {
		score := userMap[userVo.Dept]
		if userVo.Score != score {
			t.Errorf("errors happened when SelectGeneric")
		}
	}
}

func TestSelectGeneric7(t *testing.T) {
	deleteOldData()
	users := getUsers()
	gplus.InsertBatch[User](users)
	var userMap = make(map[string]int)
	for _, user := range users {
		userMap[user.Dept] += user.Score
	}
	query, u := gplus.NewQuery[User]()
	query.Select(&u.Dept, gplus.Sum(&u.Score).As("score")).Group(&u.Dept)
	UserVos, resultDb := gplus.SelectGeneric[User, []map[string]any](query)

	if resultDb.Error != nil {
		t.Errorf("errors happened when resultDb : %v", resultDb.Error)
	}

	for _, umap := range UserVos {
		scoreStr := umap["score"].(string)
		score, err := strconv.Atoi(scoreStr)
		if err != nil {
			t.Errorf("errors happened when SelectGeneric")
		}
		dept := umap["dept"].(string)

		if userMap[dept] != score {
			t.Errorf("errors happened when SelectGeneric")
		}
	}
}

func TestTx(t *testing.T) {
	deleteOldData()
	users := getUsers()
	err := gplus.Tx(func(tx *gorm.DB) error {
		err := gplus.InsertBatch[User](users, gplus.Db(tx)).Error
		return err
	})
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestInsertBaseDb(t *testing.T) {
	deleteOldDataBaseDb()

	user := &User{Username: "afumu", Password: "123456", Age: 18, Score: 100, Dept: "开发部门"}
	resultDb := gplus.Insert(user, gplus.DbConnName(gormDbConnName))

	if resultDb.Error != nil {
		t.Fatalf("errors happened when insert: %v", resultDb.Error)
	} else if resultDb.RowsAffected != 1 {
		t.Fatalf("rows affected expects: %v, got %v", 1, resultDb.RowsAffected)
	}

	newUser, db := gplus.SelectById[User](user.ID, gplus.DbConnName(gormDbConnName))
	if db.Error != nil {
		t.Fatalf("errors happened when SelectById: %v", db.Error)
	}
	AssertObjEqual(t, newUser, user, "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")
}

func TestInsertBatchBaseDb(t *testing.T) {
	deleteOldDataBaseDb()
	users := getUsers()
	resultDb := gplus.InsertBatch[User](users, gplus.DbConnName(gormDbConnName))
	if resultDb.RowsAffected != int64(len(users)) {
		t.Errorf("affected rows should be %v, but got %v", len(users), resultDb.RowsAffected)
	}

	for _, user := range users {
		newUser, db := gplus.SelectById[User](user.ID, gplus.DbConnName(gormDbConnName))
		if db.Error != nil {
			t.Fatalf("errors happened when SelectById: %v", db.Error)
		}
		AssertObjEqual(t, newUser, user, "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")
	}
}

func TestDeleteByIdBaseDb(t *testing.T) {
	deleteOldDataBaseDb()
	users := getUsers()
	gplus.InsertBatchSize[User](users, 2, gplus.DbConnName(gormDbConnName))

	if res := gplus.DeleteById[User](users[1].ID, gplus.DbConnName(gormDbConnName)); res.Error != nil || res.RowsAffected != 1 {
		t.Errorf("errors happened when deleteById: %v, affected: %v", res.Error, res.RowsAffected)
	}

	_, resultDb := gplus.SelectById[User](users[1].ID, gplus.DbConnName(gormDbConnName))
	if !errors.Is(resultDb.Error, gorm.ErrRecordNotFound) {
		t.Errorf("should returns record not found error, but got %v", resultDb.Error)
	}
}

func TestDeleteBaseDb(t *testing.T) {
	deleteOldDataBaseDb()
	users := getUsers()
	opt := gplus.DbConnName(gormDbConnName)
	gplus.InsertBatch[User](users, opt)

	query, u := gplus.NewQueryBaseDb[User](opt)
	query.Eq(&u.Username, "afumu1")
	if res := gplus.Delete[User](query, gplus.DbConnName(gormDbConnName)); res.Error != nil || res.RowsAffected != 1 {
		t.Errorf("errors happened when Delete: %v, affected: %v", res.Error, res.RowsAffected)
	}

	_, resultDb := gplus.SelectOne[User](query, gplus.DbConnName(gormDbConnName))
	if !errors.Is(resultDb.Error, gorm.ErrRecordNotFound) {
		t.Errorf("should returns record not found error, but got %v", resultDb.Error)
	}
}

func TestUpdateByIdBaseDb(t *testing.T) {
	deleteOldDataBaseDb()
	users := getUsers()
	gplus.InsertBatch[User](users, gplus.DbConnName(gormDbConnName))

	user := users[0]
	user.Score = 100
	user.Age = 25

	if res := gplus.UpdateById[User](user, gplus.DbConnName(gormDbConnName)); res.Error != nil || res.RowsAffected != 1 {
		t.Errorf("errors happened when deleteByIds: %v, affected: %v", res.Error, res.RowsAffected)
	}

	newUser, db := gplus.SelectById[User](user.ID, gplus.DbConnName(gormDbConnName))
	if db.Error != nil {
		t.Fatalf("errors happened when SelectById: %v", db.Error)
	}
	AssertObjEqual(t, newUser, user, "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")

}

func TestSelectByIdBaseDb(t *testing.T) {
	deleteOldDataBaseDb()
	users := getUsers()
	gplus.InsertBatch[User](users, gplus.DbConnName(gormDbConnName))
	user := users[0]
	resultUser, db := gplus.SelectById[User](user.ID, gplus.DbConnName(gormDbConnName))
	if db.Error != nil {
		t.Errorf("errors happened when selectById : %v", db.Error)
	} else {
		AssertObjEqual(t, resultUser, user, "ID", "Username", "Password", "Address", "Age", "Phone", "Score", "Dept", "CreatedAt", "UpdatedAt")
	}
}

func TestSelectGeneric6BaseDb(t *testing.T) {
	deleteOldDataBaseDb()
	users := getUsers()
	opt := gplus.DbConnName(gormDbConnName)
	gplus.InsertBatch[User](users, opt)
	type UserVo struct {
		Dept  string
		Score int
	}
	var userMap = make(map[string]int)
	for _, user := range users {
		userMap[user.Dept] += user.Score
	}
	//测试NewQueryBaseDb和GetModelBaseDb
	query, u := gplus.NewQueryBaseDb[User](opt)
	uvo := gplus.GetModelBaseDb[UserVo](opt)
	query.Select(&u.Dept, gplus.Sum(&u.Score).As(&uvo.Score)).Group(&u.Dept)
	UserVos, resultDb := gplus.SelectGeneric[User, []UserVo](query, opt)

	if resultDb.Error != nil {
		t.Errorf("errors happened when resultDb : %v", resultDb.Error)
	}

	for _, userVo := range UserVos {
		score := userMap[userVo.Dept]
		if userVo.Score != score {
			t.Errorf("errors happened when SelectGeneric")
		}
	}

	//测试NewQueryModelBaseDb
	type UserV1 struct {
		Name string
		Age  int64
	}
	query, user, userV1 := gplus.NewQueryModelBaseDb[User, UserV1](opt)
	query.Eq(&user.Username, "afumu").And(func(q *gplus.QueryCond[User]) {
		q.Eq(&user.Address, "北京").Or().Eq(&user.Age, 20)
	}).Select(gplus.As(&user.Username, &userV1.Name), &user.Age)
	gplus.SelectGeneric[User, []UserV1](query, opt)

	//如果还是使用旧有的方法测试
	query, u = gplus.NewQuery[User]()
	uvo = gplus.GetModel[UserVo]()
	query.Select(&u.Dept, gplus.Sum(&u.Score).As(&uvo.Score)).Group(&u.Dept)
	UserVos, resultDb = gplus.SelectGeneric[User, []UserVo](query, opt)

	if resultDb.Error != nil {
		t.Errorf("errors happened when resultDb : %v", resultDb.Error)
	}
}

func TestQueryByIdBaseDb(t *testing.T) {
	opt := gplus.DbConnName(gormDbConnName)
	values := url.Values{}
	values["q"] = []string{"id=1"}
	query := gplus.BuildQueryBaseDb[User](values, opt)
	gplus.SelectList[User](query, opt)
}

func deleteOldData() {
	q, u := gplus.NewQuery[User]()
	q.IsNotNull(&u.ID)
	gplus.Delete(q)
}

func deleteOldDataBaseDb() {
	opt := gplus.DbConnName(gormDbConnName)
	q, u := gplus.NewQueryBaseDb[User](opt)
	q.IsNotNull(&u.ID)
	gplus.Delete(q, gplus.DbConnName(gormDbConnName))
}

func getUsers() []*User {
	user1 := &User{Username: "afumu1", Password: "123456", Age: 18, Score: 12, Dept: "开发部门"}
	user2 := &User{Username: "afumu2", Password: "123456", Age: 16, Score: 34, Dept: "行政部门"}
	user3 := &User{Username: "afumu3", Password: "123456", Age: 26, Score: 33, Dept: "研发部门"}
	user4 := &User{Username: "afumu4", Password: "123456", Age: 30, Score: 11, Dept: "产品部门"}
	user5 := &User{Username: "afumu5", Password: "123456", Age: 12, Score: 34, Dept: "生产部门"}
	user6 := &User{Username: "afumu6", Password: "123456", Age: 12, Score: 34, Dept: "生产部门"}
	user7 := &User{Username: "afumu7", Password: "123456", Age: 45, Score: 123, Dept: "销售部门"}
	user8 := &User{Username: "afumu7", Password: "123456", Age: 45, Score: 123, Dept: "销售部门"}
	users := []*User{user1, user2, user3, user4, user5, user6, user7, user8}
	return users
}
