package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestEventLoop(t *testing.T) {

	var countEvent int

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startServerCmd := func(ctx context.Context) Msg {
		countEvent++
		fmt.Println("> server starting")
		return "SERVER_STARTED"
	}
	loadConfigCmd := func(ctx context.Context) Msg {
		fmt.Println("> pretend load config")

		s := rand.NewSource(time.Now().UnixNano())
		r := rand.New(s)
		if r.Intn(2) == 0 {
			fmt.Println("> config failed")
			return "CONFIG_LOAD_FAILED"
		}
		fmt.Println("> config ok")
		return "CONFIG_LOAD_OK"
	}
	fileWaitForUpdateCmd := func(ctx context.Context) Msg {
		fmt.Println("> waiting for a file update")
		fmt.Println("> -------------------------")
		return "FILE_UPDATED"
	}
	fileUpdateCmd := func(ctx context.Context) Msg {
		fmt.Println("> pretend receive file update")
		fmt.Println("> pretend stopping server if running")
		return "FILE_UPDATED"
	}

	el, err := NewEventLoop(
		[]LabelledCmd{
			LabelledCmd{"CONFIG_LOAD_OK", startServerCmd},
			LabelledCmd{"CONFIG_LOAD_FAILED", fileWaitForUpdateCmd},
			LabelledCmd{"FILE_UPDATED", loadConfigCmd},
			LabelledCmd{"SERVER_STARTED", fileWaitForUpdateCmd},
		},
		fileUpdateCmd,
		fileWaitForUpdateCmd,
	)
	// update server started to exit on 4th "server restart"
	el.cmdMap["SERVER_STARTED"] = func(ctx context.Context) Msg {
		countEvent++
		if countEvent > 3 {
			cancel()
		}
		fmt.Println("> waiting for a file update")
		fmt.Println("> -------------------------")
		return "FILE_UPDATED"
	}

	if err != nil {
		fmt.Println("new event loop error", err)
		os.Exit(1)
	}
	el.Run(ctx)
}
