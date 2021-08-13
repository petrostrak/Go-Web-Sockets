package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

var (
	views = jet.NewSet(
		jet.NewOSFileSystemLoader("html"),
		jet.InDevelopmentMode(),
	)
	upgradeConnection = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	wsChan  = make(chan WsPayload)
	clients = make(map[WebSocketConnection]string)
)

type WebSocketConnection struct {
	*websocket.Conn
}

// Defines the response sent back from websocket
type WsJSONResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

type WsPayload struct {
	Action   string              `json:"action"`
	Username string              `json:"username"`
	Message  string              `json:"message"`
	Conn     WebSocketConnection `json:"-"`
}

func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client connected to endpoint")

	var response WsJSONResponse
	response.Message = `<em><small>Connected to server</small></em>`

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = ""

	if err := ws.WriteJSON(response); err != nil {
		log.Println(err)
	}

	go ListenForWs(&conn)
}

func ListenForWs(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error", fmt.Sprintf("%v\n", r))
		}
	}()

	var payload WsPayload

	for {
		if err := conn.ReadJSON(&payload); err != nil {
			// DO NOTHING
		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

func ListenToWsChannel() {
	var respose WsJSONResponse

	for {
		e := <-wsChan

		switch e.Action {
		case "username":
			// get a list of all users and send it back via BroadcastAll
			clients[e.Conn] = e.Username
			users := getUserList()
			respose.Action = "list_users"
			respose.ConnectedUsers = users
			BroadcastToAll(respose)
		case "left":
			// handle the situation where a user leaves the page
			respose.Action = "list_users"
			delete(clients, e.Conn)
			users := getUserList()
			respose.ConnectedUsers = users
			BroadcastToAll(respose)
		}
		// respose.Action = "Got here"
		// respose.Message = fmt.Sprintf("Some message, and action was %s", e.Action)
		// BroadcastToAll(respose)
	}
}

func getUserList() []string {
	var userList []string
	for _, x := range clients {
		userList = append(userList, x)
	}
	sort.Strings(userList)
	return userList
}

func BroadcastToAll(r WsJSONResponse) {
	for client := range clients {
		if err := client.WriteJSON(r); err != nil {
			log.Println("websocket err")
			client.Close()
			delete(clients, client)
		}
	}
}

func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	if err := view.Execute(w, data, nil); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func Home(w http.ResponseWriter, r *http.Request) {
	if err := renderPage(w, "home.jet", nil); err != nil {
		log.Println(err)
	}
}
