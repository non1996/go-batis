package sql

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/non1996/go-jsonobj/container"
)

var (
	regexParam = regexp.MustCompile(`#\{(.*?)}`)
	regexProp  = regexp.MustCompile(`\$\{(.*?)}`)
)

func translateProps(stmt string) (string, []string) {
	var gIdx int
	var propMap = container.NewOrderedMap[string, int]()

	stmt = regexProp.ReplaceAllStringFunc(stmt, func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.Trim(s, "${}")
		s = strings.TrimSpace(s)

		var idx int
		if !propMap.Exist(s) {
			idx = gIdx
			propMap.Add(s, idx)
			gIdx++
		} else {
			idx = propMap.Get(s)
		}
		return fmt.Sprintf("${%d}", idx)
	})

	return stmt, propMap.Keys()
}

func translateParams(stmt string) (string, []string) {
	var gIdx int
	var params []string

	stmt = regexParam.ReplaceAllStringFunc(stmt, func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.Trim(s, "#{}")
		s = strings.TrimSpace(s)

		params = append(params, s)

		pl := fmt.Sprintf("#{%d}", gIdx)
		gIdx++
		return pl
	})

	return stmt, params
}
