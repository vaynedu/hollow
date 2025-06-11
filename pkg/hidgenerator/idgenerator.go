package hidgenerator

// id 生成器
type IdGenerator interface {
	GenerateRequestID() string
}
