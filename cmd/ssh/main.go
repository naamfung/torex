// Command ssh is the ssh client of torex's Windows runtime: a pane runs
// WinKit's POSIX shell, which has no ssh of its own, so the app carries this
// one, built from Go and embedded beside the shell.
//
//	ssh [-p port] [-l user] [-i key] [-v] [user@]host [command...]
//
// It logs in with the keys in ~/.ssh, else a password, checks the host
// against ~/.ssh/known_hosts, and runs a terminal session of the size of the
// pane.
package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func main() { os.Exit(run()) }

func run() int {
	fs := flag.NewFlagSet("ssh", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	port := fs.String("p", "22", "port to connect to on the remote host")
	login := fs.String("l", "", "user to log in as")
	identity := fs.String("i", "", "identity file")
	verbose := fs.Bool("v", false, "print the connection's progress")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: ssh [-p port] [-l user] [-i key] [-v] [user@]host [command...]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		return 2
	}
	if fs.NArg() == 0 {
		fs.Usage()
		return 2
	}
	t, err := parseTarget(fs.Arg(0), *login, *port)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ssh:", err)
		return 2
	}
	client, err := dial(t, *identity, *verbose)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ssh:", err)
		return 255
	}
	defer client.Close()
	sess, err := client.NewSession()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ssh:", err)
		return 255
	}
	defer sess.Close()
	if command := strings.Join(fs.Args()[1:], " "); command != "" {
		return runCommand(sess, command)
	}
	return runShell(sess, *verbose)
}

// target is the host to connect to, as the command line names it.
type target struct {
	user, host, port string
}

// parseTarget reads [user@]host, with the user and the port of the flags as
// the defaults.
func parseTarget(arg, user, port string) (target, error) {
	if i := strings.LastIndex(arg, "@"); i >= 0 {
		user, arg = arg[:i], arg[i+1:]
	}
	if arg == "" {
		return target{}, errors.New("no host to connect to")
	}
	if user == "" {
		user = os.Getenv("USERNAME")
		if user == "" {
			user = os.Getenv("USER")
		}
	}
	return target{user: user, host: arg, port: port}, nil
}

// dial connects to the target and authenticates with it.
func dial(t target, identity string, verbose bool) (*ssh.Client, error) {
	addr := net.JoinHostPort(t.host, t.port)
	a := &auth{user: t.user, host: t.host}
	config := &ssh.ClientConfig{
		User:            t.user,
		Auth:            a.methods(identity),
		HostKeyCallback: hostKeyCallback(),
		Timeout:         15 * time.Second,
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "connecting to %s as %s\n", addr, t.user)
	}
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		// The handshake's errors already name ssh.
		return nil, errors.New(strings.TrimPrefix(err.Error(), "ssh: "))
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "connected, %s\n", client.ServerVersion())
	}
	return client, nil
}

// runShell runs an interactive shell in a terminal of the pane's size.
func runShell(sess *ssh.Session, verbose bool) int {
	width, height := termSize()
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := sess.RequestPty("xterm-256color", height, width, modes); err != nil {
		fmt.Fprintln(os.Stderr, "ssh:", err)
		return 255
	}
	sess.Stdin, sess.Stdout, sess.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := sess.Shell(); err != nil {
		fmt.Fprintln(os.Stderr, "ssh:", err)
		return 255
	}
	if verbose {
		fmt.Fprintln(os.Stderr, "shell started")
	}
	restore := rawMode()
	defer restore()

	// Windows has no SIGWINCH: ask the pane for its size as the shell runs,
	// as a remote vim redraws to it.
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-time.After(time.Second):
			}
			if w, h := termSize(); w != width || h != height {
				width, height = w, h
				sess.WindowChange(h, w)
			}
		}
	}()

	err := sess.Wait()
	restore()
	return exitStatus(err)
}

// runCommand runs a command on the remote host, without a terminal.
func runCommand(sess *ssh.Session, command string) int {
	sess.Stdin, sess.Stdout, sess.Stderr = os.Stdin, os.Stdout, os.Stderr
	return exitStatus(sess.Run(command))
}

// exitStatus returns the status the remote command exited with.
func exitStatus(err error) int {
	if err == nil {
		return 0
	}
	var exit *ssh.ExitError
	if errors.As(err, &exit) {
		return exit.ExitStatus()
	}
	fmt.Fprintln(os.Stderr, "ssh:", err)
	return 255
}

// termSize returns the size of the pane, as the terminal's or, without one,
// of eighty columns by twenty-four rows.
func termSize() (width, height int) {
	for _, fd := range []int{int(os.Stdout.Fd()), int(os.Stdin.Fd())} {
		if w, h, err := term.GetSize(fd); err == nil && w > 0 && h > 0 {
			return w, h
		}
	}
	return 80, 24
}

// rawMode puts the terminal into raw mode, so that keys reach the remote
// shell as they are, and returns what undoes it.
func rawMode() func() {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return func() {}
	}
	state, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ssh:", err)
		return func() {}
	}
	return func() { term.Restore(fd, state) }
}
