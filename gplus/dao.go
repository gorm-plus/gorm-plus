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
package gplus

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/acmestack/gorm-plus/constants"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/gorm/utils"
	"reflect"
	"strings"
	"time"
)

var globalDbMap = make(map[string]*gorm.DB)
var globalDbKeys []string
var defaultBatchSize = 1000

func Init(db *gorm.DB) {
	InitDb(db, constants.DefaultGormPlusConnName)
}

func InitDb(db *gorm.DB, dbConnName string) error {
	if len(dbConnName) == 0 {
		return errors.New("InitMultiple dbConnName is empty please check")
	}
	_, exists := globalDbMap[dbConnName]
	if !exists {
		// db instance register to global variable
		globalDbMap[dbConnName] = db
		globalDbKeys = append(globalDbKeys, dbConnName)
		return nil
	}
	return errors.New("InitMultiple have same name:" + dbConnName + ",please check")
}

// GetDb 获取数据库连接
func GetDb(dbConnName string) (*gorm.DB, error) {
	db, exists := globalDbMap[dbConnName]
	if exists {
		return db, nil
	}
	return nil, errors.New("MultipleDbChange not exists dbConn:" + dbConnName + ",please check")
}

type Page[T any] struct {
	Current    int   `json:"current"`
	Size       int   `json:"size"`
	Total      int64 `json:"total"`
	Records    []*T  `json:"records"`
	RecordsMap []T   `json:"recordsMap"`
}

type Dao[T any] struct{}

func (dao Dao[T]) NewQuery() (*QueryCond[T], *T) {
	return NewQuery[T]()
}

func (dao Dao[T]) NewQueryBaseDb(opt OptionFunc) (*QueryCond[T], *T) {
	return NewQueryBaseDb[T](opt)
}

func NewPage[T any](current, size int) *Page[T] {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	return &Page[T]{Current: current, Size: size}
}

type Comparable interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64 | time.Time
}

type StreamingPage[T any, V Comparable] struct {
	ColumnName any   `json:"columnName"` // 进行分页的列字段名称
	StartValue V     `json:"startValue"` // 分页起始值
	Limit      int   `json:"limit"`      // 页大小
	Forward    bool  `json:"forward"`    // 上下页翻页标识
	Total      int64 `json:"total"`      // 总记录数
	Records    []*T  `json:"records"`    // 查询记录
	RecordsMap []T   `json:"recordsMap"` // 查询记录Map
}

func NewStreamingPage[T any, V Comparable](columnName any, startValue V, limit int) *StreamingPage[T, V] {
	return &StreamingPage[T, V]{
		ColumnName: columnName,
		StartValue: startValue,
		Limit:      limit,
		Forward:    true,
	}
}

// Insert 插入一条记录
func Insert[T any](entity *T, opts ...OptionFunc) *gorm.DB {
	db := getDb(opts...)
	resultDb := db.Create(entity)
	return resultDb
}

// InsertBatch 批量插入多条记录
func InsertBatch[T any](entities []*T, opts ...OptionFunc) *gorm.DB {
	db := getDb(opts...)
	if len(entities) == 0 {
		return db
	}
	resultDb := db.CreateInBatches(entities, defaultBatchSize)
	return resultDb
}

// InsertBatchSize 批量插入多条记录
func InsertBatchSize[T any](entities []*T, batchSize int, opts ...OptionFunc) *gorm.DB {
	db := getDb(opts...)
	if len(entities) == 0 {
		return db
	}
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}
	resultDb := db.CreateInBatches(entities, batchSize)
	return resultDb
}

// DeleteById 根据 ID 删除记录
func DeleteById[T any](id any, opts ...OptionFunc) *gorm.DB {
	db := getDb(opts...)
	var entity T
	resultDb := db.Where(getPkColumnName[T](), id).Delete(&entity)
	return resultDb
}

// DeleteByIds 根据 ID 批量删除记录
func DeleteByIds[T any](ids any, opts ...OptionFunc) *gorm.DB {
	q, _ := NewQuery[T]()
	q.In(getPkColumnName[T](), ids)
	resultDb := Delete[T](q, opts...)
	return resultDb
}

// Delete 根据条件删除记录
func Delete[T any](q *QueryCond[T], opts ...OptionFunc) *gorm.DB {
	var entity T
	resultDb := buildCondition[T](q, opts...)
	resultDb.Delete(&entity)
	return resultDb
}

