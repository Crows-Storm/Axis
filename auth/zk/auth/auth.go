package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/Crows-Storm/Axis/auth/zk/permission"
	"github.com/Crows-Storm/Axis/auth/zk/schnorr"
)

type User struct {
	Username    string
	Salt        []byte
	PublicX     *big.Int
	PublicY     *big.Int
	Permissions []byte
}

type UserStore struct {
	mu     sync.RWMutex
	users  map[string]*User
	nonces map[string]nonceEntry
}

type nonceEntry struct {
	nonce   []byte
	expires time.Time
	used    bool
}

func NewUserStore() *UserStore {
	return &UserStore{
		users:  make(map[string]*User),
		nonces: make(map[string]nonceEntry),
	}
}

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidProof = errors.New("invalid proof")
	ErrNonceInvalid = errors.New("nonce invalid or reused")
)

func (s *UserStore) Register(username string, salt, pubX, pubY []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[username]; ok {
		return ErrUserExists
	}
	s.users[username] = &User{
		Username: username,
		Salt:     append([]byte(nil), salt...),
		PublicX:  new(big.Int).SetBytes(pubX),
		PublicY:  new(big.Int).SetBytes(pubY),
	}
	return nil
}

func (s *UserStore) Salt(username string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	return append([]byte(nil), u.Salt...), nil
}

func (s *UserStore) IssueChallenge(username string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[username]; !ok {
		return nil, ErrUserNotFound
	}
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	s.nonces[username] = nonceEntry{
		nonce:   nonce,
		expires: time.Now().Add(2 * time.Minute),
		used:    false,
	}
	return nonce, nil
}

func (s *UserStore) VerifyLogin(username string, proof *schnorr.Proof) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[username]
	if !ok {
		return ErrUserNotFound
	}
	entry, ok := s.nonces[username]
	if !ok || entry.used || time.Now().After(entry.expires) {
		return ErrNonceInvalid
	}
	entry.used = true
	s.nonces[username] = entry

	ctx := buildCtx(username, entry.nonce)
	if !schnorr.Verify(u.PublicX, u.PublicY, proof, ctx) {
		return ErrInvalidProof
	}
	return nil
}

func ProvePermissionAbove(v, threshold, r *big.Int) (*permission.PermissionProof, error) {
	return permission.ProvePermissionAbove(v, threshold, r)
}

func VerifyPermissionAbove(proof *permission.PermissionProof, threshold *big.Int) bool {
	return permission.VerifyPermissionAbove(proof, threshold)
}

func buildCtx(username string, nonce []byte) []byte {
	return append([]byte("login:"+username+":"), nonce...)
}

func Hex(b []byte) string { return hex.EncodeToString(b) }
