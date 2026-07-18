package main

import "context"

type HTTPServer struct {
	app app.Application
}

type Application