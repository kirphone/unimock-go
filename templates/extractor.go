package templates

import "unimock/util"

type MessageExtractor interface {
	Extract(message *util.Message) (string, bool)
}

type HeaderExtractor struct {
	headerName string
}

func (extractor HeaderExtractor) Extract(message *util.Message) (string, bool) {
	value, ok := message.Headers[extractor.headerName]
	return value, ok
}
