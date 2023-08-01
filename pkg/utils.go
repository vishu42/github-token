package pkg

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
)

func CreateHash(input string) (string, error) {
	hash := sha256.New()

	_, err := hash.Write([]byte(input))
	if err != nil {
		return "", err
	}

	return string(hash.Sum(nil)), nil
}

func ReadFile(file string) (string, error) {
	f, err := os.OpenFile(file, os.O_RDONLY|os.O_CREATE, 0o666)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)

	data := ""

	for sc.Scan() {
		data += sc.Text()
	}

	return data, nil
}
