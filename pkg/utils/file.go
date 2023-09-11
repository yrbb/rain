package utils

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return err == nil
}

func FileCopy(dest, source string) error {
	df, err := os.Create(dest)
	if err != nil {
		return err
	}

	f, err := os.Open(source)
	if err != nil {
		return err
	}

	_, err = io.Copy(df, f)
	_ = f.Close()

	return err
}

func FileRead(path string, fn func(b []byte, line int) bool, maxLine ...int) error {
	fi, err := os.Open(path)
	if err != nil {
		return err
	}

	ms := 40960 // 40K
	if len(maxLine) > 0 {
		ms = maxLine[0]
	}

	line := 0

	br := bufio.NewReader(fi)
	eb := string([]byte{0})
	tb := make([]byte, ms)

	for {
		b, isPrefix, err := br.ReadLine()
		tb = append(tb, b...)

		if isPrefix {
			continue
		}

		if err == io.EOF {
			break
		}

		line++

		tb = bytes.Trim(tb, eb)

		r := fn(tb, line)

		tb = []byte{}

		if !r {
			break
		}
	}

	return nil
}

func IsWritable(path string) (isWritable bool, err error) {
	if err = syscall.Access(path, syscall.O_RDWR); err == nil {
		isWritable = true
	}

	return
}

func RootPath() (s string) {
	dir, _ := filepath.Abs(filepath.Dir(s))

	return strings.ReplaceAll(dir, "\\", "/")
}

func IsFile(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fi.Mode().IsRegular()
}

func IsDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fi.Mode().IsDir()
}

func MakeDir(s string) error {
	return os.MkdirAll(s, 0777)
}

func ScanDir(path string) ([]string, error) {
	files, err := os.ReadDir(path)

	if err != nil {
		return nil, err
	}

	var res []string
	for _, fi := range files {
		if !fi.IsDir() {
			res = append(res, fi.Name())
		}
	}

	return res, nil
}
