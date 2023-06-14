package hash_test

import (
	"testing"

	"github.com/demmydemon/abventure/hash"
)

func TestSingleHash(t *testing.T) {
	knownGoodResult := "0044ff4a"
	result := hash.Single("This is a test")
	if result != knownGoodResult {
		t.Errorf("hash.Single generated hash was wrong. Expected %s, got %s", knownGoodResult, result)
	}
}

func TestStart(t *testing.T) {
	result := hash.Single("Start")
	if result != hash.PrecalcStart {
		t.Errorf("hash.PrecalcStart does not match. Expected %s, got %s", hash.PrecalcStart, result)
	}
}

func TestHashes(t *testing.T) {

	precalculated := map[string]string{
		"0044ff4a": "This is a test",
		"3751877a": "same every time!",
		"794a4dee": "I feel compelled to include a much longer string in the tests as well. Shouldn't matter, but hey, here we are.",
	}

	input := make([]string, len(precalculated))
	i := 0
	for _, str := range precalculated {
		input[i] = str
		i++
	}

	result := hash.Mapped(input...)
	if len(result) != len(precalculated) {
		t.Errorf("hash.Mapped generated wrong size map, expected %d key, got %d", len(precalculated), len(result))
	}
	for key, value := range precalculated {
		val, ok := result[key]

		if !ok {
			t.Errorf("hash.Mapped result is missing expected key %s", key)
			continue
		}
		if val != value {
			t.Errorf("hash.Mapped result has unexpected value, expected %s, got %s", value, val)
			continue
		}
	}
}
