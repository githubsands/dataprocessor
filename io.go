package main

import "context"

type IO interface {
	ProcessLogs(func(context.Context), string, []string)
}
