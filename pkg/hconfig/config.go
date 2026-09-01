package hconfig

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/spf13/viper"
)

// ErrInvalidTarget 表示配置接收目标不是非 nil 指针。
var ErrInvalidTarget = errors.New("hconfig: target must be a non-nil pointer")

// Validator 允许配置对象在加载后执行自身校验。
type Validator interface {
	Validate() error
}

// Load 从 path/name.yaml 加载配置到 target。
// target 中预先设置、且 YAML 未覆盖的默认值会被保留。
func Load(path, name string, target any) error {
	value := reflect.ValueOf(target)
	if target == nil || value.Kind() != reflect.Ptr || value.IsNil() {
		return ErrInvalidTarget
	}

	file := filepath.Join(path, name+".yaml")
	v := viper.New()
	v.SetConfigFile(file)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("hconfig: read %s: %w", file, err)
	}
	if err := v.Unmarshal(target); err != nil {
		return fmt.Errorf("hconfig: decode %s: %w", file, err)
	}
	if validator, ok := target.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("hconfig: validate %s: %w", file, err)
		}
	}
	return nil
}
