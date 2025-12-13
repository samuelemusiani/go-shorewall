package goshorewall

import (
	"fmt"
	"io"
	"os"
)

func readWriteFile[S any](path string, f func([]byte, S) ([]byte, error), i S) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	buff, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	buff, err = f(buff, i)
	if err != nil {
		return err
	}

	n, err := file.WriteAt(buff, 0)
	if err != nil {
		return err
	}
	if n < len(buff) {
		return fmt.Errorf("failed to write complete data to zones file")
	}
	return file.Truncate(int64(n))
}
