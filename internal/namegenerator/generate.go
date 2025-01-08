package namegenerator

import (
	_ "embed"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

//go:embed animals.txt
var animalsb []byte

//go:embed colors.txt
var colorsb []byte

//go:embed adjectives.txt
var adjectivesb []byte

// Generate a random name in the format adjective-color-animal
func Generate() string {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	animals := strings.Split(string(animalsb), "\n")
	colors := strings.Split(string(colorsb), "\n")
	adjectives := strings.Split(string(adjectivesb), "\n")
	n := strconv.FormatInt(int64(rand.Intn(8999)+1000), 10)

	return fmt.Sprintf("%s-%s-%s-%s",
		adjectives[r.Intn(len(adjectives))],
		colors[r.Intn(len(colors))],
		animals[r.Intn(len(colors))],
		n,
	)
}
