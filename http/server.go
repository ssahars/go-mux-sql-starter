package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Server encapsulates all HTTP exposed functionality of app
type Server interface {
	Run()
	IsReady() bool
	Port() (int, error)
	Shutdown() error
}

type server struct {
	port        int
	routes      []Routes
	middlewares []mux.MiddlewareFunc

	listener   net.Listener
	httpServer *http.Server
}

// Run initializes and starts the server.
// Once started, server listens for requests till app termination
func (server *server) Run() {

	// init router
	router := mux.NewRouter()

	// init and register all routes
	routesAgent := &routesAgent{router}
	for _, r := range server.routes {
		r.Register(routesAgent)
	}

	// register all middlewares
	if len(server.middlewares) > 0 {
		router.Use(server.middlewares...)
	}

	// Create a listener for port
	var err error
	server.listener, err = net.Listen("tcp", fmt.Sprintf(":%v", server.port))
	if err != nil {
		panic(err)
	}

	// Create server
	httpServer := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler: handlers.CORS(
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{"POST", "GET", "OPTIONS", "PUT", "DELETE"}),
			handlers.AllowedHeaders([]string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Access-Control-Request-Headers", "Access-Control-Request-Method", "Connection", "Host", "Origin", "User-Agent", "Referer", "Cache-Control", "X-header"}),
		)(router),
	}

	// listen for requests till app termination
	server.httpServer = httpServer
	server.httpServer.Serve(server.listener)
}

func (server *server) IsReady() bool {
	return server.httpServer != nil
}

func (server *server) Port() (int, error) {

	if !server.IsReady() {
		return -1, fmt.Errorf("Server is not running")
	}

	return server.listener.Addr().(*net.TCPAddr).Port, nil
}

func (server *server) Shutdown() error {

	if !server.IsReady() {
		return fmt.Errorf("Server is not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := server.httpServer.Shutdown(ctx)
	server.httpServer = nil
	return err
}
