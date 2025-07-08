package structs

import (
	"github.com/ncobase/ncore/data/paging"
)

type Result[T paging.CursorProvider] = paging.Result[T]
