package signals

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Default is the set of OS signals intercepted when no explicit signals are
// supplied to Context.
//
//   - SIGINT  (Ctrl-C) is the interactive interrupt sent from a terminal.
//   - SIGTERM is the termination signal sent by orchestrators (Kubernetes,
//     systemd, Docker) to ask a process to shut down gracefully before it is
//     killed with SIGKILL.
//
// SIGTERM is the signal that matters in production; SIGINT is mostly for local
// development. Handling both is the baseline for a well-behaved service.
var Default = []os.Signal{syscall.SIGINT, syscall.SIGTERM}

// Context returns a copy of the parent context that is marked Done when one of
// the given signals arrives, or when the parent context is cancelled —
// whichever happens first. If no signals are supplied, Default is used.
//
// The returned stop function releases the signal handler and must be called
// (typically deferred) to avoid leaking resources. After stop is called, or
// after the first signal is delivered, signal handling is detached: a second
// signal is handled by Go's default disposition and force-terminates the
// process. This gives operators an escape hatch — pressing Ctrl-C twice aborts
// a shutdown that has wedged.
//
//	ctx, stop := signals.Context(context.Background())
//	defer stop()
//
//	<-ctx.Done() // unblocks on SIGINT/SIGTERM
func Context(parent context.Context, sig ...os.Signal) (context.Context, context.CancelFunc) {
	if len(sig) == 0 {
		sig = Default
	}
	return signal.NotifyContext(parent, sig...)
}
