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
	"fmt"
	"github.com/acmestack/gorm-plus/constants"
	"reflect"
	"strings"
)

type QueryCond[T any] struct {
	selectColumns    []string
	omitColumns      []string
	distinctColumns  []string
	queryExpressions []any
	orderBuilder     strings.Builder
	groupBuilder     strings.Builder
	havingBuilder    strings.Builder
	havingArgs       []any
	queryArgs        []any
	last             any
	limit            *int
	offset           int
	updateMap        map[string]any
	columnTypeMap    map[string]reflect.Type
}

func (q *QueryCond[T]) getSqlSegment() string {
	return ""
}

// NewQuery 构建查询条件
func NewQuery[T any]() (*QueryCond[T], *T) {
	dbConnName := getDefaultDbConnName()
	return NewQueryBaseDb[T](DbConnName(dbConnName))
}

// NewQueryBaseDb 构建查询条件
func NewQueryBaseDb[T any](opt OptionFunc) (*QueryCond[T], *T) {
	q := &QueryCond[T]{}
	modelTypeStr := reflect.TypeOf((*T)(nil)).Elem().String()
	if model, ok := modelInstanceCache.Load(modelTypeStr); ok {
		m, isReal := model.(*T)
		if isReal {
			return q, m
		}
	}
	m := new(T)
	Cache(opt, m)
	return q, m
}

// NewQueryModel 构建查询条件
func NewQueryModel[T any, R any]() (*QueryCond[T], *T, *R) {
	dbConnName := getDefaultDbConnName() //兼容之前设计
	return NewQueryModelBaseDb[T, R](DbConnName(dbConnName))
}

// NewQueryModelBaseDb 构建查询条件
func NewQueryModelBaseDb[T any, R any](opt OptionFunc) (*QueryCond[T], *T, *R) {
	q := &QueryCond[T]{}
	var t *T
	var r *R
	entityTypeStr := reflect.TypeOf((*T)(nil)).Elem().String()
	if model, ok := modelInstanceCache.Load(entityTypeStr); ok {
		m, isReal := model.(*T)
		if isReal {
			t = m
		}
	}

	modelTypeStr := reflect.TypeOf((*R)(nil)).Elem().String()
	if model, ok := modelInstanceCache.Load(modelTypeStr); ok {
		m, isReal := model.(*R)
		if isReal {
			r = m
		}
	}

	if t == nil {
		t = new(T)
		Cache(opt, t)
	}

	if r == nil {
		r = new(R)
		Cache(opt, r)
	}

	return q, t, r
}

// Eq 等于 =
func (q *QueryCond[T]) Eq(column any, val any) *QueryCond[T] {
	q.addExpression(q.buildSqlSegment(column, constants.Eq, val)...)
	return q
}

// Ne 不等于 !=
func (q *QueryCond[T]) Ne(column any, val any) *QueryCond[T] {
	q.addExpression(q.buildSqlSegment(column, constants.Ne, val)...)
	return q
}

// Gt 大于 >
func (q *QueryCond[T]) Gt(column any, val any) *QueryCond[T] {
	q.addExpression(q.buildSqlSegment(column, constants.Gt, val)...)
	return q
}

// Ge 大于等于 >=
func (q *QueryCond[T]) Ge(column any, val any) *QueryCond[T] {
	q.addExpression(q.buildSqlSegment(column, constants.Ge, val)...)
	return q
}

// Lt 小于 <
func (q *QueryCond[T]) Lt(column any, val any) *QueryCond[T] {
	q.addExpression(q.buildSqlSegment(column, constants.Lt, val)...)
	return q
}

// Le 小于等于 <=
func (q *QueryCond[T]) Le(column any, val any) *QueryCond[T] {
	q.addExpression(q.buildSqlSegment(column, constants.Le, val)...)
	return q
}

// Like 模糊 LIKE '%值%'
func (q *QueryCond[T]) Like(column any, val any) *QueryCond[T] {
	s := fmt.Sprintf("%v", val)
	q.addExpression(q.buildSqlSegment(column, constants.Like, "%"+s+"%")...)
	return q
}

// NotLike 非模糊 NOT LIKE '%值%'
func (q *QueryCond[T]) NotLike(column any, val any) *QueryCond[T] {
	s := fmt.Sprintf("%v", val)
	q.addExpression(q.buildSqlSegment(column, constants.Not+" "+constants.Like, "%"+s+"%")...)
	return q
}

