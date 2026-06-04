package pgdb

import (
	"time"

	"github.com/Masterminds/squirrel"
)

const (
	carReadModelsTableName = "core.car_read_models"
	carEventsTableName     = "core.car_events"
)

const (
	DefaultHostCheckInerval = 10 * time.Minute // can be tweaked respectively to the expected load
	DefaultMaxConns         = 50
	DefaultMaxConnLifetime  = "20s"
	DefaultMaxConnIdleTime  = "5s"
)

const (
	PGErrUniqueViolation = "23505"
	PGErrROTransaction   = "25006"
	PGErrBadEnumValue    = "22P02"
)

type SelectOption = func(*squirrel.SelectBuilder)

var (
	ForUpdateOption = func(sq *squirrel.SelectBuilder) {
		*sq = sq.Suffix("FOR UPDATE")
	}
)
