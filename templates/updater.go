package templates

import "unimock/util"

type MessageUpdater interface {
	Update(message *util.Message, value string)
}

type HeaderUpdater struct {
	headerName string
}

func (updater HeaderUpdater) Update(message *util.Message, value string) {
	message.Headers[updater.headerName] = value
}
