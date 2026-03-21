package handlers_test

import (
	"os"
	"testing"

	"github.com/surveyflow/be/internal/testutil"
)

func TestMain(m *testing.M) {
	testutil.Init()
	code := m.Run()
	testutil.Cleanup()
	os.Exit(code)
}
