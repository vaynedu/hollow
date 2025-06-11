package hidgenerator

import "github.com/google/uuid"

// Uuid 实现 IdGenerator 接口
type Uuid struct {
}

func NewUuid() *Uuid {
	return &Uuid{}
}

// GenerateRequestID 实现 IdGenerator 接口方法
func (u *Uuid) GenerateRequestID() string {
	return uuid.New().String()
}
