package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/madstone-tech/loko/internal/adapters/devserver"
	"github.com/madstone-tech/loko/internal/adapters/hclsource"
	"github.com/madstone-tech/loko/internal/adapters/watch"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// ServeOptions carries the parsed flags.
type ServeOptions struct {
	Root string
	// Host is the interface to bind; "" means 127.0.0.1.
	Host   string
	Port   int
	Stdout io.Writer
	Stderr io.Writer
	// Ready, when set, receives the bound address (tests use port 0).
	Ready func(net.Addr)
}

// runServeWith serves the site until SIGINT/SIGTERM or ctx ends. A compile
// error is shown in the browser, not fatal; failing to bind the port is.
func runServeWith(ctx context.Context, opts ServeOptions) (int, error) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	server := devserver.New()
	listening := make(chan error, 1)
	go func() {
		listening <- server.ListenAndServe(ctx, opts.Host, opts.Port, func(a net.Addr) {
			_, _ = fmt.Fprintf(opts.Stdout, "serving http://%s (ctrl-c to stop)\n", a) // best effort: the server runs regardless
			warnIfExposed(opts.Stderr, a)
			if opts.Ready != nil {
				opts.Ready(a)
			}
		})
	}()
	serving := make(chan error, 1)
	go func() { serving <- usecases.Serve(ctx, newServeDeps(server, opts), newServeRequest(opts)) }()

	select {
	case err := <-listening:
		cancel()
		<-serving
		if err != nil {
			return usecases.ExitErrors, err
		}
	case err := <-serving:
		cancel()
		<-listening
		if err != nil {
			return usecases.ExitErrors, err
		}
	}
	return usecases.ExitSuccess, nil
}

func newServeRequest(opts ServeOptions) usecases.ServeRequest {
	return usecases.ServeRequest{Root: opts.Root, BuildVersion: buildVersion(), OutDir: filepath.Join(opts.Root, "dist")}
}

// describeDiagnostics renders diagnostics for the browser's error page and
// echoes them to stderr, re-reading sources so snippets are current.
func describeDiagnostics(opts ServeOptions) func(usecases.Diagnostics) string {
	return func(d usecases.Diagnostics) string {
		renderer, err := hclsource.NewRendererForRoot(opts.Root, false)
		if err != nil {
			return err.Error()
		}
		var b bytes.Buffer
		if _, _, err := renderer.Write(&b, d); err != nil {
			return err.Error()
		}
		_, _ = io.WriteString(opts.Stderr, b.String()) // best effort: the browser shows it too
		return b.String()
	}
}

// newServeDeps wires the ports `loko serve` adds to the build's.
func newServeDeps(server *devserver.Server, opts ServeOptions) usecases.ServeDeps {
	return usecases.ServeDeps{
		Build:    newBuildDeps(),
		Watcher:  watch.New(watch.Interval),
		Preview:  server,
		Describe: describeDiagnostics(opts),
	}
}

// warnIfExposed says so when the site is reachable beyond this machine:
// --host is an explicit choice, but an easy one to leave in a script.
func warnIfExposed(w io.Writer, a net.Addr) {
	if tcp, ok := a.(*net.TCPAddr); ok && !tcp.IP.IsLoopback() {
		_, _ = fmt.Fprintf(w, "warning: listening on %s, not loopback; anyone who can reach it can read the site\n", a) // best effort, like the line above
	}
}
