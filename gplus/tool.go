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
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type Condition struct {
	Group       string
	ColumnName  string
	Op          string
	ColumnValue any
}

var columnTypeCache sync.Map

var operators = []string{"!~<=", "!~>=", "~<=", "~>=", "!?=", "!^=", "!~=", "?=", "^=", "~=", "!=", ">=", "<=", "=", ">", "<"}
var builders = map[string]func(query *QueryCond[any], name string, value any){
	"!~<=": notLikeLeft,
	"!~>=": notLikeRight,
	"~<=":  LikeLeft,
	"~>=":  LikeRight,
	"!?=":  notIn,
	"!^=":  notBetween,
	"!~=":  notLike,
	"?=":   in,
	"^=":   between,
	"~=":   like,
	"!=":   ne,
	">=":   ge,
	"<=":   le,
	"=":    eq,
	">":    gt,
	"<":    lt,
}

func BuildQuery[T any](queryParams url.Values) *QueryCond[T] {
	dbConnName := getDefaultDbConnName()
	return BuildQueryBaseDb[T](queryParams, DbConnName(dbConnName))
}

func BuildQueryBaseDb[T any](queryParams url.Values, opt OptionFunc) *QueryCond[T] {

	columnCondMap, conditionMap, gcond := parseParams(queryParams)

	parentQuery := buildParentQuery[T](conditionMap)

	option := getOneOption(opt)
	queryCondMap := buildQueryCondMap[T](columnCondMap, option.DbConnName)

	// 如果没有分组条件，直接返回默认的查询条件
	if len(gcond) == 0 {
		if q, ok := queryCondMap["default"]; ok {
			q.orderBuilder = parentQuery.orderBuilder
			q.selectColumns = parentQuery.selectColumns
			q.omitColumns = parentQuery.omitColumns
			return q
		}

		// 如果没有分组条件，但是有分组设置，返回第一个查询条件。主要为了兼容只有一个分组但是没有设置条件的情况。
		if len(queryCondMap) == 1 {
			for _, q := range queryCondMap {
				q.orderBuilder = parentQuery.orderBuilder
				q.selectColumns = parentQuery.selectColumns
				q.omitColumns = parentQuery.omitColumns
				return q
			}
		}
	}

	return buildGroupQuery[T](gcond, queryCondMap, parentQuery)
}

func parseParams(queryParams url.Values) (map[string][]*Condition, map[string]string, string) {
	var gcond string
	var columnCondMap = make(map[string][]*Condition)
	var conditionMap = make(map[string]string)
	for key, values := range queryParams {
		switch key {
		case "q":
			columnCondMap = buildColumnCondMap(values)
		case "sort":
			if len(values) > 0 {
				conditionMap["sort"] = values[len(values)-1]
			}
		case "select":
			if len(values) > 0 {
				conditionMap["select"] = values[len(values)-1]
			}
		case "omit":
			if len(values) > 0 {
				conditionMap["omit"] = values[len(values)-1]
			}
		case "gcond":
			gcond = values[0]
		}
	}
	return columnCondMap, conditionMap, gcond
}

// buildColumnCondMap 根据url参数构建字段条件
func buildColumnCondMap(values []string) map[string][]*Condition {
	var maps = make(map[string][]*Condition)
	for _, value := range values {
		currentOperator := getCurrentOp(value)
		params := strings.SplitN(value, currentOperator, 2)
		if len(params) == 2 {
			condition := &Condition{}
			groups := strings.Split(params[0], ".")
			var groupName string
			var columnName string
			// 如果不包含组，默认分为同一个组
			if len(groups) == 1 {
				groupName = "default"
				columnName = groups[0]
			} else if len(groups) == 2 {
				groupName = groups[0]
				columnName = groups[1]
			}
			condition.Group = groupName
			condition.ColumnName = columnName
			condition.Op = currentOperator
			condition.ColumnValue = params[1]
			conditions, ok := maps[groupName]
			if ok {
				conditions = append(conditions, condition)
			} else {
				conditions = []*Condition{condition}
			}
			maps[groupName] = conditions
		}
	}
	return maps
}

func getCurrentOp(value string) string {
	var currentOperator string
	for _, op := range operators {
		if strings.Contains(value, op) {
			currentOperator = op
			break
		}
	}
	return currentOperator
}