// UpdateById 根据 ID 更新,默认零值不更新
func UpdateById[T any](entity *T, opts ...OptionFunc) *gorm.DB {
	db := getDb(opts...)
	resultDb := db.Model(entity).Updates(entity)
	return resultDb
}

// UpdateZeroById 根据 ID 零值更新
func UpdateZeroById[T any](entity *T, opts ...OptionFunc) *gorm.DB {
	db := getDb(opts...)

	// 如果用户没有设置选择更新的字段，默认更新所有的字段，包括零值更新
	updateAllIfNeed(entity, opts, db)

	resultDb := db.Model(entity).Updates(entity)
	return resultDb
}

func updateAllIfNeed(entity any, opts []OptionFunc, db *gorm.DB) {
	option := getOption(opts)
	if len(option.Selects) == 0 {
		columnNameMap := getColumnNameMap(entity, option.DbConnName)
		var columnNames []string
		for _, columnName := range columnNameMap {
			columnNames = append(columnNames, columnName)
		}
		db.Select(columnNames)
	}
}

// Update 根据 Map 更新
func Update[T any](q *QueryCond[T], opts ...OptionFunc) *gorm.DB {
	resultDb := buildCondition[T](q, opts...)
	resultDb.Updates(&q.updateMap)
	return resultDb
}

// SelectById 根据 ID 查询单条记录
func SelectById[T any](id any, opts ...OptionFunc) (*T, *gorm.DB) {
	q, _ := NewQuery[T]()
	q.Eq(getPkColumnName[T](), id)
	var entity T
	resultDb := buildCondition(q, opts...)
	return &entity, resultDb.Take(&entity)
}

// SelectByIds 根据 ID 查询多条记录
func SelectByIds[T any](ids any, opts ...OptionFunc) ([]*T, *gorm.DB) {
	q, _ := NewQuery[T]()
	q.In(getPkColumnName[T](), ids)
	return SelectList[T](q, opts...)
}

// SelectOne 根据条件查询单条记录
func SelectOne[T any](q *QueryCond[T], opts ...OptionFunc) (*T, *gorm.DB) {
	var entity T
	resultDb := buildCondition(q, opts...)
	return &entity, resultDb.Take(&entity)
}

// SelectList 根据条件查询多条记录
func SelectList[T any](q *QueryCond[T], opts ...OptionFunc) ([]*T, *gorm.DB) {
	resultDb := buildCondition(q, opts...)
	var results []*T
	resultDb.Find(&results)
	return results, resultDb
}

// SelectPage 根据条件分页查询记录
func SelectPage[T any](page *Page[T], q *QueryCond[T], opts ...OptionFunc) (*Page[T], *gorm.DB) {
	option := getOption(opts)

	// 如果需要分页忽略总数，不查询总数
	if !option.IgnoreTotal {
		total, countDb := SelectCount[T](q, opts...)
		if countDb.Error != nil {
			return page, countDb
		}
		page.Total = total
	}

	resultDb := buildCondition(q, opts...)
	var results []*T
	resultDb.Scopes(paginate(page)).Find(&results)
	page.Records = results
	return page, resultDb
}

// SelectStreamingPage 根据条件分页查询记录
func SelectStreamingPage[T any, V Comparable](page *StreamingPage[T, V], q *QueryCond[T], opts ...OptionFunc) (*StreamingPage[T, V], *gorm.DB) {
	option := getOption(opts)

	// 如果需要分页忽略总数，不查询总数
	if !option.IgnoreTotal {
		total, countDb := SelectCount[T](q, opts...)
		if countDb.Error != nil {
			return page, countDb
		}
		page.Total = total
	}

	resultDb := buildCondition(q, opts...)
	var results []*T
	resultDb.Scopes(streamingPaginate(page)).Find(&results)
	page.Records = results
	return page, resultDb
}

// SelectCount 根据条件查询记录数量
func SelectCount[T any](q *QueryCond[T], opts ...OptionFunc) (int64, *gorm.DB) {
	var count int64
	resultDb := buildCondition(q, opts...)
	//fix 查询有设置Select并且数量只有一个且有设置别名,生成sql不对问题
	resultDb.Statement.Selects = nil
	resultDb.Count(&count)
	return count, resultDb
}

// Exists 根据条件判断记录是否存在
func Exists[T any](q *QueryCond[T], opts ...OptionFunc) (bool, *gorm.DB) {
	count, resultDb := SelectCount[T](q, opts...)
	return count > 0, resultDb
}

