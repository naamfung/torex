package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

// auth is the login of a connection: its keys, else a password, asked for
// once.
type auth struct {
	user, host string

	once     sync.Once
	password string
	err      error
}

// methods returns the ways to log in, in the order to try them: the keys of
// ~/.ssh or the identity file, then a password.
func (a *auth) methods(identity string) []ssh.AuthMethod {
	methods := []ssh.AuthMethod{}
	if signers := loadSigners(identity); len(signers) > 0 {
		methods = append(methods, ssh.PublicKeys(signers...))
	}
	return append(methods,
		ssh.PasswordCallback(a.secret),
		ssh.KeyboardInteractive(func(name, instruction string, questions []string, echos []bool) ([]string, error) {
			answers := make([]string, len(questions))
			for i, question := range questions {
				if echos[i] {
					answers[i] = readLine(question)
					continue
				}
				secret, err := a.secret()
				if err != nil {
					return nil, err
				}
				answers[i] = secret
			}
			return answers, nil
		}),
	)
}

// secret returns the password of the login, asking for it once.
func (a *auth) secret() (string, error) {
	a.once.Do(func() {
		a.password, a.err = readSecret(fmt.Sprintf("%s@%s's password: ", a.user, a.host))
	})
	return a.password, a.err
}

// sshDir returns ~/.ssh, where the keys and known_hosts are.
func sshDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	return filepath.Join(home, ".ssh")
}

// loadSigners returns the keys to log in with: the identity file, or the
// usual names of ~/.ssh. A key with a passphrase asks for it.
func loadSigners(identity string) []ssh.Signer {
	paths := []string{identity}
	if identity == "" {
		for _, name := range []string{"id_ed25519", "id_ecdsa", "id_rsa"} {
			paths = append(paths, filepath.Join(sshDir(), name))
		}
	}
	var signers []ssh.Signer
	for _, path := range paths {
		if path == "" {
			continue
		}
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		signer, err := ssh.ParsePrivateKey(b)
		if err != nil {
			var missing *ssh.PassphraseMissingError
			if !errors.As(err, &missing) {
				fmt.Fprintf(os.Stderr, "ssh: %s: %v\n", path, err)
				continue
			}
			passphrase, perr := readSecret(fmt.Sprintf("Enter passphrase for key '%s': ", path))
			if perr != nil {
				continue
			}
			signer, err = ssh.ParsePrivateKeyWithPassphrase(b, []byte(passphrase))
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "ssh: %s: %v\n", path, err)
			continue
		}
		signers = append(signers, signer)
	}
	return signers
}

// hostKeyCallback checks the host against ~/.ssh/known_hosts, as the ssh of
// OpenSSH does: a host it does not know is asked about; one whose key changed
// is refused.
func hostKeyCallback() ssh.HostKeyCallback {
	path := filepath.Join(sshDir(), "known_hosts")
	check, err := knownhosts.New(path)
	if err != nil {
		check = nil
	}
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		if check != nil {
			err := check(hostname, remote, key)
			if err == nil {
				return nil
			}
			var keyErr *knownhosts.KeyError
			if !errors.As(err, &keyErr) {
				return err
			}
			if len(keyErr.Want) > 0 {
				return fmt.Errorf("host key verification failed: %s has a different key", hostname)
			}
		}
		fmt.Fprintf(os.Stderr, "The authenticity of host '%s' can't be established.\n%s key fingerprint is %s.\nAre you sure you want to continue connecting (yes/no)? ",
			hostname, key.Type(), ssh.FingerprintSHA256(key))
		switch strings.ToLower(strings.TrimSpace(readLine(""))) {
		case "yes", "y":
		default:
			return errors.New("host key verification failed")
		}
		if err := appendKnownHost(path, hostname, key); err != nil {
			fmt.Fprintf(os.Stderr, "ssh: %s: %v\n", path, err)
			return nil
		}
		fmt.Fprintf(os.Stderr, "Warning: Permanently added '%s' (%s) to the list of known hosts.\n", knownhosts.Normalize(hostname), key.Type())
		return nil
	}
}

// appendKnownHost adds the host and its key to the known_hosts file.
func appendKnownHost(path, hostname string, key ssh.PublicKey) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, knownhosts.Line([]string{knownhosts.Normalize(hostname)}, key))
	return err
}

// readSecret asks for a secret on the terminal, without echoing it.
func readSecret(prompt string) (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", errors.New("no terminal to ask for a password")
	}
	fmt.Fprint(os.Stderr, prompt)
	b, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	return string(b), err
}

// readLine writes prompt and returns the line typed after it.
func readLine(prompt string) string {
	if prompt != "" {
		fmt.Fprint(os.Stderr, prompt+" ")
	}
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return ""
	}
	return strings.TrimRight(line, "\r\n")
}
