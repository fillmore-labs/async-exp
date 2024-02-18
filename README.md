# Fillmore Labs Async (Experimental)

[![Go Reference](https://pkg.go.dev/badge/fillmore-labs.com/exp/async.svg)](https://pkg.go.dev/fillmore-labs.com/exp/async)
[![Build Status](https://badge.buildkite.com/06fc8f7bdcfc5c380ea0c7c8bb92a7cee8b1676b841f3c65c8.svg)](https://buildkite.com/fillmore-labs/async-exp)
[![GitHub Workflow](https://github.com/fillmore-labs/exp-async/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/fillmore-labs/async-exp/actions/workflows/test.yml)
[![Test Coverage](https://codecov.io/gh/fillmore-labs/async-exp/graph/badge.svg?token=GQUJA8PKJI)](https://codecov.io/gh/fillmore-labs/async-exp)
[![Maintainability](https://api.codeclimate.com/v1/badges/72fe9626fb821fc70251/maintainability)](https://codeclimate.com/github/fillmore-labs/async-exp/maintainability)
[![Go Report Card](https://goreportcard.com/badge/fillmore-labs.com/exp/async)](https://goreportcard.com/report/fillmore-labs.com/exp/async)
[![License](https://img.shields.io/github/license/fillmore-labs/exp-async)](https://www.apache.org/licenses/LICENSE-2.0)

The `async` package provides interfaces and utilities for writing asynchronous code in Go.

## Motivation

This is an experimental package which has a similar API as
[fillmore-labs.com/promise](https://pkg.go.dev/fillmore-labs.com/promise), but is implemented with a structure instead.

## Usage

Assuming you have a synchronous function `func getMyIP(ctx context.Context) (string, error)` returning your external IP
address (see [GetMyIP](#getmyip) for an example).

Now you can do

```go
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	query := func() (string, error) {
		return getMyIP(ctx) // Capture context with timeout
	}
	future := async.NewAsync(query)
```

and elsewhere in your program, even in a different goroutine

```go
	if ip, err := future.Await(ctx); err == nil {
		slog.Info("Found IP", "ip", ip)
	} else {
		slog.Error("Failed to fetch IP", "error", err)
	}
```

decoupling query construction from result processing.

### GetMyIP

Sample code to retrieve your IP address:

```go
const serverURL = "https://httpbin.org/ip"

func getMyIP(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	ipResponse := struct {
		Origin string `json:"origin"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&ipResponse)

	return ipResponse.Origin, err
}
```

## Links

- [Futures and Promises](https://en.wikipedia.org/wiki/Futures_and_promises) in the English Wikipedia
