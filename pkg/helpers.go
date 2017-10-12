package pkg

import (
	"strconv"
	"time"
)

const (
	SegmentDur = 5 * time.Minute // we only work in around 5 mins
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
	HourlyCounterSet   string
	HourlyCounterHSet  string
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

	k.Src.CurrentCounterSet, k.Src.HourlyCounterSet = generateSegmentPrefixes("set:counter:src", tr)
	k.Src.CurrentCounterHSet, k.Src.HourlyCounterHSet = generateSegmentPrefixes("hset:counter:src", tr)
	k.Dst.CurrentCounterSet, k.Dst.HourlyCounterSet = generateSegmentPrefixes("set:counter:dst", tr)
	k.Dst.CurrentCounterHSet, k.Dst.HourlyCounterHSet = generateSegmentPrefixes("hset:counter:dst", tr)

	return k
}

// HashSetNames combines the prefix and the srcMember
func (k *KeyNames) HashSetNames(srcMember string) (string, string) {
	var (
		current = k.CurrentCounterHSet + seperator + srcMember
		hourly  = k.HourlyCounterHSet + seperator + srcMember
	)

	return current, hourly
}

func generateSegmentPrefixes(directionSuffix string, tr time.Time) (current string, hourly string) {
	var (
		currentSuffix = directionSuffix + seperator + strconv.FormatInt(tr.Unix(), 10)
		hourlySuffix  = directionSuffix + seperator + strconv.FormatInt(tr.Add(-(time.Hour/2)).Round(time.Hour).Unix(), 10)
	)

	return currentSuffix, hourlySuffix
}
