package main

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type clientMsg struct {
	Type         string `json:"type"`
	Room         int    `json:"room"`
	Index        int    `json:"index"`
	Participants int    `json:"participants"`
	BattleID     string `json:"battle_id"`
	RaidID       string `json:"raid_id"`
}

type serverMsg struct {
	Type string `json:"type"`
	roomSnapshot
}

type client struct {
	conn   *websocket.Conn
	send   chan []byte
	roomID int
	index  int
}

type room struct {
	state   *RoomState
	clients map[*client]struct{}
}

type Hub struct {
	rooms map[int]*room
	mu    sync.Mutex
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[int]*room)}
}

func (h *Hub) getOrCreateRoom(roomID, participants int) *room {
	r, ok := h.rooms[roomID]
	if !ok {
		state := newRoomState(roomID, participants)
		r = &room{
			state:   state,
			clients: make(map[*client]struct{}),
		}
		state.onReset = func() {
			h.mu.Lock()
			rm, exists := h.rooms[roomID]
			h.mu.Unlock()
			if exists {
				h.broadcast(rm, "room_state")
			}
		}
		h.rooms[roomID] = r
	}
	return r
}

func (h *Hub) broadcast(r *room, msgType string) {
	snap := r.state.Snapshot()
	msg := serverMsg{Type: msgType, roomSnapshot: snap}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("broadcast marshal error: %v", err)
		return
	}
	for c := range r.clients {
		select {
		case c.send <- data:
		default:
		}
	}
}

func (h *Hub) broadcastAllDone(r *room) {
	snap := r.state.Snapshot()
	msg := serverMsg{Type: "all_done", roomSnapshot: snap}
	data, _ := json.Marshal(msg)
	for c := range r.clients {
		select {
		case c.send <- data:
		default:
		}
	}
}

func (h *Hub) handle(c *client, msg clientMsg) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch msg.Type {
	case "join":
		r := h.getOrCreateRoom(msg.Room, msg.Participants)
		if old, ok := h.rooms[c.roomID]; ok && c.roomID != msg.Room {
			delete(old.clients, c)
		}
		c.roomID = msg.Room
		c.index = msg.Index
		r.clients[c] = struct{}{}
		snap := r.state.Snapshot()
		data, _ := json.Marshal(serverMsg{Type: "room_state", roomSnapshot: snap})
		select {
		case c.send <- data:
		default:
		}

	case "signal_done":
		r := h.getOrCreateRoom(msg.Room, msg.Participants)
		index := msg.Index
		if index == 0 {
			index = c.index
		}
		participants := msg.Participants
		if participants == 0 {
			participants = r.state.ParticipantCount
		}
		allDone := r.state.SignalDone(index-1, participants)
		h.broadcast(r, "room_state")
		if allDone {
			h.broadcastAllDone(r)
		}

	case "post_raid_id":
		r := h.getOrCreateRoom(msg.Room, msg.Participants)
		index := msg.Index
		if index == 0 {
			index = c.index
		}
		participants := msg.Participants
		if participants == 0 {
			participants = r.state.ParticipantCount
		}
		battleID := msg.BattleID
		if battleID == "" {
			battleID = msg.RaidID
		}
		r.state.PostRaidID(index-1, participants, battleID)
		h.broadcast(r, "room_state")

	case "reset_lamps":
		if r, ok := h.rooms[msg.Room]; ok {
			r.state.ResetLamps()
			h.broadcast(r, "room_state")
		}
	}
}

func (h *Hub) removeClient(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if r, ok := h.rooms[c.roomID]; ok {
		delete(r.clients, c)
	}
	close(c.send)
}

func (h *Hub) readPump(c *client) {
	defer func() {
		h.removeClient(c)
		c.conn.Close()
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("ws read error: %v", err)
			}
			return
		}
		var msg clientMsg
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("json unmarshal error: %v", err)
			continue
		}
		h.handle(c, msg)
	}
}

func writePump(c *client) {
	defer c.conn.Close()
	for data := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("ws write error: %v", err)
			return
		}
	}
}
