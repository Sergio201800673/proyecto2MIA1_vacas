package analizador

import (
	"regexp"
	"strings"
)

func GetComandsParams(input string) (string, string) {
	parts := strings.Fields(input)
	if len(parts) > 0 {
		command := strings.ToLower(parts[0])
		params := strings.Join(parts[1:], " ")
		if params == "" {
			return command, "vacio"
		}
		return command, params
	}
	return "vacio", ""
}

var re = regexp.MustCompile(`-(\w+)=("[^"]+"|\S+)`)

func AnaliceRegExp(params string) [][]string {
	tokens := re.FindAllStringSubmatch(params, -1)
	return tokens
}
