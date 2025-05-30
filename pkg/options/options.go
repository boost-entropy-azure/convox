package options

import "time"

func Bool(value bool) *bool {
	v := value
	return &v
}

func Duration(value time.Duration) *time.Duration {
	v := value
	return &v
}

func Int(value int) *int {
	v := value
	return &v
}

func Int32(value int32) *int32 {
	v := value
	return &v
}

func Int64(value int64) *int64 {
	v := value
	return &v
}

func String(value string) *string {
	v := value
	return &v
}

func StringOrNil(value string) *string {
	if value == "" {
		return nil
	}
	v := value
	return &v
}

func Time(value time.Time) *time.Time {
	v := value
	return &v
}

func StringValueSafe(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func StringPtrArray(arr []string) []*string {
	v := []*string{}
	for i := range arr {
		v = append(v, &arr[i])
	}
	return v
}
