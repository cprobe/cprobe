package fileutil

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// IsExist checks whether a file or directory exists.
// It returns false when the file or directory does not exist.
func IsExist(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}

// IsFile checks whether the path is a file,
// it returns false when it's a directory or does not exist.
func IsFile(fp string) bool {
	f, e := os.Stat(fp)
	if e != nil {
		return false
	}
	return !f.IsDir()
}

func ReadBytes(cpath string) ([]byte, error) {
	if !IsExist(cpath) {
		return nil, fmt.Errorf("%s not exists", cpath)
	}

	if !IsFile(cpath) {
		return nil, fmt.Errorf("%s not file", cpath)
	}

	return os.ReadFile(cpath)
}

func ReadYaml(cpath string, cptr interface{}) error {
	bs, err := ReadBytes(cpath)
	if err != nil {
		return fmt.Errorf("cannot read %s: %s", cpath, err.Error())
	}

	err = yaml.Unmarshal(bs, cptr)
	if err != nil {
		return fmt.Errorf("cannot parse %s: %s", cpath, err.Error())
	}

	return nil
}
