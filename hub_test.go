package main

import "testing"

func TestPostRaidIDAcceptsLegacyRaidIDField(t *testing.T) {
	h := NewHub()
	c := &client{send: make(chan []byte, 4)}

	h.handle(c, clientMsg{Type: "join", Room: 1, Index: 1, Participants: 2})
	h.handle(c, clientMsg{Type: "post_raid_id", Room: 1, RaidID: "12345678"})

	room := h.rooms[1]
	snap := room.state.Snapshot()
	if got := snap.RaidIDs[0]; got != "12345678" {
		t.Fatalf("raid id = %q, want %q", got, "12345678")
	}
}

func TestSignalDoneUsesJoinedClientIndex(t *testing.T) {
	h := NewHub()
	c := &client{send: make(chan []byte, 4)}

	h.handle(c, clientMsg{Type: "join", Room: 1, Index: 2, Participants: 2})
	h.handle(c, clientMsg{Type: "signal_done", Room: 1})

	room := h.rooms[1]
	snap := room.state.Snapshot()
	if !snap.Lamps[1] {
		t.Fatal("joined client index should be used when signal_done omits index")
	}
}
