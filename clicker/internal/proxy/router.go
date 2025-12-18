package proxy

import (
	"fmt"
	"sync"

	"github.com/vibium/clicker/internal/bidi"
	"github.com/vibium/clicker/internal/browser"
)

// BrowserSession represents a browser session connected to a client.
type BrowserSession struct {
	LaunchResult *browser.LaunchResult
	BidiConn     *bidi.Connection
	Client       *ClientConn
	mu           sync.Mutex
	closed       bool
	stopChan     chan struct{}
}

// Router manages browser sessions for connected clients.
type Router struct {
	sessions sync.Map // map[uint64]*BrowserSession (client ID -> session)
	headless bool
}

// NewRouter creates a new router.
func NewRouter(headless bool) *Router {
	return &Router{
		headless: headless,
	}
}

// OnClientConnect is called when a new client connects.
// It launches a browser and establishes a BiDi connection.
func (r *Router) OnClientConnect(client *ClientConn) {
	fmt.Printf("[router] Launching browser for client %d...\n", client.ID)

	// Launch browser
	launchResult, err := browser.Launch(browser.LaunchOptions{
		Headless: r.headless,
	})
	if err != nil {
		fmt.Printf("[router] Failed to launch browser for client %d: %v\n", client.ID, err)
		client.Send(fmt.Sprintf(`{"error":{"code":-32000,"message":"Failed to launch browser: %s"}}`, err.Error()))
		client.Close()
		return
	}

	fmt.Printf("[router] Browser launched for client %d, WebSocket: %s\n", client.ID, launchResult.WebSocketURL)

	// Connect to browser BiDi WebSocket
	bidiConn, err := bidi.Connect(launchResult.WebSocketURL)
	if err != nil {
		fmt.Printf("[router] Failed to connect to browser BiDi for client %d: %v\n", client.ID, err)
		launchResult.Close()
		client.Send(fmt.Sprintf(`{"error":{"code":-32000,"message":"Failed to connect to browser: %s"}}`, err.Error()))
		client.Close()
		return
	}

	fmt.Printf("[router] BiDi connection established for client %d\n", client.ID)

	session := &BrowserSession{
		LaunchResult: launchResult,
		BidiConn:     bidiConn,
		Client:       client,
		stopChan:     make(chan struct{}),
	}

	r.sessions.Store(client.ID, session)

	// Start routing messages from browser to client
	go r.routeBrowserToClient(session)
}

// OnClientMessage is called when a message is received from a client.
// It forwards the message to the browser.
func (r *Router) OnClientMessage(client *ClientConn, msg string) {
	sessionVal, ok := r.sessions.Load(client.ID)
	if !ok {
		fmt.Printf("[router] No session for client %d\n", client.ID)
		return
	}

	session := sessionVal.(*BrowserSession)

	session.mu.Lock()
	if session.closed {
		session.mu.Unlock()
		return
	}
	session.mu.Unlock()

	// Forward message to browser
	if err := session.BidiConn.Send(msg); err != nil {
		fmt.Printf("[router] Failed to send to browser for client %d: %v\n", client.ID, err)
	}
}

// OnClientDisconnect is called when a client disconnects.
// It closes the browser session.
func (r *Router) OnClientDisconnect(client *ClientConn) {
	sessionVal, ok := r.sessions.LoadAndDelete(client.ID)
	if !ok {
		return
	}

	session := sessionVal.(*BrowserSession)
	r.closeSession(session)
}

// routeBrowserToClient reads messages from the browser and forwards them to the client.
func (r *Router) routeBrowserToClient(session *BrowserSession) {
	for {
		select {
		case <-session.stopChan:
			return
		default:
		}

		msg, err := session.BidiConn.Receive()
		if err != nil {
			session.mu.Lock()
			closed := session.closed
			session.mu.Unlock()

			if !closed {
				fmt.Printf("[router] Browser connection closed for client %d: %v\n", session.Client.ID, err)
				// Browser died, close the client
				session.Client.Close()
			}
			return
		}

		// Forward message to client
		if err := session.Client.Send(msg); err != nil {
			fmt.Printf("[router] Failed to send to client %d: %v\n", session.Client.ID, err)
			return
		}
	}
}

// closeSession closes a browser session and cleans up resources.
func (r *Router) closeSession(session *BrowserSession) {
	session.mu.Lock()
	if session.closed {
		session.mu.Unlock()
		return
	}
	session.closed = true
	session.mu.Unlock()

	fmt.Printf("[router] Closing browser session for client %d\n", session.Client.ID)

	// Signal the routing goroutine to stop
	close(session.stopChan)

	// Close BiDi connection
	if session.BidiConn != nil {
		session.BidiConn.Close()
	}

	// Close browser
	if session.LaunchResult != nil {
		session.LaunchResult.Close()
	}

	fmt.Printf("[router] Browser session closed for client %d\n", session.Client.ID)
}

// CloseAll closes all browser sessions.
func (r *Router) CloseAll() {
	r.sessions.Range(func(key, value interface{}) bool {
		session := value.(*BrowserSession)
		r.closeSession(session)
		r.sessions.Delete(key)
		return true
	})
}
