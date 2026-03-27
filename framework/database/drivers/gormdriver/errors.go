package gormdriver

import (
	"fmt"
	"strings"

	"github.com/zhoudm1743/go-fast/framework/contracts"

	"gorm.io/gorm"
)

// wrapError 将 GORM 错误映射为框架级 Sentinel Error。
func wrapError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case isErr(err, gorm.ErrRecordNotFound):
		return fmt.Errorf("%w: %v", contracts.ErrRecordNotFound, err)
	case isErr(err, gorm.ErrDuplicatedKey):
		return fmt.Errorf("%w: %v", contracts.ErrDuplicatedKey, err)
	case isErr(err, gorm.ErrInvalidTransaction):
		return fmt.Errorf("%w: %v", contracts.ErrInvalidTransaction, err)
	default:
		msg := err.Error()
		if strings.Contains(msg, "Deadlock") || strings.Contains(msg, "deadlock") {
			return fmt.Errorf("%w: %v", contracts.ErrDeadlock, err)
		}
		if strings.Contains(msg, "Duplicate entry") || strings.Contains(msg, "duplicate key") ||
			strings.Contains(msg, "UNIQUE constraint failed") {
			return fmt.Errorf("%w: %v", contracts.ErrDuplicatedKey, err)
		}
		return err
	}
}

func isErr(err, target error) bool {
	return err == target || (err != nil && err.Error() == target.Error())
}
