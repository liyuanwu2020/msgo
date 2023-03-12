package msstrings

import (
	"fmt"
	"reflect"
	"strings"
)

func JoinStrings(data ...any) string {
	var sBuilder strings.Builder
	for _, datum := range data {
		sBuilder.WriteString(checkString(datum))
	}
	return sBuilder.String()
}

func checkString(v any) string {
	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.String:
		return v.(string)
	default:
		return fmt.Sprintf("%v", v)
	}
}
