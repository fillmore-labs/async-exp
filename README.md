# Async (Experimental)

[![Go Reference](https://pkg.go.dev/badge/fillmore-labs.com/exp/async.svg)](https://pkg.go.dev/fillmore-labs.com/exp/async)
[![Build Status](https://badge.buildkite.com/06fc8f7bdcfc5c380ea0c7c8bb92a7cee8b1676b841f3c65c8.svg)](https://buildkite.com/fillmore-labs/async-exp)
[![Test Coverage](https://codecov.io/gh/fillmore-labs/async-exp/graph/badge.svg?token=GQUJA8PKJI)](https://codecov.io/gh/fillmore-labs/async-exp)
[![Maintainability](https://api.codeclimate.com/v1/badges/72fe9626fb821fc70251/maintainability)](https://codeclimate.com/github/fillmore-labs/async-exp/maintainability)
[![Go Report Card](https://goreportcard.com/badge/fillmore-labs.com/exp/async)](https://goreportcard.com/report/fillmore-labs.com/exp/async)
[![License](https://img.shields.io/github/license/fillmore-labs/exp-async)](https://github.com/fillmore-labs/exp-async/blob/main/LICENSE)

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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	future := async.NewAsyncFuture(func() (string, error) {
		return getMyIP(ctx)
	})
```

and elsewhere in your program, even in a different goroutine

```go
	if ip, err := future.Wait(ctx); err == nil {
		log.Printf("My IP is %s", ip)
	} else {
		log.Printf("Error fetching IP: %v", err)
	}
```

decoupling query construction from result processing.

### GetMyIP

Sample code to retrieve your IP address:

```go
const (
	serverURL = "https://httpbin.org/ip"
	timeout   = 2 * time.Second
)

type IPResponse struct {
	Origin string `json:"origin"`
}

func getMyIP(ctx context.Context) (string, error) {
	resp, err := sendRequest(ctx)
	if err != nil {
		return "", err
	}

	ipResponse, err := decodeResponse(resp)
	if err != nil {
		return "", err
	}

	return ipResponse.Origin, nil
}

func sendRequest(ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	return http.DefaultClient.Do(req)
}

func decodeResponse(response *http.Response) (*IPResponse, error) {
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		return nil, err
	}

	ipResponse := &IPResponse{}
	err = json.Unmarshal(body, ipResponse)
	if err != nil {
		return nil, err
	}

	return ipResponse, nil
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