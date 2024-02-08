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

Futures and promises are constructs used for asynchronous and concurrent programming, allowing developers to work with
values that may not be immediately available and can be evaluated in a different execution context.

Go is known for its built-in concurrency features like goroutines and channels.
The select statement further allows for efficient multiplexing and synchronization of multiple channels, thereby
enabling developers to coordinate and orchestrate asynchronous operations effectively.
Additionally, the context package offers a standardized way to manage cancellation, deadlines, and timeouts within
concurrent and asynchronous code.

On the other hand, Go's error handling mechanism, based on explicit error values returned from functions, provides a
clear and concise way to handle errors.

The purpose of this package is to provide a thin layer over channels which simplifies the integration of concurrent
code while providing a cohesive strategy for handling asynchronous errors.
By adhering to Go's standard conventions for asynchronous and concurrent code, as well as error propagation, this
package aims to enhance developer productivity and code reliability in scenarios requiring asynchronous operations.

## Usage

Assuming you have a synchronous function `func getMyIP(ctx context.Context) (string, error)` returning your external IP
address (see [GetMyIP](#getmyip) for an example).

Now you can do

```go
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	future := async.NewFutureAsync(func() (string, error) {
		return getMyIP(ctx)
	})
```

and elsewhere in your program, even in a different goroutine

```go
	if ip, err := future.Wait(ctx); err == nil {
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

	ipResponse := &struct {
		Origin string `json:"origin"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(ipResponse); err != nil {
		return "", err
	}

	return ipResponse.Origin, nil
}
```

## Concurrency Correctness

When utilizing plain Go channels for concurrency, reasoning over the correctness of concurrent code becomes simpler
compared to some other implementations of futures and promises.
Channels provide a clear and explicit means of communication and synchronization between concurrent goroutines, making
it easier to understand and reason about the flow of data and control in a concurrent program.

Therefore, this library provides a straightforward and idiomatic approach to achieving concurrency correctness.

## Links

- [Futures and Promises](https://en.wikipedia.org/wiki/Futures_and_promises) in the English Wikipedia
