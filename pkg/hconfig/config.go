package hconfig

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

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
	return load(path, name, target, "", false)
}

// LoadWithEnv 从 YAML 加载配置，并允许环境变量覆盖 YAML 中的配置项。
// 环境变量使用 prefix 前缀，配置路径中的点和横线会转换为下划线。
// 例如 prefix 为 LIFE_CORE 时，scheduler.timezone 对应
// LIFE_CORE_SCHEDULER_TIMEZONE。
func LoadWithEnv(path, name, prefix string, target any) error {
	return load(path, name, target, prefix, true)
}

func load(path, name string, target any, envPrefix string, enableEnv bool) error {
	value := reflect.ValueOf(target)
	if target == nil || value.Kind() != reflect.Ptr || value.IsNil() {
		return ErrInvalidTarget
	}

	file := filepath.Join(path, name+".yaml")
	v := viper.New()
	v.SetConfigFile(file)
	v.SetConfigType("yaml")
	if enableEnv {
		v.SetEnvPrefix(envPrefix)
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
		v.AutomaticEnv()
		if err := bindStructEnv(v, value.Elem().Type(), ""); err != nil {
			return err
		}
	}
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

func bindStructEnv(v *viper.Viper, structType reflect.Type, prefix string) error {
	for structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}
	if structType.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}
		name, squash, skip := mapstructureField(field)
		if skip {
			continue
		}
		key := name
		if squash {
			key = prefix
		} else if prefix != "" {
			key = prefix + "." + name
		}
		fieldType := field.Type
		for fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Struct {
			if err := bindStructEnv(v, fieldType, key); err != nil {
				return err
			}
			continue
		}
		if err := v.BindEnv(key); err != nil {
			return fmt.Errorf("hconfig: bind env %s: %w", key, err)
		}
	}
	return nil
}

func mapstructureField(field reflect.StructField) (name string, squash, skip bool) {
	parts := strings.Split(field.Tag.Get("mapstructure"), ",")
	name = parts[0]
	if name == "-" {
		return "", false, true
	}
	for _, option := range parts[1:] {
		if option == "squash" {
			squash = true
		}
	}
	if name == "" {
		name = strings.ToLower(field.Name)
	}
	return name, squash, false
}
