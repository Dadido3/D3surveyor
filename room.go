package main

import (
	"time"
)

type Room struct {
	Site *Site
	key  string

	Name      string
	CreatedAt time.Time
}

func (s *Site) NewRoom(name string) *Room {
	key := s.shortIDGen.MustGenerate()

	room := &Room{
		Site:      s,
		key:       key,
		Name:      name,
		CreatedAt: time.Now(),
	}

	s.Rooms[key] = room

	return room
}

func (r *Room) Key() string {
	return r.key
}

func (r *Room) Delete() {
	delete(r.Site.Rooms, r.Key())
}
