package user

import (
	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/user/utils"
	"github.com/sirupsen/logrus"
)

// The PasswordHasher is domain power
type PasswordHasher interface {
	Hash(h1Hex string) (string, error)
	Compare(storedEncrypted, clientHash, requestID string) (bool, error)
}

type Hash256PasswordHelper struct{}

func NewHash256PasswordHelper() *Hash256PasswordHelper {
	return &Hash256PasswordHelper{}
}

func (p Hash256PasswordHelper) Hash(h1Hex string) (string, error) {
	return utils.HashForStorage(h1Hex)
}

func (p Hash256PasswordHelper) Compare(storedEncrypted, clientHash, requestID string) (bool, error) {
	if err := utils.VerifyPassword(storedEncrypted, clientHash, requestID); err != nil {
		logger.WithFields(logrus.Fields{
			"error":     err,
			"requestID": requestID,
		})
		return false, err
	}
	return true, nil
}
