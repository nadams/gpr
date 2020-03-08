package gps

import (
	"fmt"
	"os"
)

func BrightnessContrast(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}

	defer f.Close()

	b, err := Brightness(f)
	if err != nil {
		return 0, 0, err
	}

	c, err := Contrast(f)
	if err != nil {
		return 0, 0, err
	}

	return b, c, nil
}

func Brightness(f *os.File) (int, error) {
	_, err := f.Seek(0x17C, 0)
	if err != nil {
		return 0, err
	}

	c, err := getNumber(f)
	if err != nil {
		return 0, fmt.Errorf("could not get brightness: %w", err)
	}

	return c, nil
}

func Contrast(f *os.File) (int, error) {
	_, err := f.Seek(0x183, 0)
	if err != nil {
		return 0, err
	}

	c, err := getNumber(f)
	if err != nil {
		return 0, fmt.Errorf("could not get contrast: %w", err)
	}

	return c, nil
}

func getNumber(f *os.File) (int, error) {
	x := make([]byte, 1)
	n, err := f.Read(x)
	if err != nil {
		return 0, err
	}

	if n > 0 {
		return int(x[0]), nil
	}

	return 0, fmt.Errorf("could not get number")
}