// SelectPageGeneric 根据传入的泛型封装分页记录
// 第一个泛型代表数据库表实体
// 第二个泛型代表返回记录实体
func SelectPageGeneric[T any, R any](page *Page[R], q *QueryCond[T], opts ...OptionFunc) (*Page[R], *gorm.DB) {
	option := getOption(opts)
	// 如果需要分页忽略总数，不查询总数
	if !option.IgnoreTotal {
		total, countDb := SelectCount[T](q, opts...)
		if countDb.Error != nil {
			return page, countDb
		}
		page.Total = total
	}
	resultDb := buildCondition(q, opts...)
	var r R
	switch any(r).(type) {
	case map[string]any:
		var results []R
		resultDb.Scopes(paginate(page)).Scan(&results)
		page.RecordsMap = results
	default:
		var results []*R
		resultDb.Scopes(paginate(page)).Scan(&results)
		page.Records = results
	}
	return page, resultDb
}

// SelectStreamingPageGeneric 根据传入的泛型封装分页记录
// 第一个泛型代表数据库表实体
// 第二个泛型代表返回记录实体
func SelectStreamingPageGeneric[T any, R any, V Comparable](page *StreamingPage[R, V], q *QueryCond[T], opts ...OptionFunc) (*StreamingPage[R, V], *gorm.DB) {
	option := getOption(opts)
	// 如果需要分页忽略总数，不查询总数
	if !option.IgnoreTotal {
		total, countDb := SelectCount[T](q, opts...)
		if countDb.Error != nil {
			return page, countDb
		}
		page.Total = total
	}
	resultDb := buildCondition(q, opts...)
	var r R
	switch any(r).(type) {
	case map[string]any:
		var results []R
		resultDb.Scopes(streamingPaginate(page)).Scan(&results)
		page.RecordsMap = results
	default:
		var results []*R
		resultDb.Scopes(streamingPaginate(page)).Scan(&results)
		page.Records = results
	}
	return page, resultDb
}

// SelectGeneric 根据传入的泛型封装记录
// 第一个泛型代表数据库表实体
// 第二个泛型代表返回记录实体
func SelectGeneric[T any, R any](q *QueryCond[T], opts ...OptionFunc) (R, *gorm.DB) {
	var entity R
	resultDb := buildCondition(q, opts...)
	return entity, resultDb.Scan(&entity)
}

func Begin(opts ...*sql.TxOptions) *gorm.DB {
	db := getDb()
	return db.Begin(opts...)
}

// Tx 事务
func Tx(txFunc func(tx *gorm.DB) error, opts ...OptionFunc) error {
	db := getDb(opts...)
	return db.Transaction(txFunc)
}

