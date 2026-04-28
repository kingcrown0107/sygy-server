package main

import (
	"sync"
	"time"
)

const gracePeriod = 1 * time.Second

type RoomState struct {
	RoomID           int
	ParticipantCount int
	Lamps            []bool
	RaidIDs          []string
	AllDoneAt        *time.Time

	mu         sync.Mutex
	graceTimer *time.Timer
	onReset    func()
}

func newRoomState(roomID, participants int) *RoomState {
	return &RoomState{
		RoomID:           roomID,
		ParticipantCount: participants,
		Lamps:            make([]bool, participants),
		RaidIDs:          make([]string, participants),
	}
}

func (r *RoomState) grow(participants int) {
	if participants <= r.ParticipantCount {
		return
	}
	for len(r.Lamps) < participants {
		r.Lamps = append(r.Lamps, false)
	}
	for len(r.RaidIDs) < participants {
		r.RaidIDs = append(r.RaidIDs, "")
	}
	r.ParticipantCount = participants
}

func (r *RoomState) SignalDone(index, participants int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.grow(participants)

	if index < 0 || index >= r.ParticipantCount {
		return false
	}
	r.Lamps[index] = true

	allDone := r.allDoneLocked()
	if allDone && r.AllDoneAt == nil {
		now := time.Now()
		r.AllDoneAt = &now
		r.startGraceTimer()
	}
	return allDone
}

func (r *RoomState) PostRaidID(index, participants int, battleID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.grow(participants)

	if index >= 0 && index < r.ParticipantCount {
		r.RaidIDs[index] = battleID
	}
}

func (r *RoomState) ResetLamps() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cancelGraceTimer()
	for i := range r.Lamps {
		r.Lamps[i] = false
	}
	r.AllDoneAt = nil
}

func (r *RoomState) Snapshot() roomSnapshot {
	r.mu.Lock()
	defer r.mu.Unlock()

	lamps := make([]bool, len(r.Lamps))
	copy(lamps, r.Lamps)

	ids := make([]string, len(r.RaidIDs))
	copy(ids, r.RaidIDs)

	var allDoneAt *time.Time
	if r.AllDoneAt != nil {
		t := *r.AllDoneAt
		allDoneAt = &t
	}

	return roomSnapshot{
		RoomID:           r.RoomID,
		ParticipantCount: r.ParticipantCount,
		Lamps:            lamps,
		RaidIDs:          ids,
		AllDoneAt:        allDoneAt,
	}
}

func (r *RoomState) allDoneLocked() bool {
	if r.ParticipantCount == 0 {
		return false
	}
	for i := 0; i < r.ParticipantCount; i++ {
		if !r.Lamps[i] {
			return false
		}
	}
	return true
}

func (r *RoomState) startGraceTimer() {
	r.cancelGraceTimer()
	r.graceTimer = time.AfterFunc(gracePeriod, func() {
		r.ResetLamps()
		if r.onReset != nil {
			r.onReset()
		}
	})
}

func (r *RoomState) cancelGraceTimer() {
	if r.graceTimer != nil {
		r.graceTimer.Stop()
		r.graceTimer = nil
	}
}

type roomSnapshot struct {
	RoomID           int        `json:"room"`
	ParticipantCount int        `json:"participant_count"`
	Lamps            []bool     `json:"lamps"`
	RaidIDs          []string   `json:"raid_ids"`
	AllDoneAt        *time.Time `json:"all_done_at"`
}
