package util

import "github.com/sony/sonyflake"

var Flake *sonyflake.Sonyflake

func init() {
	Flake = sonyflake.NewSonyflake(sonyflake.Settings{
		MachineID: func() (uint16, error) {
			return uint16(1), nil
		},
	})
	if Flake == nil {
		panic("sonyflake init failed")
	}
}

// GenerateID Generate a primary key ID for int64
func GenerateID() int64 {
	id, err := Flake.NextID()
	if err != nil {
		panic(err)
	}
	return int64(id)
}
