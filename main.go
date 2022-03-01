package main

import (
	"bufio"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func usage() {
	fmt.Fprintln(os.Stderr, `usage: fzfcache [-h] <SHELL COMMAND...>

Caches the input from the shell command and/or prints the cached results
This is typically piped into fzf, to decrease the time till interactive`)
}

func parseFlags() []string {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		fmt.Fprintln(os.Stderr, `Error: Not enough arguments -- needs a command`)
		os.Exit(1)
	}
	if len(args) == 1 && (args[0] == "-h" || args[0] == "-help" || args[0] == "--help") {
		usage()
		os.Exit(0)
	}
	return args
}

func getCacheDir() (cachedir string) {
	cachedir = os.Getenv("FZFCACHE_DIR")
	if cachedir == "" {
		cdir := os.Getenv("XDG_CACHE_HOME")
		if cdir == "" {
			cdir = path.Join(os.Getenv("HOME"), ".cache")
		}
		cachedir = path.Join(cdir, "fzfcache")
	}
	err := os.MkdirAll(cachedir, 0700)
	if err != nil {
		log.Fatalf("Could not create cache directory: %s\n", err)
	}
	return
}

func copyFile(in, out string) (int64, error) {
	i, e := os.Open(in)
	if e != nil {
		return 0, e
	}
	defer i.Close()

	o, e := os.Create(out)
	if e != nil {
		return 0, e
	}
	defer o.Close()
	return o.ReadFrom(i)
}

func commandCacheFile(command string) string {
	h := sha1.New()
	io.WriteString(h, command)
	sum := fmt.Sprintf("%x", h.Sum(nil))
	return sum
}

func cachedCommand(command string) (string, error) {

	commandHash := commandCacheFile(command)
	cacheFile := path.Join(getCacheDir(), commandHash)

	// whether or not something has already been printed
	lines := make(map[string]bool)

	// print from cache file, if it exists
	if _, err := os.Stat(cacheFile); err == nil {
		cf, err := os.Open(cacheFile)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Error opening cachefile: %s\n", err))
		}
		reader := bufio.NewScanner(cf)
		for reader.Scan() {
			txt := reader.Text()
			// dont need to check for duplicates since we did when we wrote
			fmt.Println(txt)
			lines[txt] = true
		}
		if err := reader.Err(); err != nil {
			return "", errors.New(fmt.Sprintf("Error reading from cachefile: %s\n", err))
		}
		cf.Close()
	}

	// create process to run passed command
	shell := os.Getenv("SHELL")
	if len(shell) == 0 {
		shell = "sh"
	}
	cmd := exec.Command(shell, "-c", command)

	// get STDOUT/STDERR pipes
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error getting STDOUT for command: %s\n", err))
	}

	errReader, err := cmd.StderrPipe()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error getting STDERR for command: %s\n", err))
	}

	// create tempfile to hold output of current command
	tf, err := os.CreateTemp("", "fzfcache-")
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error creating temporary output file: %s\n", err))
	}

	defer tf.Close()
	defer func() {
		_, err := copyFile(tf.Name(), cacheFile)
		if err != nil {
			log.Fatalf("Error copying tempfile to fzfcache: %s", err)
		}
	}()

	// loop over cmd STDOUT
	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			txt := scanner.Text()
			// if this line hasn't already been printed
			if !lines[txt] {
				// save it, and print
				lines[txt] = true
				fmt.Println(txt)
			}
			// append to tempfile, so we can overwrite previous results
			if _, err := tf.WriteString(txt + "\n"); err != nil {
				log.Fatal(err)
			}
			// print if not already in in-mem set which keeps track of printed lines
		}
	}()

	// write errors to stderr
	errScanner := bufio.NewScanner(errReader)
	go func() {
		for errScanner.Scan() {
			fmt.Fprintln(os.Stderr, errScanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error starting command: %s\n", err))
	}

	err = cmd.Wait()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error waiting for command: %s\n", err))
	}
	return tf.Name(), nil
}

func fzfcache() error {
	shellCmd := strings.Join(parseFlags(), " ")
	tempFile, err := cachedCommand(shellCmd)
	if err != nil {
		return err
	}
	err = os.Remove(tempFile)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	if err := fzfcache(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
}
