package service

import "sync"

type UserStore interface {
	Save(user *User) error
	Find(username string) (*User, error)
}

type InMemoryUserStore struct {
	mutex sync.RWMutex
	users map[string]*User
}

func NewInMemoryUserStore() UserStore {
	return &InMemoryUserStore{
		users: make(map[string]*User, 0),
	}
}

func (m *InMemoryUserStore) Save(user *User) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.users[user.Username] != nil {
		return ErrAlreadyExist
	}
	m.users[user.Username] = user
	return nil
}

func (m *InMemoryUserStore) Find(username string) (*User, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	user := m.users[username]
	if user == nil {
		return nil, nil
	}
	return user.Clone(), nil
}
