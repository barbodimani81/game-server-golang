package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

type GameMediator interface {
	GetMap(mapID int) (*Map, error)
}

type Player struct {
	name     string
	mapID    int
	ch       chan string
	mediator GameMediator
	mu       sync.Mutex
}

type Map struct {
	id       int
	players  map[string]*Player
	messages chan string
	mu       sync.RWMutex
}

type Game struct {
	players map[string]*Player
	maps    map[int]*Map
	mu      sync.RWMutex
}

func NewGame(mapIds []int) (*Game, error) {
	g := &Game{
		players: make(map[string]*Player),
		maps:    make(map[int]*Map),
	}

	for _, id := range mapIds {
		if id <= 0 {
			return nil, errors.New("map ID cannot be zero or negative")
		}
		g.maps[id] = &Map{
			id:       id,
			players:  make(map[string]*Player),
			messages: make(chan string, 100),
		}

		go g.maps[id].FanOutMessages()
	}

	return g, nil
}

func (g *Game) ConnectPlayer(name string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	n := strings.Title(strings.ToLower(name))
	if _, exists := g.players[n]; exists {
		return errors.New("player with this name is already exists")
	}

	p := &Player{
		name:     n,
		ch:       make(chan string, 100),
		mediator: g,
	}

	g.players[n] = p

	return nil
}

func (g *Game) SwitchPlayerMap(name string, mapId int) error {
	p, err := g.GetPlayer(name)
	if err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	newMap, exists := g.maps[mapId]
	if !exists {
		return errors.New("this map does not exists")
	}

	if p.mapID == mapId {
		return errors.New("player is in the current map")
	}

	// delete player from old map in the game
	if p.mapID != 0 {
		oldMap, exists := g.maps[p.mapID]
		if exists {
			oldMap.mu.Lock()
			delete(oldMap.players, p.GetName())
			oldMap.mu.Unlock()
		}
	}

	// add player to new map in the game
	newMap.mu.Lock()
	newMap.players[p.GetName()] = p
	newMap.mu.Unlock()

	p.mapID = mapId

	return nil
}

func (g *Game) GetPlayer(name string) (*Player, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	n := strings.Title(strings.ToLower(name))
	p, exists := g.players[n]
	if !exists {
		return nil, errors.New("player not found")
	}

	return p, nil
}

func (g *Game) GetMap(mapId int) (*Map, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	m, exists := g.maps[mapId]
	if !exists {
		return nil, errors.New("map not found")
	}

	return m, nil
}

func (m *Map) FanOutMessages() {
	for msg := range m.messages {
		senderName := strings.Split(msg, " says: ")[0]
		m.mu.RLock()

		for _, p := range m.players {
			if p.GetName() == senderName {
				continue
			}

			select {
			case p.ch <- msg:
			default:
				// skip to preven blocking
			}
		}
		m.mu.RUnlock()
	}
}

func (p *Player) GetChannel() <-chan string {
	return p.ch
}

func (p *Player) SendMessage(msg string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.mapID == 0 {
		return errors.New("player is not in any map")
	}

	m, err := p.mediator.GetMap(p.mapID)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("%s says: %s", p.GetName(), msg)

	select {
	case m.messages <- message:
		return nil
	default:
		return errors.New("channel is full")
	}
}

func (p *Player) GetName() string {
	return p.name
}
