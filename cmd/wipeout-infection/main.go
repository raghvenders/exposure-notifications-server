// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This package is the service that deletes old infection keys; it is intended to be invoked over HTTP by Cloud Scheduler.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"cambio/pkg/api"
	"cambio/pkg/database"
	"cambio/pkg/logging"
)

const (
	timeoutEnvVar  = "WIPEOUT_TIMEOUT"
	defaultTimeout = 10 * time.Minute
)

func main() {
	ctx := context.Background()
	logger := logging.FromContext(ctx)

	timeout := defaultTimeout
	if timeoutStr := os.Getenv(timeoutEnvVar); timeoutStr != "" {
		var err error
		timeout, err = time.ParseDuration(timeoutStr)
		if err != nil {
			logger.Warnf("Failed to parse $%s value %q, using default.", timeoutEnvVar, timeoutStr)
			timeout = defaultTimeout
		}
	}
	logger.Infof("Using timeout %v (override with $%s)", timeout, timeoutEnvVar)

	cleanup, err := database.Initialize(ctx)
	if err != nil {
		logger.Fatalf("unable to connect to database: %v", err)
	}
	defer cleanup(ctx)

	http.Handle("/", api.InfectionWipeoutHandler{Timeout: timeout})
	logger.Info("starting wipeout server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}