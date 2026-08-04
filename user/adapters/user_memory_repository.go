package adapters

import (
	"context"
	"sync"
	"time"

	"github.com/Crows-Storm/Axis/common/config/logger"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
	"github.com/sirupsen/logrus"
)

type MemoryUserRepository struct {
	lock  *sync.RWMutex
	store []*domain.User
}

func NewMemoryUserRepository() *MemoryUserRepository {
	infoStore := make([]*domain.User, 0)

	// init a data to memory repository use in test
	infoStore = append(infoStore, &domain.User{
		Id:         123,
		LoginId:    "Sander",
		Password:   "ddferewevrewwdwdwdwe2fwfwefwfwfwewqd",
		Email:      "sanderQiu@hotmail.com",
		Status:     1,
		Deleted:    0,
		CreateTime: time.Time{},
		UpdateTime: time.Time{},
	})
	return &MemoryUserRepository{
		lock:  &sync.RWMutex{},
		store: infoStore,
	}
}

func (m *MemoryUserRepository) GetInfo(id int64) (*domain.User, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, user := range m.store {
		logger.Infof("👀 user memory repository all store data: %v", user)
	}
	if len(m.store) == 0 {
		return nil, domain.NotFoundError{UserId: id}
	}

	for _, v := range m.store {
		if v.Id == id {
			logrus.Debugf("memory_user_repo_get || id=%d || res=%+v", id, *v)
			return v, nil
		}
	}
	return nil, domain.NotFoundError{UserId: id}
}

func (m *MemoryUserRepository) Create(_ context.Context, user *domain.User) (*domain.User, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	newUser := &domain.User{
		Id:         user.Id,
		LoginId:    user.LoginId,
		Password:   user.Password,
		Email:      user.Email,
		Status:     1, // use db default
		Deleted:    0, // use db default
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	m.store = append(m.store, newUser)

	logger.WithFields(logrus.Fields{
		"input_user":         user,
		"store_after_create": m.store,
	}).Debug("memory_user_repo_create")
	return newUser, nil
}

func (m *MemoryUserRepository) Update(ctx context.Context, user *domain.User, updateFun func(context.Context, *domain.User) (*domain.User, error)) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if len(m.store) == 0 {
		return domain.NotFoundError{UserId: user.Id}
	}

	found := false
	for i, v := range m.store {
		if v.Id == user.Id {
			found = true
			updatedUser, err := updateFun(ctx, user)
			if err != nil {
				return err
			}
			m.store[i] = updatedUser
		}
	}

	if found {
		return domain.NotFoundError{UserId: user.Id}
	}
	return nil

}
