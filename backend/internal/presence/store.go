package presence

import (
	"sync"
	"time"
)

const defaultTTL = 40 * time.Second

type UserInfo struct {
	UserID   uint
	Fullname string
	Email    string
}

type Session struct {
	UserID   uint
	Fullname string
	Email    string
	LastSeen time.Time
}

type Store struct {
	mu   sync.Mutex
	ttl  time.Duration
	data map[uint]Session
}

func NewStore() *Store {
	return &Store{
		ttl:  defaultTTL,
		data: make(map[uint]Session),
	}
}

func (s *Store) Touch(fileID uint, user UserInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[fileID] = Session{
		UserID:   user.UserID,
		Fullname: user.Fullname,
		Email:    user.Email,
		LastSeen: time.Now(),
	}
}

func (s *Store) Leave(fileID, userID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.data[fileID]
	if !ok || session.UserID != userID {
		return
	}
	delete(s.data, fileID)
}

func (s *Store) Get(fileID uint) (Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked()
	session, ok := s.data[fileID]
	return session, ok
}

func (s *Store) GetAll() map[uint]Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked()
	out := make(map[uint]Session, len(s.data))
	for id, session := range s.data {
		out[id] = session
	}
	return out
}

func (s *Store) cleanupLocked() {
	now := time.Now()
	for id, session := range s.data {
		if now.Sub(session.LastSeen) > s.ttl {
			delete(s.data, id)
		}
	}
}
