package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var (
	seedSpread    = 100
	seedCellCount = 100
	canvasX       = 256 // set to -1 for infinite (limited by int64) point coordinates
	canvasY       = 256 // set to -1 for infinite (limited by int64) point coordinates

	events = make(chan event, 256)

	addr string
)

type event func(*cells)

type cells struct {
	current map[point]bool
	next    map[point]bool
}

type point struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

type websocketClient struct {
	pointsOut chan []point
	pointsIn  chan []point
	conn      *websocket.Conn
	connected bool
}

func (p *point) UnmarshalJSON(b []byte) error {
	proxy := struct {
		X json.Number
		Y json.Number
	}{}

	err := json.Unmarshal(b, &proxy)
	if err != nil {
		return err
	}

	p.X, err = proxy.X.Int64()
	if err != nil {
		return err
	}

	p.Y, err = proxy.Y.Int64()
	if err != nil {
		return err
	}

	return nil
}

func init() {
	flag.StringVar(&addr, "address", ":8080", "Listening address for the server")
	flag.Parse()
}

func main() {
	go runHTTPServer()
	gameLoop()
}

func gameLoop() {
	c := &cells{
		map[point]bool{},
		map[point]bool{},
	}

	for {
		select {
		case fn := <-events:
			fn(c)
		case <-time.After(50 * time.Millisecond):
			tick(c)
		}
	}
}

func tick(c *cells) {
	for p := range c.current {
		if _, ok := c.next[p]; !ok {
			if willLive(p, c) {
				c.next[p] = true
			}
		}

		for _, neighbor := range p.neighbors() {
			if _, ok := c.next[neighbor]; !ok {
				if willLive(neighbor, c) {
					c.next[neighbor] = true
				}
			}
		}
	}

	c.current = c.next
	c.next = map[point]bool{}
}

func seedCells(n, spread int, c *cells) event {
	return event(func(c *cells) {
		rand.Seed(time.Now().UnixNano())
		for i := 0; i < n; i++ {
			p := point{
				int64(rand.Intn(spread)),
				int64(rand.Intn(spread)),
			}
			c.current[p] = true
		}
	})
}

func addCells(pts []point) event {
	return event(func(c *cells) {
		for _, p := range pts {
			c.current[p] = true
		}
	})
}

func willLive(p point, c *cells) bool {
	var neighborsAlive int
	isAlive := c.current[p]

	// fmt.Println(p.X, p.Y)
	for _, neighbor := range p.neighbors() {
		if _, ok := c.current[neighbor]; ok {
			neighborsAlive++
		}
	}

	if neighborsAlive == 3 ||
		neighborsAlive == 2 && isAlive {
		return true
	}

	return false
}

func (p point) neighbors() []point {
	neighbors := []point{}

	for i := 0; i < 9; i++ {
		if i == 4 {
			continue
		}

		row := i/3 - 1
		col := i%3 - 1

		neighbor := point{
			int64((canvasX + int(p.X) + col) % canvasX),
			int64((canvasY + int(p.Y) + row) % canvasY),
		}

		neighbors = append(neighbors, neighbor)
	}

	return neighbors
}

func runHTTPServer() {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	staticDir := strings.Join([]string{currentDir, "static"}, "/")

	http.Handle("/", http.FileServer(http.Dir(staticDir)))
	http.HandleFunc("/ws", websocketHandler)
	http.ListenAndServe(addr, nil)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := websocketClient{
		make(chan []point),
		make(chan []point),
		conn,
		true,
	}

	handleClient(&client)
}

func (client *websocketClient) updateEvent() event {

	return event(func(c *cells) {
		points := []point{}
		for p := range c.current {
			points = append(points, p)
		}

		select {
		case client.pointsOut <- points:
		case <-time.After(50 * time.Millisecond):
		}

	})
}

func (client *websocketClient) send() {
	for {
		if !client.connected {
			return
		}

		err := client.conn.WriteJSON(<-client.pointsOut)
		if err != nil {
			log.Println(err)
			client.conn.Close()
			client.connected = false
			return
		}
	}
}

func (client *websocketClient) receive() {
	for {
		if !client.connected {
			return
		}

		points := []point{}
		err := client.conn.ReadJSON(&points)
		if err != nil {
			log.Println(err)
			client.conn.Close()
			client.connected = false
			return
		}

		select {
		case client.pointsIn <- points:
		case <-time.After(100 * time.Millisecond):
		}

	}
}

func handleClient(client *websocketClient) {

	go client.receive()
	go client.send()

	fmt.Println("new client")
	for {

		if !client.connected {
			return
		}

		select {
		case newCells := <-client.pointsIn:
			events <- addCells(newCells)
		case <-time.After(100 * time.Millisecond):
			events <- client.updateEvent()
		}
	}
}
