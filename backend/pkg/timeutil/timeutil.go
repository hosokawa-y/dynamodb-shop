package timeutil

import "time"

// ParseTime はRFC3339形式の文字列をtime.Timeに変換する
func ParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
