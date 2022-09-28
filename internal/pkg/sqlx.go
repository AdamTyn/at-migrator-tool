package pkg

import (
	"fmt"
	"strings"
)

func Int64InSQL(ins []int64) string {
	var b strings.Builder
	for i := 0; i < len(ins)-1; i++ {
		b.WriteString(fmt.Sprintf("%d,", ins[i]))
	}
	b.WriteString(fmt.Sprintf("%d", ins[len(ins)-1]))
	return b.String()
}

func StringInSQL(ss []string) string {
	var b strings.Builder
	for i := 0; i < len(ss)-1; i++ {
		b.WriteString(fmt.Sprintf("'%s',", ss[i]))
	}
	b.WriteString(fmt.Sprintf("'%s'", ss[len(ss)-1]))
	return b.String()
}
