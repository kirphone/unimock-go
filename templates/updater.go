package templates

import "unimock/util"

type HeaderUpdater struct {
	headerName string
}

func (updater HeaderUpdater) Update(message *util.Message, value string) {
	message.Headers[updater.headerName] = value
}
