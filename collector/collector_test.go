package main

import (
	"github.com/timeredbull/tsuru/api/app"
	"github.com/timeredbull/tsuru/api/unit"
	"io/ioutil"
	. "launchpad.net/gocheck"
	"os"
	"path/filepath"
)

func getOutput() *output {
	return &output{
		Services: map[string]Service{
			"umaappqq": Service{
				Units: map[string]unit.Unit{
					"umaappqq/0": unit.Unit{
						AgentState: "started",
						Machine:    1,
					},
				},
			},
		},
		Machines: map[int]interface{}{
			0: map[interface{}]interface{}{
				"dns-name":       "192.168.0.10",
				"instance-id":    "i-00000zz6",
				"instance-state": "running",
				"agent-state":    "running",
			},
			1: map[interface{}]interface{}{
				"dns-name":       "192.168.0.11",
				"instance-id":    "i-00000zz7",
				"instance-state": "running",
				"agent-state":    "running",
			},
		},
	}
}

func getApp(c *C) *app.App {
	a := &app.App{Name: "umaappqq", State: "STOPPED"}
	err := a.Create()
	c.Assert(err, IsNil)
	return a
}

func (s *S) TestCollectorUpdate(c *C) {
	a := getApp(c)
	var collector Collector
	out := getOutput()
	collector.Update(out)

	err := a.Get()
	c.Assert(err, IsNil)
	c.Assert(a.State, Equals, "STARTED")
	c.Assert(a.Units[0].Ip, Equals, "192.168.0.11")
	c.Assert(a.Units[0].Machine, Equals, 1)
	c.Assert(a.Units[0].InstanceState, Equals, "running")
	c.Assert(a.Units[0].AgentState, Equals, "running")
	c.Assert(a.Units[0].InstanceId, Equals, "i-00000zz7")

	a.Destroy()
}

func (s *S) TestCollectorUpdateWithMultipleUnits(c *C) {
	a := getApp(c)
	out := getOutput()
	u := unit.Unit{AgentState: "started", Machine: 2}
	out.Services["umaappqq"].Units["umaappqq/1"] = u
	out.Machines[2] = map[interface{}]interface{}{
		"dns-name":       "192.168.0.12",
		"instance-id":    "i-00000zz8",
		"instance-state": "running",
		"agent-state":    "running",
	}
	var collector Collector
	collector.Update(out)
	err := a.Get()
	c.Assert(err, IsNil)
	c.Assert(len(a.Units), Equals, 2)
	for _, u = range a.Units {
		if u.Machine == 2 {
			break
		}
	}
	c.Assert(u.Ip, Equals, "192.168.0.12")
	c.Assert(u.InstanceState, Equals, "running")
	c.Assert(u.AgentState, Equals, "running")
}

func (s *S) TestCollectorUpdateWithDownMachine(c *C) {
	a := app.App{Name: "barduscoapp", State: "STOPPED"}
	err := a.Create()
	c.Assert(err, IsNil)
	file, _ := os.Open(filepath.Join("testdata", "broken-output.yaml"))
	jujuOutput, _ := ioutil.ReadAll(file)
	file.Close()
	var collector Collector
	out := collector.Parse(jujuOutput)
	collector.Update(out)
	err = a.Get()
	c.Assert(err, IsNil)
	c.Assert(a.State, Equals, "PENDING")
}

func (s *S) TestCollectorUpdateTwice(c *C) {
	a := getApp(c)
	var collector Collector
	defer a.Destroy()
	out := getOutput()
	collector.Update(out)
	err := a.Get()
	c.Assert(err, IsNil)
	c.Assert(a.State, Equals, "STARTED")
	c.Assert(a.Units[0].Ip, Equals, "192.168.0.11")
	c.Assert(a.Units[0].Machine, Equals, 1)
	c.Assert(a.Units[0].InstanceState, Equals, "running")
	c.Assert(a.Units[0].AgentState, Equals, "running")
	collector.Update(out)
	err = a.Get()
	c.Assert(len(a.Units), Equals, 1)
}

func (s *S) TestCollectorParser(c *C) {
	var collector Collector
	file, _ := os.Open(filepath.Join("testdata", "output.yaml"))
	jujuOutput, _ := ioutil.ReadAll(file)
	file.Close()
	expected := getOutput()
	c.Assert(collector.Parse(jujuOutput), DeepEquals, expected)
}

func (s *S) TestAppStatusAgentPending(c *C) {
	u := unit.Unit{AgentState: "not-started", InstanceState: "running"}
	st := appState(&u)
	c.Assert(st, Equals, "STOPPED")
}

func (s *S) TestAppStatusInstanceStateError(c *C) {
	u := unit.Unit{AgentState: "not-started", InstanceState: "error"}
	st := appState(&u)
	c.Assert(st, Equals, "ERROR")
}

func (s *S) TestAppStatusInstanceStatePending(c *C) {
	u := unit.Unit{AgentState: "pending", InstanceState: ""}
	st := appState(&u)
	c.Assert(st, Equals, "PENDING")
}

func (s *S) TestAppStatusAgentAndInstanceRunning(c *C) {
	u := unit.Unit{AgentState: "running", InstanceState: "running"}
	st := appState(&u)
	c.Assert(st, Equals, "STARTED")
}

func (s *S) TestAppStatusInstancePending(c *C) {
	u := unit.Unit{AgentState: "not-started", InstanceState: "pending"}
	st := appState(&u)
	c.Assert(st, Equals, "PENDING")
}