func buildQueryCondMap[T any](columnCondMap map[string][]*Condition, dbConnName string) map[string]*QueryCond[T] {
	var queryCondMap = make(map[string]*QueryCond[T])
	columnTypeMap := getColumnTypeMap[T](dbConnName)
	for key, conditions := range columnCondMap {
		query := &QueryCond[any]{}
		query.columnTypeMap = columnTypeMap
		for _, condition := range conditions {
			name := condition.ColumnName
			op := condition.Op
			value := condition.ColumnValue
			builders[op](query, name, value)
		}
		newQuery, _ := NewQuery[T]()
		newQuery.queryExpressions = append(newQuery.queryExpressions, query.queryExpressions...)
		queryCondMap[key] = newQuery
	}
	return queryCondMap
}

func buildParentQuery[T any](conditionMap map[string]string) *QueryCond[T] {
	parentQuery, _ := NewQuery[T]()
	for key, value := range conditionMap {
		if key == "sort" {
			orderColumns := strings.Split(value, ",")
			for _, column := range orderColumns {
				if strings.HasPrefix(column, "-") {
					newValue := strings.TrimLeft(column, "-")
					parentQuery.OrderByDesc(newValue)
				} else {
					parentQuery.OrderByAsc(column)
				}
			}
		} else if key == "select" {
			selectColumns := strings.Split(value, ",")
			for _, column := range selectColumns {
				parentQuery.Select(column)
			}
		} else if key == "omit" {
			omitColumns := strings.Split(value, ",")
			for _, column := range omitColumns {
				parentQuery.Omit(column)
			}
		}
	}
	return parentQuery
}

func buildGroupQuery[T any](gcond string, queryMaps map[string]*QueryCond[T], query *QueryCond[T]) *QueryCond[T] {
	var tempQuerys []*QueryCond[T]
	tempQuerys = append(tempQuerys, query)
	for i, char := range gcond {
		str := string(char)
		tempQuery := tempQuerys[len(tempQuerys)-1]
		// 如果是 左括号 开头，则代表需要嵌套查询
		if str == "(" && i != len(gcond)-1 {
			if i != 0 && string(gcond[i-1]) == "|" {
				tempQuery.Or(func(q *QueryCond[T]) {
					paramQuery, isOk := queryMaps[string(gcond[i+1])]
					if isOk {
						q.queryExpressions = append(q.queryExpressions, paramQuery.queryExpressions...)
						tempQuerys = append(tempQuerys, q)
					}
				})
				continue
			} else {
				tempQuery.And(func(q *QueryCond[T]) {
					paramQuery, isOk := queryMaps[string(gcond[i+1])]
					if isOk {
						q.queryExpressions = append(q.queryExpressions, paramQuery.queryExpressions...)
						tempQuerys = append(tempQuerys, q)
					}
				})
			}
			continue
		}

		// 如果当前为 | ,而且不是最后一个字符，而且下一个字符不是 ( ,则为 or
		if str == "|" && i != len(gcond)-1 {
			paramQuery, isOk := queryMaps[string(gcond[i+1])]
			if isOk {
				tempQuery.Or().queryExpressions = append(tempQuery.queryExpressions, paramQuery.queryExpressions...)
				tempQuery.last = paramQuery.queryExpressions[len(paramQuery.queryExpressions)-1]
			}
			continue
		}

		if str == "*" && i != len(gcond)-1 {
			paramQuery, isOk := queryMaps[string(gcond[i+1])]
			if isOk {
				tempQuery.And()
				tempQuery.queryExpressions = append(tempQuery.queryExpressions, paramQuery.queryExpressions...)
				tempQuery.last = paramQuery.queryExpressions[len(paramQuery.queryExpressions)-1]
			}
			continue
		}

		if str == ")" {
			// 删除最后一个query对象
			tempQuerys = tempQuerys[:len(tempQuerys)-1]
			continue
		}

		// 如果上面的条件不满足，而且是第一个的话，那么就直接添加条件
		if i == 0 {
			paramQuery, isOk := queryMaps[string(gcond[i])]
			if isOk {
				tempQuery.queryExpressions = append(tempQuery.queryExpressions, paramQuery.queryExpressions...)
				tempQuery.last = paramQuery.queryExpressions[len(paramQuery.queryExpressions)-1]
			}
		}
	}
	return query
}

