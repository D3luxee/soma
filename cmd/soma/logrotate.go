/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package main

import (
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/mjolnir42/soma/internal/msg"
)

func logrotate(sigChan chan os.Signal) {
	for {
		select {
		case <-sigChan:
		fileloop:
			for name, lfHandle := range logFileMap {
				// treekeeper startup logs do not get rotated
				if strings.HasPrefix(name, `startup_`) {
					continue
				}
				err := lfHandle.Reopen()
				if err != nil {
					log.Printf("Error rotating logfile %s: %s\n", name, err)
					log.Println(`Shutting down system`)

					returnChannel := make(chan msg.Result)
					handler := handlerMap[`grimReaper`].(*grimReaper)
					handler.system <- msg.Request{
						Section:    `runtime`,
						Action:     `shutdown`,
						Reply:      returnChannel,
						RemoteAddr: `::1`,
						AuthUser:   `root`,
					}
					<-returnChannel
					break fileloop
				}
			}
		}
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
