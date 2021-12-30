package main

import (
	"fmt"
	"sync"
	"time"
)

type (
	Event struct {
		data string
	}

	Observer interface {
		NotifyCallback(Event)
	}

	Subject interface {
		AddListener(Observer)
		RemoveListener(Observer)
		Notify(Event)
	}

	eventObserver struct {
		id   int
		time time.Time
	}

	eventSubject struct {
		observers sync.Map
	}
)

func (e *eventObserver) NotifyCallback(event Event) {
	fmt.Printf("Observer: %d recieved: %s after %v\n", e.id, event.data, time.Since(e.time))
}

func (s *eventSubject) AddListener(obs Observer) {
	s.observers.Store(obs, struct{}{})
}

func (s *eventSubject) RemoveListener(obs Observer) {
	s.observers.Delete(obs)
}

func (s *eventSubject) Notify(event Event) {
	s.observers.Range(func(key interface{}, value interface{}) bool {
		if key == nil || value == nil {
			return false
		}
		key.(Observer).NotifyCallback(event)
		return true
	})
}
