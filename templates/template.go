package templates

import (
	"regexp"
	"strconv"
	"strings"
	"unimock/util"
)

var variableRegexp = regexp.MustCompile(`\$\{\d+}`)

type Template struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	Subsystem  string `json:"subsystem"`
	extractors map[int]MessageExtractor
}

func (template *Template) validate() bool {
	return template.Name != ""
}

func (template *Template) ProcessMessage(message *util.Message) *util.Message {

	resultBody := variableRegexp.ReplaceAllStringFunc(template.Body, func(match string) string {
		idStr := strings.TrimSuffix(strings.TrimPrefix(match, "${"), "}")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return match
		}

		if extractor, ok := template.extractors[id]; ok {
			if result, ok2 := extractor.Extract(message); ok2 {
				return result
			}
		}

		return match
	})

	return &util.Message{
		Body:    resultBody,
		Headers: map[string]string{},
	}
}