// LikeLeft 左模糊 LIKE '%值'
func (q *QueryCond[T]) LikeLeft(column any, val any) *QueryCond[T] {
	s := fmt.Sprintf("%v", val)
	q.addExpression(q.buildSqlSegment(column, constants.Like, "%"+s)...)
	return q
}

// NotLikeLeft 非左模糊 NOT LIKE '%值'
func (q *QueryCond[T]) NotLikeLeft(column any, val any) *QueryCond[T] {
	s := fmt.Sprintf("%v", val)
	q.addExpression(q.buildSqlSegment(column, constants.Not+" "+constants.Like, "%"+s)...)
	return q
}

// LikeRight 右模糊 LIKE '值%'
func (q *QueryCond[T]) LikeRight(column any, val any) *QueryCond[T] {
	s := fmt.Sprintf("%v", val)
	q.addExpression(q.buildSqlSegment(column, constants.Like, s+"%")...)
	return q
}

// NotLikeRight 非右模糊 NOT LIKE '值%'
func (q *QueryCond[T]) NotLikeRight(column any, val any) *QueryCond[T] {
	s := fmt.Sprintf("%v", val)
	q.addExpression(q.buildSqlSegment(column, constants.Not+" "+constants.Like, s+"%")...)
	return q
}

// IsNull 是否为空 字段 IS NULL
func (q *QueryCond[T]) IsNull(column any) *QueryCond[T] {
	q.addExpression(q.buildSqlSegment(column, constants.IsNull, "")...)
	return q
}

// IsNotNull 是否非空 字段 IS NOT NULL
func (q *QueryCond[T]) IsNotNull(column any) *QueryCond[T] {
	q.addExpression(q.buildSqlSegment(column, constants.IsNotNull, "")...)
	return q
}

// In 字段 IN (值1, 值2, ...)
func (q *QueryCond[T]) In(column any, val any) *QueryCond[T] {
	q.addExpression(q.buildSqlSegment(column, constants.In, val)...)
	return q
}

// NotIn 字段 NOT IN (值1, 值2, ...)
func (q *QueryCond[T]) NotIn(column any, val any) *QueryCond[T] {
	q.addExpression(q.buildSqlSegment(column, constants.Not+" "+constants.In, val)...)
	return q
}

// Between BETWEEN 值1 AND 值2
func (q *QueryCond[T]) Between(column any, start, end any) *QueryCond[T] {
	q.addExpression(q.buildSqlSegment(column, constants.Between, start, constants.And, end)...)
	return q
}

// NotBetween NOT BETWEEN 值1 AND 值2
func (q *QueryCond[T]) NotBetween(column any, start, end any) *QueryCond[T] {
	q.addExpression(q.buildSqlSegment(column, constants.Not+" "+constants.Between, start, constants.And, end)...)
	return q
}

// Distinct 去除重复字段值
func (q *QueryCond[T]) Distinct(columns ...any) *QueryCond[T] {
	for _, v := range columns {
		q.distinctColumns = append(q.distinctColumns, getColumnName(v))
	}
	return q
}

// Group 分组：GROUP BY 字段1,字段2
func (q *QueryCond[T]) Group(columns ...any) *QueryCond[T] {
	for _, v := range columns {
		columnName := getColumnName(v)
		if q.groupBuilder.Len() > 0 {
			q.groupBuilder.WriteString(constants.Comma)
		}
		q.groupBuilder.WriteString(columnName)
	}
	return q
}

// OrderByDesc 排序：ORDER BY 字段1,字段2 Desc
func (q *QueryCond[T]) OrderByDesc(columns ...any) *QueryCond[T] {
	var columnNames []string
	for _, v := range columns {
		columnName := getColumnName(v)
		columnNames = append(columnNames, columnName)
	}
	q.buildOrder(constants.Desc, columnNames...)
	return q
}

// OrderByAsc 排序：ORDER BY 字段1,字段2 ASC
func (q *QueryCond[T]) OrderByAsc(columns ...any) *QueryCond[T] {
	var columnNames []string
	for _, v := range columns {
		columnName := getColumnName(v)
		columnNames = append(columnNames, columnName)
	}
	q.buildOrder(constants.Asc, columnNames...)
	return q
}

