/*
Copyright 2021 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package statechange

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dapr/components-contrib/bindings"
	"github.com/dapr/components-contrib/metadata"
	"github.com/dapr/kit/logger"
)

func getTestMetadata() map[string]string {
	return map[string]string{
		"address":  "127.0.0.1:28015",
		"database": "dapr",
		"username": "admin",
		"password": "rethinkdb",
		"table":    "daprstate",
	}
}

func getNewRethinkActorBinding() *Binding {
	l := logger.NewLogger("test")
	if os.Getenv("DEBUG") != "" {
		l.SetOutputLevel(logger.DebugLevel)
	}

	return NewRethinkDBStateChangeBinding(l).(*Binding)
}

/*
go test github.com/dapr/components-contrib/bindings/rethinkdb/statechange \
	-run ^TestBinding$ -count 1
*/

func TestBinding(t *testing.T) {
	if os.Getenv("RUN_LIVE_RETHINKDB_TEST") != "true" {
		t.SkipNow()
	}
	testDuration := 10 * time.Second
	testDurationStr := os.Getenv("RETHINKDB_TEST_DURATION")
	if testDurationStr != "" {
		d, err := time.ParseDuration(testDurationStr)
		if err != nil {
			t.Fatalf("invalid test duration: %s, expected time.Duration", testDurationStr)
		}
		testDuration = d
	}

	m := bindings.Metadata{Base: metadata.Base{
		Name:       "test",
		Properties: getTestMetadata(),
	}}

	b := getNewRethinkActorBinding()
	err := b.Init(t.Context(), m)
	require.NoErrorf(t, err, "error initializing")

	ctx, cancel := context.WithCancel(t.Context())
	err = b.Read(ctx, func(_ context.Context, res *bindings.ReadResponse) ([]byte, error) {
		assert.NotNil(t, res)
		t.Logf("state change event:\n%s", string(res.Data))

		return nil, nil
	})
	require.NoErrorf(t, err, "error on read")

	testTimer := time.AfterFunc(testDuration, func() {
		t.Log("done")
		cancel()
	})
	defer testTimer.Stop()
}
