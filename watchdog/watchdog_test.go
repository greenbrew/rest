// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Roberto Mier Escandon <rmescandon@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package watchdog

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	check "gopkg.in/check.v1"

	"github.com/greenbrew/rest/system"
)

type mockPinger struct {
	counter int
}

func newMockPinger(count int) *mockPinger {
	return &mockPinger{counter: count}
}

// will return true mockPingerCount times and then false
func (m *mockPinger) Ping() bool {
	m.counter = m.counter - 1
	return m.counter >= 0
}

func Test(t *testing.T) { check.TestingT(t) }

type WatchdogSuite struct {
	socketFolder string
	ln           *net.UnixConn
}

var _ = check.Suite(&WatchdogSuite{})

func (s *WatchdogSuite) SetUpSuite(c *check.C) {
	var err error
	s.socketFolder, err = ioutil.TempDir("", "notify-socket-")
	c.Assert(err, check.IsNil)

	socketPath := s.socketFolder + "/notify-socket.sock"

	os.Setenv("NOTIFY_SOCKET", socketPath)

	socketAddr := net.UnixAddr{
		Name: socketPath,
		Net:  "unixgram",
	}

	s.ln, err = net.ListenUnixgram("unixgram", &socketAddr)
	c.Assert(err, check.IsNil)
}

func (s *WatchdogSuite) TearDownSuite(c *check.C) {
	if system.PathExists(s.socketFolder) {
		os.RemoveAll(s.socketFolder)
	}
}

func (s *WatchdogSuite) TestWatchdogNoPing(c *check.C) {
	// Start pinger with 0 positive responses
	w := New(10*time.Millisecond, newMockPinger(0))

	doneListen := make(chan struct{})
	doneRun := make(chan struct{})
	preparing := make(chan struct{})

	go func() {
		defer close(doneListen)
		<-preparing

		expectFromSocket("READY=1", s.ln, c)
	}()

	go func() {
		defer close(doneRun)
		close(preparing)

		err := w.Run(context.Background())
		c.Assert(err, check.IsNil)
	}()

	<-doneListen
	<-doneRun
}

func (s *WatchdogSuite) TestWatchdogDeterministicPing(c *check.C) {
	pingsCount := 5
	w := New(10*time.Millisecond, newMockPinger(pingsCount))

	alive := make(chan struct{})
	doneListen := make(chan struct{})
	doneRun := make(chan struct{})
	preparing := make(chan struct{})

	go func() {
		defer close(doneListen)
		<-preparing

		expectFromSocket("READY=1", s.ln, c)
		for i := 0; i < pingsCount; i++ {
			expectFromSocket("WATCHDOG=1", s.ln, c)
		}

		close(alive)
		expectFromSocket("STOPPING=1", s.ln, c)
	}()

	go func() {
		defer close(doneRun)
		close(preparing)

		err := w.Run(context.Background())
		c.Assert(err, check.IsNil)
	}()

	<-alive
	w.Stop()
	<-doneListen
	<-doneRun
}

func (s *WatchdogSuite) TestWatchdogStoppedWhilePinging(c *check.C) {
	// enoung pings to receive a stopping signal
	w := New(10*time.Millisecond, newMockPinger(100))

	doneListen := make(chan struct{})
	doneRun := make(chan struct{})
	preparing := make(chan struct{})

	go func() {
		defer close(doneListen)
		<-preparing

		expectFromSocket("READY=1", s.ln, c)
		for {
			msg := readFromSocket(s.ln, c)
			switch msg {
			case "WATCHDOG=1":
				continue
			case "STOPPING=1":
				return
			default:
				c.Fail()
			}
		}
	}()

	go func() {
		defer close(doneRun)
		close(preparing)

		err := w.Run(context.Background())
		c.Assert(err, check.IsNil)
	}()

	// wait a bit before stopping
	time.Sleep(100 * time.Millisecond)

	w.Stop()
	<-doneListen
	<-doneRun
}

func readFromSocket(ln *net.UnixConn, c *check.C) string {
	msgChan := make(chan string)
	var msg string

	go func() {
		buff := make([]byte, 1024)

		ns, _, err := ln.ReadFromUnix(buff)
		c.Assert(err, check.IsNil)

		msgChan <- string(buff[0:ns])
	}()

	select {
	case msg = <-msgChan:
	case <-time.After(5 * time.Second):
		c.Fail()
	}
	return msg
}

func expectFromSocket(expected string, ln *net.UnixConn, c *check.C) {
	data := readFromSocket(ln, c)
	c.Assert(data, check.Equals, expected)
}
