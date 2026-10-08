package query

import (
	"fmt"
	"testing"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/user/utils"
)

func TestPassword(t *testing.T) {
	n, err := utils.HashForStorage("abb7bd807f5777f29fea952b55e25de87b1aed0080a5f7a5a8da3c98377c5e69")
	if err != nil {
		fmt.Println(err)
	}
	logger.Info(n)
}
