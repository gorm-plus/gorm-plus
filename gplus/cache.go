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
	"gorm.io/gorm/schema"
	"reflect"
	"sync"
)

// 缓存项目中所有实体字段名，储存格式：key为字段指针值，value为字段名
// 通过缓存实体的字段名，方便gplus通过字段指针获取到对应的字段名
var columnNameCache sync.Map

// 缓存实体对象，主要给NewQuery方法返回使用
var modelInstanceCache sync.Map

// Cache 缓存实体对象所有的字段名
func Cache(opt Option, models ...any) {
	db, _, _ := getDefaultDbByOpt(opt)
	for _, model := range models {
		columnNameMap := getColumnNameMap(model, db.Config.NamingStrategy)
		for pointer, columnName := range columnNameMap {
			columnNameCache.Store(pointer, columnName)
		}
		// 缓存对象
		modelTypeStr := reflect.TypeOf(model).Elem().String()
		modelInstanceCache.Store(modelTypeStr, model)
	}
}

func getColumnNameMap(model any, namingStrategy schema.Namer) map[uintptr]string {
	var columnNameMap = make(map[uintptr]string)
	valueOf := reflect.ValueOf(model).Elem()
	typeOf := reflect.TypeOf(model).Elem()
	for i := 0; i < valueOf.NumField(); i++ {
		field := typeOf.Field(i)
		// 如果当前实体嵌入了其他实体，同样需要缓存它的字段名
		if field.Anonymous {
			// 如果存在多重嵌套，通过递归方式获取他们的字段名
			subFieldMap := getSubFieldColumnNameMap(valueOf, field, namingStrategy)
			for pointer, columnName := range subFieldMap {
				columnNameMap[pointer] = columnName
			}
		} else {
			// 获取对象字段指针值
			pointer := valueOf.Field(i).Addr().Pointer()
			columnName := parseColumnName(field, namingStrategy)
			columnNameMap[pointer] = columnName
		}
	}
	return columnNameMap
}

// GetModel 获取
func GetModel[T any](opts ...OptionFunc) *T {
	opt := getDefaultOptionInfo(opts...) //兼容设计
	modelTypeStr := reflect.TypeOf((*T)(nil)).Elem().String()
	if model, ok := modelInstanceCache.Load(modelTypeStr); ok {
		m, isReal := model.(*T)
		if isReal {
			return m
		}
	}
	t := new(T)
	Cache(opt, t)
	return t
}

// 递归获取嵌套字段名
func getSubFieldColumnNameMap(valueOf reflect.Value, field reflect.StructField, namingStrategy schema.Namer) map[uintptr]string {
	result := make(map[uintptr]string)
	modelType := field.Type
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	for j := 0; j < modelType.NumField(); j++ {
		subField := modelType.Field(j)
		if subField.Anonymous {
			nestedFields := getSubFieldColumnNameMap(valueOf, subField, namingStrategy)
			for key, value := range nestedFields {
				result[key] = value
			}
		} else {
			pointer := valueOf.FieldByName(modelType.Field(j).Name).Addr().Pointer()
			name := parseColumnName(modelType.Field(j), namingStrategy)
			result[pointer] = name
		}
	}

	return result
}

// 解析字段名称 兼容多数据库切换
func parseColumnName(field reflect.StructField, namingStrategy schema.Namer) string {
	tagSetting := schema.ParseTagSetting(field.Tag.Get("gorm"), ";")
	name, ok := tagSetting["COLUMN"]
	if ok {
		return name
	}
	return namingStrategy.ColumnName("", field.Name)
}

func getColumnName(v any) string {
	var columnName string
	valueOf := reflect.ValueOf(v)
	switch valueOf.Kind() {
	case reflect.String:
		return v.(string)
	case reflect.Pointer:
		if name, ok := columnNameCache.Load(valueOf.Pointer()); ok {
			return name.(string)
		}
		// 如果是Function类型，解析字段名称
		if reflect.TypeOf(v).Elem() == reflect.TypeOf((*Function)(nil)).Elem() {
			f := v.(*Function)
			return f.funStr
		}
	}
	return columnName
}
