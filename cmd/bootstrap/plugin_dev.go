//go:build dev || development || c2hlbgo

package bootstrap

// plugins to be loaded in development mode.
import (
	_ "ncobase/feature/content"
	_ "ncobase/feature/resource"
)
