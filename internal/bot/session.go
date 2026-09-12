package bot

import (
	"sync"

	"github.com/google/uuid"
	"time"
)

// bookingSession — черновик бронирования клиента в памяти.
type bookingSession struct {
	ClientID         uuid.UUID
	ChatID           int64
	Date             time.Time
	Hours            int
	StartsAt         time.Time
	RentID           uuid.UUID
	AwaitingReceipt  bool
}

type sessionStore struct {
	mu   sync.Mutex
	byTG map[int64]*bookingSession
}

func newSessionStore() *sessionStore {
	return &sessionStore{byTG: make(map[int64]*bookingSession)}
}

func (s *sessionStore) get(telegramID int64) *bookingSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.byTG[telegramID]
}

func (s *sessionStore) put(telegramID int64, sess *bookingSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byTG[telegramID] = sess
}

func (s *sessionStore) clear(telegramID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byTG, telegramID)
}
