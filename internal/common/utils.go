package common

import (
	"net/url"
	"path"
)

func Join(basePath string, paths ...string) (*url.URL, error) {
	u, err := url.Parse(basePath)
	if err != nil {
		return nil, err
	}

	p2 := append([]string{u.Path}, paths...)
	result := path.Join(p2...)
	u.Path = result

	return u, nil
}