func getColumnTypeMap[T any](dbConnName string) map[string]reflect.Type {
	modelTypeStr := reflect.TypeOf((*T)(nil)).Elem().String()
	if model, ok := columnTypeCache.Load(modelTypeStr); ok {
		if columnNameMap, isOk := model.(map[string]reflect.Type); isOk {
			return columnNameMap
		}
	}
	var columnTypeMap = make(map[string]reflect.Type)
	typeOf := reflect.TypeOf((*T)(nil)).Elem()
	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)
		if field.Anonymous {
			nestedFields := getSubFieldColumnTypeMap(field, dbConnName)
			for key, value := range nestedFields {
				columnTypeMap[key] = value
			}
		}
		columnName := parseColumnName(field, dbConnName)
		columnTypeMap[columnName] = field.Type
	}
	columnTypeCache.Store(modelTypeStr, columnTypeMap)
	return columnTypeMap
}

func getSubFieldColumnTypeMap(field reflect.StructField, dbConnName string) map[string]reflect.Type {
	columnTypeMap := make(map[string]reflect.Type)
	modelType := field.Type
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	for j := 0; j < modelType.NumField(); j++ {
		subField := modelType.Field(j)
		if subField.Anonymous {
			nestedFields := getSubFieldColumnTypeMap(subField, dbConnName)
			for key, value := range nestedFields {
				columnTypeMap[key] = value
			}
		} else {
			columnName := parseColumnName(subField, dbConnName)
			columnTypeMap[columnName] = subField.Type
		}
	}
	return columnTypeMap
}

func notLikeLeft(query *QueryCond[any], name string, value any) {
	query.NotLikeLeft(name, convert(query.columnTypeMap, name, value))
}

func notLikeRight(query *QueryCond[any], name string, value any) {
	query.NotLikeRight(name, convert(query.columnTypeMap, name, value))
}

func LikeLeft(query *QueryCond[any], name string, value any) {
	query.LikeLeft(name, convert(query.columnTypeMap, name, value))
}

func LikeRight(query *QueryCond[any], name string, value any) {
	query.LikeRight(name, convert(query.columnTypeMap, name, value))
}

func notIn(query *QueryCond[any], name string, value any) {
	values := strings.Split(fmt.Sprintf("%s", value), ",")
	var queryValues []any
	for _, v := range values {
		queryValues = append(queryValues, convert(query.columnTypeMap, name, v))
	}
	query.NotIn(name, queryValues)
}

func notBetween(query *QueryCond[any], name string, value any) {
	values := strings.Split(fmt.Sprintf("%s", value), ",")
	if len(values) == 2 {
		query.NotBetween(name, convert(query.columnTypeMap, name, values[0]), convert(query.columnTypeMap, name, values[1]))
	}
}

func notLike(query *QueryCond[any], name string, value any) {
	query.NotLike(name, convert(query.columnTypeMap, name, value))
}

func in(query *QueryCond[any], name string, value any) {
	values := strings.Split(fmt.Sprintf("%s", value), ",")
	var queryValues []any
	for _, v := range values {
		queryValues = append(queryValues, convert(query.columnTypeMap, name, v))
	}
	query.In(name, queryValues)
}

func between(query *QueryCond[any], name string, value any) {
	values := strings.Split(fmt.Sprintf("%s", value), ",")
	if len(values) == 2 {
		query.Between(name, convert(query.columnTypeMap, name, values[0]), convert(query.columnTypeMap, name, values[1]))
	}
}

func like(query *QueryCond[any], name string, value any) {
	query.Like(name, convert(query.columnTypeMap, name, value))
}

func ne(query *QueryCond[any], name string, value any) {
	if strings.ToLower(fmt.Sprintf("%s", value)) == "null" {
		query.IsNotNull(name)
	} else {
		query.Ne(name, convert(query.columnTypeMap, name, value))
	}
}

func ge(query *QueryCond[any], name string, value any) {
	query.Ge(name, convert(query.columnTypeMap, name, value))
}

func le(query *QueryCond[any], name string, value any) {
	query.Le(name, convert(query.columnTypeMap, name, value))
}

func eq(query *QueryCond[any], name string, value any) {
	if strings.ToLower(fmt.Sprintf("%s", value)) == "null" {
		query.IsNull(name)
	} else {
		query.Eq(name, convert(query.columnTypeMap, name, value))
	}
}

func gt(query *QueryCond[any], name string, value any) {
	query.Gt(name, convert(query.columnTypeMap, name, value))
}

func lt(query *QueryCond[any], name string, value any) {
	query.Lt(name, convert(query.columnTypeMap, name, value))
}

func convert(columnTypeMap map[string]reflect.Type, name string, value any) any {
	columnType, ok := columnTypeMap[name]
	if ok {
		switch columnType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			atoi, err := strconv.Atoi(fmt.Sprintf("%s", value))
			if err == nil {
				value = atoi
			}
		}
	}
	return value
}