// Having HAVING SQl语句
func (q *QueryCond[T]) Having(having string, args ...any) *QueryCond[T] {
	q.havingBuilder.WriteString(having)
	if len(args) == 1 {
		// 兼容function方法中in返回切片类型数据
		if anies, ok := args[0].([]any); ok {
			q.havingArgs = append(q.havingArgs, anies...)
			return q
		}
	}
	q.havingArgs = append(q.havingArgs, args...)
	return q
}

// And 拼接 AND
func (q *QueryCond[T]) And(fn ...func(q *QueryCond[T])) *QueryCond[T] {
	if len(fn) > 0 {
		// fix bug: https://github.com/acmestack/gorm-plus/issues/74
		q.addExpression(&sqlKeyword{keyword: constants.And})
		nestQuery := &QueryCond[T]{}
		fn[0](nestQuery)
		q.queryExpressions = append(q.queryExpressions, nestQuery)
		q.last = nestQuery
		return q
	}
	q.addExpression(&sqlKeyword{keyword: constants.And})
	return q
}

// Or 拼接 OR
func (q *QueryCond[T]) Or(fn ...func(q *QueryCond[T])) *QueryCond[T] {
	if len(fn) > 0 {
		// fix bug: https://github.com/acmestack/gorm-plus/issues/74
		q.addExpression(&sqlKeyword{keyword: constants.Or})
		nestQuery := &QueryCond[T]{}
		fn[0](nestQuery)
		q.queryExpressions = append(q.queryExpressions, nestQuery)
		q.last = nestQuery
		return q
	}
	q.addExpression(&sqlKeyword{keyword: constants.Or})
	return q
}

// Select 查询字段
func (q *QueryCond[T]) Select(columns ...any) *QueryCond[T] {
	for _, v := range columns {
		columnName := getColumnName(v)
		q.selectColumns = append(q.selectColumns, columnName)
	}
	return q
}

// Omit 忽略字段
func (q *QueryCond[T]) Omit(columns ...any) *QueryCond[T] {
	for _, v := range columns {
		columnName := getColumnName(v)
		q.omitColumns = append(q.omitColumns, columnName)
	}
	return q
}

// Set 设置更新的字段
func (q *QueryCond[T]) Set(column any, val any) *QueryCond[T] {
	columnName := getColumnName(column)
	if q.updateMap == nil {
		q.updateMap = make(map[string]any)
	}
	q.updateMap[columnName] = val
	return q
}

/*
* 自定义条件
 */

// AndCond 拼接 AND
func (q *QueryCond[T]) AndCond(cond bool, fn ...func(q *QueryCond[T])) *QueryCond[T] {
	if cond {
		return q.And(fn...)
	}
	return q
}

// OrCond 拼接 OR
func (q *QueryCond[T]) OrCond(cond bool, fn ...func(q *QueryCond[T])) *QueryCond[T] {
	if cond {
		return q.Or(fn...)
	}
	return q
}

// EqCond 等于 =
func (q *QueryCond[T]) EqCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Eq(column, val)
	}
	return q
}

// NeCond 不等于 !=
func (q *QueryCond[T]) NeCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Ne(column, val)
	}
	return q
}

// GtCond 大于 >
func (q *QueryCond[T]) GtCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Gt(column, val)
	}
	return q
}

// GeCond 大于等于 >=
func (q *QueryCond[T]) GeCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Ge(column, val)
	}
	return q
}

// LtCond 小于 <
func (q *QueryCond[T]) LtCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Lt(column, val)
	}
	return q
}

// LeCond 小于等于 <=
func (q *QueryCond[T]) LeCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Le(column, val)
	}
	return q
}

// LikeCond 模糊 LIKE '%值%'
func (q *QueryCond[T]) LikeCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Like(column, val)
	}
	return q
}

// NotLikeCond 非模糊 NOT LIKE '%值%'
func (q *QueryCond[T]) NotLikeCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.NotLike(column, val)
	}
	return q
}

// LikeLeftCond 左模糊 LIKE '%值'
func (q *QueryCond[T]) LikeLeftCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.LikeLeft(column, val)
	}
	return q
}

// NotLikeLeftCond 非左模糊 NOT LIKE '%值'
func (q *QueryCond[T]) NotLikeLeftCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.NotLike(column, val)
	}
	return q
}

