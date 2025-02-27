package db_test

import (
	"testing"

	"flamingo.me/flamingo/v3/framework/config"

	"github.com/osdir/flamingo-mysql/db"
)

func TestModule_Configure(t *testing.T) {
	if err := config.TryModules(nil, new(db.Module)); err != nil {
		t.Error(err)
	}
}
