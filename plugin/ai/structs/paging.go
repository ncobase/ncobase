package structs

import "github.com/ncobase/ncore/paging"

type Result[T paging.CursorProvider] = paging.Result[T]