// LikeRightCond 右模糊 LIKE '值%'
func (q *QueryCond[T]) LikeRightCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.LikeRight(column, val)
	}
	return q
}

// NotLikeRightCond 非右模糊 NOT LIKE '值%'
func (q *QueryCond[T]) NotLikeRightCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.NotLikeRight(column, val)
	}
	return q
}

// InCond 字段 IN (值1, 值2, ...)
func (q *QueryCond[T]) InCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.In(column, val)
	}
	return q
}

// AndEqCond 并且等于 =
func (q *QueryCond[T]) AndEqCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.And().Eq(column, val)
	}
	return q
}

// AndNeCond 并且不等于 !=
func (q *QueryCond[T]) AndNeCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.And().Ne(column, val)
	}
	return q
}

// AndGtCond 并且大于 >
func (q *QueryCond[T]) AndGtCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.And().Gt(column, val)
	}
	return q
}

// AndGeCond 并且大于等于 >=
func (q *QueryCond[T]) AndGeCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.And().Ge(column, val)
	}
	return q
}

// AndLtCond 并且小于 <
func (q *QueryCond[T]) AndLtCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.And().Lt(column, val)
	}
	return q
}

// AndLeCond 并且小于等于 <=
func (q *QueryCond[T]) AndLeCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.And().Le(column, val)
	}
	return q
}

// AndLikeCond 并且模糊 LIKE '%值%'
func (q *QueryCond[T]) AndLikeCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.And().Like(column, val)
	}
	return q
}

// AndNotLikeCond 并且非模糊 NOT LIKE '%值%'
func (q *QueryCond[T]) AndNotLikeCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.And().NotLike(column, val)
	}
	return q
}

// AndLikeLeftCond 并且左模糊 LIKE '%值'
func (q *QueryCond[T]) AndLikeLeftCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.And().LikeLeft(column, val)
	}
	return q
}

// AndNotLikeLeftCond 并且非左模糊 NOT LIKE '%值'
func (q *QueryCond[T]) AndNotLikeLeftCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.And().NotLikeLeft(column, val)
	}
	return q
}

// AndLikeRightCond 并且右模糊 LIKE '值%'
func (q *QueryCond[T]) AndLikeRightCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.And().LikeRight(column, val)
	}
	return q
}

// AndNotLikeRightCond 并且非右模糊 NOT LIKE '值%'
func (q *QueryCond[T]) AndNotLikeRightCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.And().NotLikeRight(column, val)
	}
	return q
}

// AndInCond 并且字段 IN (值1, 值2, ...)
func (q *QueryCond[T]) AndInCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.And().In(column, val)
	}
	return q
}

// OrEqCond 或者等于 =
func (q *QueryCond[T]) OrEqCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Or().Eq(column, val)
	}
	return q
}

// OrNeCond 或者不等于 !=
func (q *QueryCond[T]) OrNeCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Or().Ne(column, val)
	}
	return q
}

// OrGtCond 或者大于 >
func (q *QueryCond[T]) OrGtCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Or().Gt(column, val)
	}
	return q
}

// OrGeCond 或者大于等于 >=
func (q *QueryCond[T]) OrGeCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Or().Ge(column, val)
	}
	return q
}

// OrLtCond 或者小于 <
func (q *QueryCond[T]) OrLtCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Or().Lt(column, val)
	}
	return q
}

// OrLeCond 或者小于等于 <=
func (q *QueryCond[T]) OrLeCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Or().Le(column, val)
	}
	return q
}

// OrLikeCond 或者模糊 LIKE '%值%'
func (q *QueryCond[T]) OrLikeCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Or().Like(column, val)
	}
	return q
}

// OrNotLikeCond 或者非模糊 NOT LIKE '%值%'
func (q *QueryCond[T]) OrNotLikeCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Or().NotLike(column, val)
	}
	return q
}

// OrLikeLeftCond 或者左模糊 LIKE '%值'
func (q *QueryCond[T]) OrLikeLeftCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Or().LikeLeft(column, val)
	}
	return q
}

// OrNotLikeLeftCond 或者非左模糊 NOT LIKE '%值'
func (q *QueryCond[T]) OrNotLikeLeftCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Or().NotLikeLeft(column, val)
	}
	return q
}

// OrLikeRightCond 或者右模糊 LIKE '值%'
func (q *QueryCond[T]) OrLikeRightCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Or().LikeRight(column, val)
	}
	return q
}

