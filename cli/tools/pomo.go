package tools

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	cli "github.com/pressly/cli"
)

var pomoCmd = &cli.Command{
	Name:      "pomo",
	ShortHelp: "A simple pomodoro timer",
	Usage:     "envoy tools pomo [--duration=DURATION]",
	Flags: cli.FlagsFunc(func(f *flag.FlagSet) {
		f.String("duration", "25m", "Duration of the timer (e.g., 25m, 30s, 1h)")
	}),
	Exec: func(ctx context.Context, s *cli.State) error {
		durationStr := cli.GetFlag[string](s, "duration")

		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}

		runPomodoro(s.Stdout, duration)
		return nil
	},
}

func runPomodoro(stdout io.Writer, duration time.Duration) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	remaining := duration
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	clearLine := func() {
		fmt.Fprintf(stdout, "\r%s\r", strings.Repeat(" ", 50))
	}

	printTimer := func() {
		clearLine()
		mins := int(remaining.Seconds()) / 60
		secs := int(remaining.Seconds()) % 60
		fmt.Fprintf(stdout, "\rTime remaining: %02d:%02d", mins, secs)
	}

	fmt.Fprintf(stdout, "Starting pomodoro timer for %v...\n", duration)
	fmt.Fprintf(stdout, "Press Ctrl+C to stop early.\n\n")

	done := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				remaining -= time.Second
				if remaining <= 0 {
					done <- true
					return
				}
				printTimer()
			case <-sigChan:
				ticker.Stop()
				done <- false
				return
			}
		}
	}()

	stopped := <-done

	if stopped {
		fmt.Fprintf(stdout, "\n\nTime's up! Take a break.\n")
	} else {
		mins := int(remaining.Seconds()) / 60
		secs := int(remaining.Seconds()) % 60
		fmt.Fprintf(stdout, "\n\nTimer stopped. Remaining: %02d:%02d\n", mins, secs)
	}
}
