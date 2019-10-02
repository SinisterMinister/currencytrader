package coinbasepro

import (
	"sync"

	"github.com/go-playground/log"
	ws "github.com/gorilla/websocket"
	"github.com/preichenberger/go-coinbasepro"
)

type websocketHandler struct {
	client *coinbasepro.Client

	mutex         sync.Mutex
	subscriptions Subscriptions
	streams       map[<-chan bool]chan interface{}
	connection    *ws.Conn
}

func newWebSocketHandler(client *coinbasepro.Client) *websocketHandler {
	handler := &websocketHandler{
		client:  client,
		streams: make(map[<-chan bool]chan interface{}),
	}
	go handler.handleConnection()

	return handler
}

func (h *websocketHandler) Subscriptions() Subscriptions {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return h.subscriptions
}

func (h *websocketHandler) Subscribe(req Subscribe) (err error) {
	// Make sure the message type is correct
	req.Type = "subscribe"

	return h.processSub(req)
}

func (h *websocketHandler) Unsubscribe(req Subscribe) (err error) {
	// Make sure the message type is correct
	req.Type = "unsubscribe"

	return h.processSub(req)
}

func (h *websocketHandler) GetStream(stop <-chan bool) (stream <-chan interface{}) {
	// First, create the channel
	rawStream := make(chan interface{})
	stream = rawStream

	// Add the stream to the collection
	h.addStream(stop, rawStream)

	// Watch the stop channel and remove the stream if closed
	go func() {
		select {
		case <-stop:
			// Stop closed. Kill stream
			h.removeStream(stop)
		}
	}()
	return
}

func (h *websocketHandler) processSub(sub Subscribe) (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	err = h.connection.WriteJSON(sub)

	return
}

func (h *websocketHandler) addStream(stop <-chan bool, stream chan interface{}) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Add stream to the map
	h.streams[stop] = stream
}

func (h *websocketHandler) removeStream(stop <-chan bool) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Remove stream
	delete(h.streams, stop)
}

func (h *websocketHandler) handleConnection() {
	// Initialize the connection
	var err error
	err = h.initializeConnection()
	if err != nil {
		log.Error("Could not setup websocket connection", err)
		return
	}

	// Start reading the data
	err = h.readConnection()
	if err != nil {
		if ws.IsUnexpectedCloseError(err, ws.CloseNormalClosure) {
			log.Error("socket closed unexpectedly", err)

			// Restart the handler
			go h.handleConnection()
		}
	}

	return
}

func (h *websocketHandler) initializeConnection() (err error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	conn, _, err := ws.DefaultDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		return
	}

	h.connection = conn

	return
}

func (h *websocketHandler) readConnection() (err error) {
	h.mutex.Lock()
	conn := h.connection
	h.mutex.Unlock()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Send data to streams
		h.sendToStreams(message)
	}
	return
}

func (h *websocketHandler) sendToStreams(data []byte) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Iterate over the streams and send the data
	for _, stream := range h.streams {
		select {
		// Send the data to the stream
		case stream <- data:

		// Skip the stream if it is blocked
		default:
			log.Warn("skipping blocked stream")
		}
	}
}
