// Modifications Copyright 2018 The klaytn Authors
// Copyright 2014 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from cmd/utils/cmd.go (2018/06/04).
// Modified and improved for the klaytn development.

package utils

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/aidan-kwon/toy-klaytn/log"
	"github.com/aidan-kwon/toy-klaytn/node"
)

const (
	importBatchSize = 2500
)

var logger = log.NewModuleLogger(log.CMDUtils)

func StartNode(stack *node.Node) {
	if err := stack.Start(); err != nil {
		log.Fatalf("Error starting protocol stack: %v", err)
	}
	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		logger.Info("Got interrupt, shutting down...")
		go stack.Stop()
		for i := 10; i > 0; i-- {
			<-sigc
			if i > 1 {
				logger.Info("Already shutting down, interrupt more to panic.", "times", i-1)
			}
		}
	}()
}