// paginate offset分页
func paginate[T any](p *Page[T]) func(db *gorm.DB) *gorm.DB {
	page := p.Current
	pageSize := p.Size
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 10
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// streamingPaginate 流式分页，根据自增ID、雪花ID、时间等数值类型或者时间类型分页
// Tips: 相比于 offset 分页性能更好，走的是 range，缺点是没办法跳页查询
func streamingPaginate[T any, V Comparable](p *StreamingPage[T, V]) func(db *gorm.DB) *gorm.DB {
	column := getColumnName(p.ColumnName)
	startValue := p.StartValue
	limit := p.Limit
	return func(db *gorm.DB) *gorm.DB {
		// 下一页
		if p.Forward {
			return db.Where(fmt.Sprintf("%v > ?", column), startValue).Limit(limit)
		}
		// 上一页
		return db.Where(fmt.Sprintf("%v < ?", column), startValue).Order(fmt.Sprintf("%v DESC", column)).Limit(limit)
	}
}

func buildCondition[T any](q *QueryCond[T], opts ...OptionFunc) *gorm.DB {
	db := getDb(opts...)
	resultDb := db.Model(new(T))
	if q != nil {
		// 这里清空参数，避免用户重复使用一个query条件
		q.queryArgs = make([]any, 0)

		if len(q.distinctColumns) > 0 {
			resultDb.Distinct(q.distinctColumns)
		}

		if len(q.selectColumns) > 0 {
			resultDb.Select(q.selectColumns)
		}

		if len(q.omitColumns) > 0 {
			resultDb.Omit(q.omitColumns...)
		}

		expressions := q.queryExpressions
		if len(expressions) > 0 {
			var sqlBuilder strings.Builder
			q.queryArgs = buildSqlAndArgs[T](expressions, &sqlBuilder, q.queryArgs)
			resultDb.Where(sqlBuilder.String(), q.queryArgs...)
		}

		if q.orderBuilder.Len() > 0 {
			resultDb.Order(q.orderBuilder.String())
		}

		if q.groupBuilder.Len() > 0 {
			resultDb.Group(q.groupBuilder.String())
		}

		if q.havingBuilder.Len() > 0 {
			resultDb.Having(q.havingBuilder.String(), q.havingArgs...)
		}

		if q.limit != nil {
			resultDb.Limit(*q.limit)
		}

		if q.offset != 0 {
			resultDb.Offset(q.offset)
		}
	}
	return resultDb
}

func buildSqlAndArgs[T any](expressions []any, sqlBuilder *strings.Builder, queryArgs []any) []any {
	for _, v := range expressions {
		// 判断是否是columnValue类型
		switch segment := v.(type) {
		case *columnPointer:
			sqlBuilder.WriteString(segment.getSqlSegment() + " ")
		case *sqlKeyword:
			sqlBuilder.WriteString(segment.getSqlSegment() + " ")
		case *columnValue:
			if segment.value == constants.And {
				sqlBuilder.WriteString(segment.value.(string) + " ")
				continue
			}
			if segment.value != "" {
				sqlBuilder.WriteString("? ")
				queryArgs = append(queryArgs, segment.value)
			}
		case *QueryCond[T]:
			// 当子条件不存在查询表达式时，无需进行递归处理
			if len(segment.queryExpressions) == 0 {
				continue
			}
			sqlBuilder.WriteString(constants.LeftBracket + " ")
			// 递归处理条件
			queryArgs = buildSqlAndArgs[T](segment.queryExpressions, sqlBuilder, queryArgs)
			sqlBuilder.WriteString(constants.RightBracket + " ")
		}
	}
	return queryArgs
}

func getDb(opts ...OptionFunc) *gorm.DB {
	var db *gorm.DB
	option := getOption(opts)

	if option.Db != nil {
		db = option.Db
	} else {
		if len(option.DbConnName) == 0 {
			option.DbConnName = getDefaultDbConnName()
		}
		db, _ = GetDb(option.DbConnName)
	}

	//设置session,如果需要子句仅在当前会话生效，先调用 Session()，再调用 Clauses()。
	setSessionIfNeed(option, db)

	// Clauses()目的是为了初始化Db，如果db已经被初始化了,会直接返回db
	db = db.Clauses()

	// 设置需要忽略的字段
	setOmitIfNeed(option, db)

	// 设置选择的字段
	setSelectIfNeed(option, db)

	return db
}

func getOption(opts []OptionFunc) Option {
	var config Option
	for _, op := range opts {
		op(&config)
	}
	return config
}

func getOneOption(opt OptionFunc) Option {
	var config Option
	opt(&config)
	return config
}

func setSelectIfNeed(option Option, db *gorm.DB) {
	if len(option.Selects) > 0 {
		var columnNames []string
		for _, column := range option.Selects {
			columnName := getColumnName(column)
			columnNames = append(columnNames, columnName)
		}
		db.Select(columnNames)
	}
}

func setOmitIfNeed(option Option, db *gorm.DB) {
	if len(option.Omits) > 0 {
		var columnNames []string
		for _, column := range option.Omits {
			columnName := getColumnName(column)
			columnNames = append(columnNames, columnName)
		}
		db.Omit(columnNames...)
	}
}

func setSessionIfNeed(option Option, db *gorm.DB) {
	if option.DbSession != nil {
		db.Session(option.DbSession)
	}
}

func getPkColumnName[T any]() string {
	var entity T
	entityType := reflect.TypeOf(entity)
	numField := entityType.NumField()
	var columnName string
	for i := 0; i < numField; i++ {
		field := entityType.Field(i)
		tagSetting := schema.ParseTagSetting(field.Tag.Get("gorm"), ";")
		isPrimaryKey := utils.CheckTruth(tagSetting["PRIMARYKEY"], tagSetting["PRIMARY_KEY"])
		if isPrimaryKey {
			name, ok := tagSetting["COLUMN"]
			if !ok {
				namingStrategy := schema.NamingStrategy{}
				name = namingStrategy.ColumnName("", field.Name)
			}
			columnName = name
			break
		}
	}
	if columnName == "" {
		return constants.DefaultPrimaryName
	}
	return columnName
}