// OrNotLikeRightCond 或者非右模糊 NOT LIKE '值%'
func (q *QueryCond[T]) OrNotLikeRightCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Or().NotLikeRight(column, val)
	}
	return q
}

// OrInCond 或者字段 IN (值1, 值2, ...)
func (q *QueryCond[T]) OrInCond(cond bool, column any, val any) *QueryCond[T] {
	if cond {
		return q.Or().In(column, val)
	}
	return q
}

func (q *QueryCond[T]) addExpression(sqlSegments ...SqlSegment) {
	if len(sqlSegments) == 1 {
		q.handleSingle(sqlSegments[0])
		return
	}

	// 如果符合条件，则添加and。
	q.addAndCondIfNeed()

	for _, sqlSegment := range sqlSegments {
		q.queryExpressions = append(q.queryExpressions, sqlSegment)
	}

	if len(sqlSegments) > 0 {
		q.last = sqlSegments[len(sqlSegments)-1]
	}
}

func (q *QueryCond[T]) addAndCondIfNeed() {
	// 1.如果上一个不是keyword，并且表达式切片不为空，则添加and。
	// 2.如果上一个值不是and或or,并且表达式切片不为空,则添加and。
	lastKeyword, isKeyword := q.last.(*sqlKeyword)
	if isNotKeyword(isKeyword, q.queryExpressions) || isLastNotAndOr(lastKeyword, isKeyword, q.queryExpressions) {
		sk := sqlKeyword{keyword: constants.And}
		q.queryExpressions = append(q.queryExpressions, &sk)
		q.last = &sk
	}
}

func (q *QueryCond[T]) handleSingle(sqlSegment SqlSegment) {
	// 如何是第一次设置，则不需要添加and(),or(),防止用户首次设置条件错误
	if len(q.queryExpressions) == 0 {
		return
	}

	// 防止用户重复设置and(),or()
	isRepeat := q.handelRepeat(sqlSegment)

	// 如果不是重复设置，则添加
	if !isRepeat {
		q.queryExpressions = append(q.queryExpressions, sqlSegment)
		q.last = sqlSegment
	}
}

func (q *QueryCond[T]) handelRepeat(sqlSegment SqlSegment) bool {
	currentKeyword, isCurrentKeyword := sqlSegment.(*sqlKeyword)
	lastKeyword, isLastKeyword := q.last.(*sqlKeyword)
	if isCurrentKeyword && isLastKeyword {
		// 如果上一次是and，这一次也是and，那么就不需要重复设置了
		isAnd := lastKeyword.keyword == constants.And
		if isAnd && currentKeyword.keyword == constants.And {
			return true
		}
		// 如果上一次是or，这一次也是or，那么就不需要重复设置了
		isOr := lastKeyword.keyword == constants.Or
		if isOr && currentKeyword.keyword == constants.Or {
			return true
		}
		// 如果上一次是and，这次是or，那么删除上一次的值，使用最新的值
		// 如果上一次是or，这次是and，那么删除上一次的值，使用最新的值
		q.queryExpressions = append(q.queryExpressions, q.queryExpressions[:len(q.queryExpressions)-1]...)
	}
	return false
}

func isNotKeyword(isKeyword bool, expressions []any) bool {
	return !isKeyword && len(expressions) > 0
}

func isLastNotAndOr(lastKeyword *sqlKeyword, isKeyword bool, expressions []any) bool {
	return isKeyword && lastKeyword.keyword != constants.And && lastKeyword.keyword != constants.Or && len(expressions) > 0
}

func (q *QueryCond[T]) buildSqlSegment(column any, condType string, values ...any) []SqlSegment {
	var sqlSegments []SqlSegment
	sqlSegments = append(sqlSegments, &columnPointer{column: column}, &sqlKeyword{keyword: condType})
	for _, val := range values {
		cv := columnValue{value: val}
		sqlSegments = append(sqlSegments, &cv)
	}
	return sqlSegments
}

func (q *QueryCond[T]) buildOrder(orderType string, columns ...string) {
	for _, v := range columns {
		if q.orderBuilder.Len() > 0 {
			q.orderBuilder.WriteString(constants.Comma)
		}
		q.orderBuilder.WriteString(v)
		q.orderBuilder.WriteString(" ")
		q.orderBuilder.WriteString(orderType)
	}
}
