package handler

import "fmt"

// formatSize formats size in bytes to human-readable format
func formatSize(sizeInBytes int64) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
		TB
	)

	size := float64(sizeInBytes)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", size/TB)
	case size >= GB:
		return fmt.Sprintf("%.2f GB", size/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", size/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", size/KB)
	default:
		return fmt.Sprintf("%d bytes", sizeInBytes)
	}
}
