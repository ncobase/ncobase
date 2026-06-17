package repository

import (
	"ncobase/plugin/ai/data/ent"
)

func IsNotFound(err error) bool {
	return ent.IsNotFound(err)
}
