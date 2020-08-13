package rand

import (
	crand "crypto/rand"
	"encoding/binary"
	mrand "math/rand"
	"sync"
	"time"
)

var (
	// SecureRand is a thread-safe math/rand.Source64 backed by crypto/rand.Reader.
	SecureRand = mrand.New(cryptoRandSource{})
	// InsecureRand is a thread-safe math/rand.Source64 seeded by the program start time.
	InsecureRand = mrand.New(&lockedMathRandSource{source: mrand.NewSource(time.Now().UnixNano()).(mrand.Source64)})
)

type cryptoRandSource struct{}

func (s cryptoRandSource) Seed(seed int64) {}

func (s cryptoRandSource) Int63() int64 {
	// Reset the most significant bit to 0
	return int64(s.Uint64() & ^uint64(1<<63))
}

func (s cryptoRandSource) Uint64() uint64 {
	var v uint64
	err := binary.Read(crand.Reader, binary.BigEndian, &v)
	if err != nil {
		panic(err)
	}
	return v
}

type lockedMathRandSource struct {
	mutex  sync.Mutex
	source mrand.Source64
}

func (s *lockedMathRandSource) Seed(seed int64) {
	s.mutex.Lock()
	s.source.Seed(seed)
	s.mutex.Unlock()
}

func (s *lockedMathRandSource) Int63() int64 {
	s.mutex.Lock()
	v := s.source.Int63()
	s.mutex.Unlock()
	return v
}

func (s *lockedMathRandSource) Uint64() uint64 {
	s.mutex.Lock()
	v := s.source.Uint64()
	s.mutex.Unlock()
	return v
}

// StringWithAlphabet generates a random string of the specific length
// with the given alphabet using the given RNG.
func StringWithAlphabet(length int, alphabet string, r *mrand.Rand) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = alphabet[r.Intn(len(alphabet))]
	}
	return string(b)
}
