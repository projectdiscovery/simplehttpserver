package unit

// ToMb converts bytes to megabytes
func ToMb(n int) int64 {
	return int64(n) * 1024 * 1024
}
