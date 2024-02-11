package cmd

import (
	"blue/consts"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "embed"

	"golang.org/x/time/rate"
)

//go:embed play_static/index.html
var indexPage string

//go:embed play_static/loader.js
var loaderJS string

//go:embed play_static/kotlin.js
var kotlinJS string

type EvalCode struct {
	Code string `json:"code"`
}

type EvalCodeResponse struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

func encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

func openbrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func handlePlayCommand(argc int, arguments []string) {
	noExec := true
	timeout := 0
	host := "localhost"
	port := "4242"
	for _, arg := range arguments {
		if strings.HasPrefix(arg, "--timeout=") {
			t, err := strconv.ParseInt(strings.Split(arg, "--timeout=")[1], 10, 64)
			if err != nil {
				consts.ErrorPrinter("`play` timeout parse error: %s\n", arg)
				os.Exit(1)
			}
			if t < 0 {
				consts.ErrorPrinter("`play` timeout must be > 0, got=%d\n", t)
				os.Exit(1)
			}
			timeout = int(t)
		} else if arg == "--yes-exec" {
			noExec = false
		} else if strings.HasPrefix(arg, "--host=") {
			host = strings.Split(arg, "--host=")[1]
		} else if strings.HasPrefix(arg, "--port=") {
			port = strings.Split(arg, "--port=")[1]
		}
	}
	_ = noExec
	_ = timeout
	// Error out if Blue Not found in Path
	blueExeName := "blue"
	if runtime.GOOS == "windows" {
		blueExeName += ".exe"
	}
	if _, err := exec.LookPath(blueExeName); err != nil && !errors.Is(err, exec.ErrDot) {
		consts.ErrorPrinter("`play` error: %s not found in path. error = %s\n", blueExeName, err.Error())
		os.Exit(1)
	}

	// Start Server
	// Get request to serve index
	// Post request to handle spawning blue, running code and returning
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			log.Printf("/ invalid method: %s, expected GET or HEAD", r.Method)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		_, err := fmt.Fprint(w, indexPage)
		if err != nil {
			log.Printf("GET / error: %s", err.Error())
		}
	})
	mux.HandleFunc("/loader.js", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			log.Printf("/loader.js invalid method: %s, expected GET or HEAD", r.Method)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		_, err := fmt.Fprint(w, loaderJS)
		if err != nil {
			log.Printf("GET /loader.js error: %s", err.Error())
		}
	})
	mux.HandleFunc("/kotlin.js", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			log.Printf("/kotlin.js invalid method: %s, expected GET or HEAD", r.Method)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		_, err := fmt.Fprint(w, kotlinJS)
		if err != nil {
			log.Printf("GET /kotlin.js error: %s", err.Error())
		}
	})
	mux.HandleFunc("/eval", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			log.Printf("/eval invalid method: %s, expected POST", r.Method)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		decoded, err := decode[EvalCode](r)
		if err != nil {
			log.Printf("POST /eval error decoding: %s", err.Error())
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		file, err := os.CreateTemp("", "blue-play-")
		if err != nil {
			log.Printf("POST /eval error creating tmp file: %s", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer os.Remove(file.Name())

		if _, err := file.WriteString(decoded.Code); err != nil {
			log.Printf("POST /eval error writing to tmp file %s: %s", file.Name(), err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		var cmdArgs []string
		if runtime.GOOS == "windows" {
			cmdArgs = []string{"cmd", "/c"}
		}
		cmdArgs = append(cmdArgs, "blue", "-e", "--no-exec", file.Name())

		log.Printf("cmdArgs = %s", strings.Join(cmdArgs, " "))
		cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("POST /eval error getting stdout pipe: %s", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer stdout.Close()
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Printf("POST /eval error getting stderr pipe: %s", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer stderr.Close()
		if err := cmd.Start(); err != nil {
			log.Printf("POST /eval error starting cmd `%s`: %s", strings.Join(cmdArgs, " "), err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		stdoutScanner := bufio.NewScanner(stdout)
		var stdoutOut bytes.Buffer
		for stdoutScanner.Scan() {
			text := stdoutScanner.Text()
			stdoutOut.WriteString(text)
			stdoutOut.WriteByte('\n')
		}
		stderrScanner := bufio.NewScanner(stderr)
		var stderrOut bytes.Buffer
		for stderrScanner.Scan() {
			text := stderrScanner.Text()
			stderrOut.WriteString(text)
			stderrOut.WriteByte('\n')
		}
		if err := cmd.Wait(); err != nil {
			log.Printf("POST /eval error waiting cmd `%s`: %s", strings.Join(cmdArgs, " "), err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		log.Printf("stdoutString = %s", stdoutOut.String())
		log.Printf("stderrString = %s", stderrOut.String())
		err = encode(w, r, http.StatusOK, EvalCodeResponse{Stdout: stdoutOut.String(), Stderr: stderrOut.String()})
		if err != nil {
			log.Printf("POST /eval error encodind response: %s", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})

	var limiter = rate.NewLimiter(20, 60)

	limit := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	addr := net.JoinHostPort(host, port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: limit(mux),
	}
	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()
	serverCtx := context.Background()
	serverCtx, serverCancel := signal.NotifyContext(serverCtx, os.Interrupt)
	defer serverCancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-serverCtx.Done()
		if err := httpServer.Shutdown(serverCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()

	// Open Browser here
	openbrowser("http://" + addr)
	wg.Wait()

	// Technically I should be able to use -e 'string' and not need to escape it due to the way go's exec works
	// This may be easier all around to do what I want

	// For each request to get the code, take the 'file' run it through `blue -e <FILE> --no-exec` with timeout context (or if not needed by itself)
	// Capture the output to return from stdout/stderr
	// ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// defer cancel()

	// cmd := exec.CommandContext(ctx, "sleep", "5")
	// if err := cmd.Run(); err != nil { // TODO: Look at our bundle command for capturing output after starting it
	// 	// This will fail after 3 seconds. The 5 second sleep
	// 	// will be interrupted.
	// }
}
