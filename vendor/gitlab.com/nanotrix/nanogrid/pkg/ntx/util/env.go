package util

import (
	"os"
)

func LookupEnv(key string) (string, error) {
	if s, ok := os.LookupEnv(key); ok {
		return s, nil
	} else {
		return "", NewNotFoundError("not exists: ${" + key + "}")
	}
}

func ExpandEnv(s string) (string, error) {
	return Expand(s, LookupEnv)
}

func Expand(s string, mapping func(string) (string, error)) (string, error) {
	var err error
	ret := os.Expand(s, func(input string) string {
		if output, e := mapping(input); e != nil {
			err = e
			return "${ERROR}"
		} else {
			return output
		}

	})
	return ret, err
}
