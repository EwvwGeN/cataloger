package config

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type Validator struct {
	EmailValidate    string `mapstructure:"email"`
	PasswordValidate string `mapstructure:"password"`
}

func (v *Validator) mustBeRegex() error {
	b, _ := json.Marshal(v)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	for i, v := range m {
		_, err := regexp.Compile(v.(string))
		if err != nil {
			return fmt.Errorf("incorrect %s", i)
		}
	}
	return nil
}