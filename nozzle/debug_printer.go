// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package nozzle

import (
	"fmt"
	"github.com/cloudfoundry/gosteno"
)

//BMDebugPrinter prints consumer messages to log
type BMDebugPrinter struct {
	logger *gosteno.Logger
}

type bmDebugPrinterMessage struct {
	Title, Body string
}

//Print prints message to log
func (p *BMDebugPrinter) Print(title, body string) {
	p.logger.Debug(fmt.Sprintf("BMPrinter message %s: <%s>", title, body))
}
