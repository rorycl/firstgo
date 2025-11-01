package main

import (
	"context"
	"errors"
	"fmt"
)

// Msg is the result of an IO operation.
type Msg string

// A Cmd is a function that performs I/O and returns a Msg.
type Cmd func(context.Context) Msg

// A LabelledCmd is a command with a label used for command
// dispatching.
type LabelledCmd struct {
	label Msg
	cmd   Cmd
}

// EventLoop describes an event loop to execute a FSM determined by a
// set of labelled commands and a start and default command.
type EventLoop struct {
	defaultCmd Cmd
	startCmd   Cmd
	cmdMap     map[Msg]Cmd
	msgChan    chan Msg
}

// NewEventLoop registers a new eventloop with a set of labelled
// commands, a command to start the event loop process, and a default
// command to run when a message does not match any label.
//
// Note that no checking of the validity of the FSM is made on
// registration.
func NewEventLoop(cmds []LabelledCmd, startCmd, defaultCmd Cmd) (*EventLoop, error) {
	if len(cmds) < 1 {
		return nil, errors.New("no cmds provided to NewEventLoop")
	}

	e := &EventLoop{
		cmdMap:     map[Msg]Cmd{},
		startCmd:   startCmd,
		defaultCmd: defaultCmd,
		msgChan:    make(chan Msg),
	}
	for _, lc := range cmds {
		if _, found := e.cmdMap[lc.label]; found {
			return nil, fmt.Errorf("label %q already registered", lc.label)
		}
		e.cmdMap[lc.label] = lc.cmd
	}
	// consider checking fsm labels -> msgs but msg not extractable from
	// cmd.
	return e, nil
}

// Update does no I/O; just decides what to do next.
// It takes a message and returns the next command to run.
func (e *EventLoop) Update(msg Msg) Cmd {
	if cmd, found := e.cmdMap[msg]; found {
		return cmd
	}
	return e.defaultCmd
}

// Run execute commands and pipes the result back to msgChan.
func (e *EventLoop) Run(ctx context.Context) {

	doCmd := func(cmd Cmd) {
		if cmd == nil {
			panic("received nil command")
		}
		go func() {
			select {
			case <-ctx.Done():
				e.Stop()
			case e.msgChan <- cmd(ctx):
			}
		}()
	}
	// Start initial process.
	doCmd(e.startCmd)

	// The main event loop only ever receives from the channel.
	for msg := range e.msgChan {
		cmd := e.Update(msg)
		doCmd(cmd)
	}
}

// Stop stops the event loop.
func (e *EventLoop) Stop() {
	close(e.msgChan)
}
