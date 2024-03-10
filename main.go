package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
)

var (
	apiUrl   = os.Getenv("ASK_URL")
	apiToken = os.Getenv("ASK_TOKEN")
	model    = os.Getenv("ASK_MODEL")
)

func main() {
	if apiUrl == "" || apiToken == "" {
		fmt.Println("ASK_URL and ASK_TOKEN must be set")
		os.Exit(1)
	}
	if model == "" {
		fmt.Println("ASK_MODEL must be set")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	line, _, err := reader.ReadLine()
	if err != nil {
		panic(fmt.Sprint("read from stdin: %w", err))
	}

	fmt.Println("\nFrom AI:\n")
	if err = completions(string(line), model, apiToken, apiUrl, func(r *CompletionsResponse, done bool, err error) {
		if err != nil {
			panic(err)
		}
		if done {
			return
		}
		if r != nil {
			fmt.Print(r.Choices[0].Delta.Content)
		}
	}); err != nil {
		panic(err)
	}
}

func with(f func(r []byte, done bool, err error)) func(resp *http.Response, err error) {
	return func(resp *http.Response, err error) {
		if err != nil {
			f(nil, false, err)
		}
		defer resp.Body.Close()
		reader := bufio.NewReader(resp.Body)

		for {
			r, _, err := reader.ReadLine()
			if err != nil {
				if err.Error() == "EOF" {
					f(nil, true, nil)
					return
				}
				f(nil, true, err)
			}

			f(r, false, nil)
		}
	}
}
