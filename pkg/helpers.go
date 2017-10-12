package pkg

import (
	"strconv"
	"strings"
	"time"
)

const (
	// SegmentDur specifies the duration for work segments
	SegmentDur = 5 * time.Minute
	seperator  = ":"
)

// AllKeys holds the redis key names for processings...
type AllKeys struct {
	Dst KeyNames
	Src KeyNames
}

// KeyNames holds the key names per direction
type KeyNames struct {
	CurrentCounterSet  string
	CurrentCounterHSet string
}

// GetCurrentSegment returns the current segment's time
func GetCurrentSegment() time.Time {
	return time.Now().UTC().Add(-(SegmentDur / 2)).Round(SegmentDur)
}

// GetLastProcessibleSegment return the two previous segment time. We compact the
// values that are set 2*d duration before. d/2 is required for time.Round(d).
// It rounds up after the halfway values.
func GetLastProcessibleSegment(t time.Time) time.Time {
	return t.Add(-SegmentDur * 2).Add(-(SegmentDur / 2)).Round(SegmentDur)
}

// GenerateKeyNames generates the redis key names
func GenerateKeyNames(tr time.Time) *AllKeys {
	k := &AllKeys{
		Dst: KeyNames{},
		Src: KeyNames{},
	}

	k.Src.CurrentCounterSet = generateSegmentPrefix("set:counter:src", tr)
	k.Src.CurrentCounterHSet = generateSegmentPrefix("hset:counter:src", tr)
	k.Dst.CurrentCounterSet = generateSegmentPrefix("set:counter:dst", tr)
	k.Dst.CurrentCounterHSet = generateSegmentPrefix("hset:counter:dst", tr)
	return k
}

// HashSetName combines the prefix and the srcMember
func (k *KeyNames) HashSetName(srcMember string) string {
	return k.CurrentCounterHSet + seperator + srcMember
}

func generateSegmentPrefix(directionSuffix string, tr time.Time) string {
	return directionSuffix + seperator + strconv.FormatInt(tr.Unix(), 10)
}

// ParsedKeyName holds the parts of a key as separate entities
type ParsedKeyName struct {
	Type       string
	WorkerName string
	Direction  string
	Segment    string
	Name       string
}

// ParseKeyName parses the given key.
func ParseKeyName(s string) *ParsedKeyName {
	parts := strings.Split(s, seperator)
	if len(parts) != 4 && len(parts) != 5 {
		panic("key names should be consisted of either 4 or 5 parts")
	}

	pk := &ParsedKeyName{
		Type:       parts[0],
		WorkerName: parts[1],
		Direction:  parts[2],
		Segment:    parts[3],
	}
	if len(parts) == 5 {
		pk.Name = parts[4]
	}

	return pk
}
