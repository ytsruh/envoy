package tools

import (
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"math/big"

	cli "github.com/pressly/cli"
)

var passwordCmd = &cli.Command{
	Name:      "password",
	ShortHelp: "Generate a random password",
	Usage:     "envoy tools password [--length=N] [--encoded]",
	Flags: cli.FlagsFunc(func(f *flag.FlagSet) {
		f.Int("length", 16, "Length of the password")
		f.Bool("encoded", false, "Generate base64 encoded password instead")
	}),
	Exec: func(ctx context.Context, s *cli.State) error {
		length := cli.GetFlag[int](s, "length")
		encoded := cli.GetFlag[bool](s, "encoded")

		pwd, err := generatePassword(length, encoded)
		if err != nil {
			return fmt.Errorf("failed to generate password: %w", err)
		}

		fmt.Fprintln(s.Stdout, pwd)
		return nil
	},
}

func generatePassword(length int, encoded bool) (string, error) {
	if encoded {
		b := make([]byte, length)
		_, err := rand.Read(b)
		if err != nil {
			return "", errors.New("failed to generate random bytes")
		}
		return base64URLEncode(b), nil
	}

	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz!?@#$%&"
	r := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", errors.New("failed to generate random number")
		}
		r[i] = letters[num.Int64()]
	}
	return string(r), nil
}

func base64URLEncode(b []byte) string {
	const encodeStd = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	result := make([]byte, (len(b)+2)/3*4)
	for i, j := 0, 0; i < len(b); i, j = i+3, j+4 {
		var v int
		switch len(b) - i {
		case 1:
			v = int(b[i]) << 16
			result[j] = encodeStd[v>>18&63]
			result[j+1] = encodeStd[v>>12&63]
			result[j+2] = '='
			result[j+3] = '='
		case 2:
			v = int(b[i])<<16 | int(b[i+1])<<8
			result[j] = encodeStd[v>>18&63]
			result[j+1] = encodeStd[v>>12&63]
			result[j+2] = encodeStd[v>>6&63]
			result[j+3] = '='
		default:
			v = int(b[i])<<16 | int(b[i+1])<<8 | int(b[i+2])
			result[j] = encodeStd[v>>18&63]
			result[j+1] = encodeStd[v>>12&63]
			result[j+2] = encodeStd[v>>6&63]
			result[j+3] = encodeStd[v&63]
		}
	}
	return string(result)
}
