package jdebug

import (
	"fmt"
	"jlog"
	"reflect"
	"runtime/debug"
	"strings"
)

// ------------------------- outside -------------------------

// 结构体转字符串
func StructToString(s any) string {
	// 获取结构体类型和字段值
	val := reflect.ValueOf(s)
	typ := reflect.TypeOf(s)

	// 确保输入的是结构体类型
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // 获取指针指向的实际结构体
		typ = typ.Elem()
	}
	var builder strings.Builder
	builder.WriteString("\n")
	builder.WriteString(typ.Name() + ":")
	// 遍历结构体的字段
	for i := range val.NumField() {
		fieldName := typ.Field(i).Name
		if fieldName[0] >= 'A' && fieldName[0] <= 'Z' {
			field := val.Field(i)
			builder.WriteString(fmt.Sprintf("\n%s = %v", fieldName, field.Interface()))
		}
	}
	return builder.String()
}

// 获取调用栈
func ShowStack() {
	jlog.Debug(string(debug.Stack()))
}
