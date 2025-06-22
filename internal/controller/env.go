package controller

import "github.com/codingconcepts/env"

type Env struct {
	PodCIDR string `env:"POD_CIDR" default:"10.42.0.0/16"`
	SvcCIDR string `env:"SVC_CIDR" default:"10.43.0.0/16"`
}

func LoadEnv() (*Env, error) {
	var ev Env
	if err := env.Set(&ev); err != nil {
		return nil, err
	}
	return &ev, nil
}
