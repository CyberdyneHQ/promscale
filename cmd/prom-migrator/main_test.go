// This file and its contents are licensed under the Apache License 2.0.
// Please see the included NOTICE for copyright information and
// LICENSE for a copy of the license.

package main

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/timescale/promscale/pkg/migration-tool/utils"
)

func TestParseFlags(t *testing.T) {
	cases := []struct {
		name            string
		input           []string
		expectedConf    *config
		failsValidation bool
		errMessage      string
	}{
		{
			name:  "pass_normal",
			input: []string{"-mint=1000", "-maxt=1001", "-read-url=http://localhost:9090/api/v1/read", "-write-url=http://localhost:9201/write", "-progress-enabled=false"},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               1000000,
				mintSec:            1000,
				maxt:               1001000,
				maxtSec:            1001,
				readURL:            "http://localhost:9090/api/v1/read",
				writeURL:           "http://localhost:9201/write",
				progressMetricName: "prom_migrator_progress",
				progressMetricURL:  "",
				maxBlockSize:       "500MB",
				numShards:          4,
				progressEnabled:    false,
			},
			failsValidation: false,
		},
		{
			name:  "pass_normal_size",
			input: []string{"-mint=1000", "-maxt=1001", "-read-url=http://localhost:9090/api/v1/read", "-write-url=http://localhost:9201/write", "-progress-enabled=false", "-max-read-size=100MB"},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               1000000,
				mintSec:            1000,
				maxt:               1001000,
				maxtSec:            1001,
				readURL:            "http://localhost:9090/api/v1/read",
				writeURL:           "http://localhost:9201/write",
				progressMetricName: "prom_migrator_progress",
				progressMetricURL:  "",
				maxBlockSize:       "100MB",
				numShards:          4,
				progressEnabled:    false,
			},
			failsValidation: false,
		},
		{
			name:  "pass_normal_size_with_space",
			input: []string{"-mint=1000", "-maxt=1001", "-read-url=http://localhost:9090/api/v1/read", "-write-url=http://localhost:9201/write", "-progress-enabled=false", "-max-read-size=100 MB"},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               1000000,
				mintSec:            1000,
				maxt:               1001000,
				maxtSec:            1001,
				readURL:            "http://localhost:9090/api/v1/read",
				writeURL:           "http://localhost:9201/write",
				progressMetricName: "prom_migrator_progress",
				progressMetricURL:  "",
				maxBlockSize:       "100 MB",
				numShards:          4,
				progressEnabled:    false,
			},
			failsValidation: false,
		},
		{
			name:  "fail_normal_size",
			input: []string{"-mint=1000", "-maxt=1001", "-read-url=http://localhost:9090/api/v1/read", "-write-url=http://localhost:9201/write", "-progress-enabled=false", "-max-read-size=100MBB"},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               1000000,
				mintSec:            1000,
				maxt:               1001000,
				maxtSec:            1001,
				readURL:            "http://localhost:9090/api/v1/read",
				writeURL:           "http://localhost:9201/write",
				progressMetricName: "prom_migrator_progress",
				progressMetricURL:  "",
				maxBlockSize:       "100MBB",
				numShards:          4,
				progressEnabled:    false,
			},
			failsValidation: true,
			errMessage:      `parsing byte-size: Unrecognized size suffix MBB`,
		},
		{
			name:  "fail_invalid_suffix",
			input: []string{"-mint=1000", "-maxt=1001", "-read-url=http://localhost:9090/api/v1/read", "-write-url=http://localhost:9201/write", "-progress-enabled=false", "-max-read-size=100PP"},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               1000000,
				mintSec:            1000,
				maxt:               1001000,
				maxtSec:            1001,
				readURL:            "http://localhost:9090/api/v1/read",
				writeURL:           "http://localhost:9201/write",
				progressMetricName: "prom_migrator_progress",
				progressMetricURL:  "",
				maxBlockSize:       "100PP",
				numShards:          4,
				progressEnabled:    false,
			},
			failsValidation: true,
			errMessage:      `parsing byte-size: Unrecognized size suffix PP`,
		},
		{
			name:  "pass_normal_regex",
			input: []string{"-mint=1000", "-maxt=1001", "-read-url=http://localhost:9090/api/v1/read", "-write-url=http://localhost:9201/write", "-progress-enabled=false", "-progress-metric-name=progress_migration_up"},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               1000000,
				mintSec:            1000,
				maxt:               1001000,
				maxtSec:            1001,
				readURL:            "http://localhost:9090/api/v1/read",
				writeURL:           "http://localhost:9201/write",
				progressMetricName: "progress_migration_up",
				progressMetricURL:  "",
				progressEnabled:    false,
				maxBlockSize:       "500MB",
				numShards:          4,
			},
			failsValidation: false,
		},
		{
			name:  "fail_invalid_regex",
			input: []string{"-mint=1000", "-maxt=1001", "-read-url=http://localhost:9090/api/v1/read", "-write-url=http://localhost:9201/write", "-progress-enabled=false", "-progress-metric-name=_progress_migration-_up"},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               1000000,
				mintSec:            1000,
				maxt:               1001000,
				maxtSec:            1001,
				readURL:            "http://localhost:9090/api/v1/read",
				writeURL:           "http://localhost:9201/write",
				progressMetricName: "_progress_migration-_up",
				progressMetricURL:  "",
				progressEnabled:    false,
				maxBlockSize:       "500MB",
				numShards:          4,
			},
			failsValidation: true,
			errMessage:      `invalid metric-name regex match: prom metric must match ^[a-zA-Z_:][a-zA-Z0-9_:]*$: recieved: _progress_migration-_up`,
		},
		{
			name:  "fail_invalid_regex",
			input: []string{"-mint=1000", "-maxt=1001", "-read-url=http://localhost:9090/api/v1/read", "-write-url=http://localhost:9201/write", "-progress-enabled=false", "-progress-metric-name=0_progress_migration_up"},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               1000000,
				mintSec:            1000,
				maxt:               1001000,
				maxtSec:            1001,
				readURL:            "http://localhost:9090/api/v1/read",
				writeURL:           "http://localhost:9201/write",
				progressMetricName: "0_progress_migration_up",
				progressMetricURL:  "",
				progressEnabled:    false,
				maxBlockSize:       "500MB",
				numShards:          4,
			},
			failsValidation: true,
			errMessage:      `invalid metric-name regex match: prom metric must match ^[a-zA-Z_:][a-zA-Z0-9_:]*$: recieved: 0_progress_migration_up`,
		},
		{
			name:  "fail_no_mint",
			input: []string{""},
			expectedConf: &config{
				name:               "prom-migrator",
				maxt:               time.Now().Unix() * 1000,
				maxtSec:            time.Now().Unix(),
				readURL:            "",
				writeURL:           "",
				progressMetricName: "prom_migrator_progress",
				progressMetricURL:  "",
				progressEnabled:    true,
				maxBlockSize:       "500MB",
				numShards:          4,
			},
			failsValidation: true,
			errMessage:      `mint should be provided for the migration to begin`,
		},
		{
			name:  "fail_all_default",
			input: []string{"-mint=1"},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               1000,
				mintSec:            1,
				maxt:               time.Now().Unix() * 1000,
				maxtSec:            time.Now().Unix(),
				readURL:            "",
				writeURL:           "",
				progressMetricName: "prom_migrator_progress",
				progressMetricURL:  "",
				progressEnabled:    true,
				maxBlockSize:       "500MB",
				numShards:          4,
			},
			failsValidation: true,
			errMessage:      `remote read storage url and remote write storage url must be specified. Without these, data migration cannot begin`,
		},
		{
			name:  "fail_all_default_space",
			input: []string{"-mint=1", "-read-url=  ", "-write-url= "},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               1000,
				mintSec:            1,
				maxt:               time.Now().Unix() * 1000,
				maxtSec:            time.Now().Unix(),
				readURL:            "  ",
				writeURL:           " ",
				progressMetricName: "prom_migrator_progress",
				progressMetricURL:  "",
				progressEnabled:    true,
				maxBlockSize:       "500MB",
				numShards:          4,
			},
			failsValidation: true,
			errMessage:      `remote read storage url and remote write storage url must be specified. Without these, data migration cannot begin`,
		},
		{
			name:  "fail_empty_read_url",
			input: []string{"-mint=1", "-write-url=http://localhost:9201/write"},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               1000,
				mintSec:            1,
				maxt:               time.Now().Unix() * 1000,
				maxtSec:            time.Now().Unix(),
				readURL:            "",
				writeURL:           "http://localhost:9201/write",
				progressMetricName: "prom_migrator_progress",
				progressMetricURL:  "",
				progressEnabled:    true,
				maxBlockSize:       "500MB",
				numShards:          4,
			},
			failsValidation: true,
			errMessage:      `remote read storage url needs to be specified. Without read storage url, data migration cannot begin`,
		},
		{
			name:  "fail_empty_write_url",
			input: []string{"-mint=1", "-read-url=http://localhost:9090/api/v1/read"},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               1000,
				mintSec:            1,
				maxt:               time.Now().Unix() * 1000,
				maxtSec:            time.Now().Unix(),
				readURL:            "http://localhost:9090/api/v1/read",
				writeURL:           "",
				progressMetricName: "prom_migrator_progress",
				progressMetricURL:  "",
				progressEnabled:    true,
				maxBlockSize:       "500MB",
				numShards:          4,
			},
			failsValidation: true,
			errMessage:      `remote write storage url needs to be specified. Without write storage url, data migration cannot begin`,
		},
		{
			name:  "fail_mint_greater_than_maxt",
			input: []string{"-mint=1000000001", "-maxt=1000000000", "-read-url=http://localhost:9090/api/v1/read", "-write-url=http://localhost:9201/write"},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               1000000001000,
				mintSec:            1000000001,
				maxt:               1000000000000,
				maxtSec:            1000000000,
				readURL:            "http://localhost:9090/api/v1/read",
				writeURL:           "http://localhost:9201/write",
				progressMetricName: "prom_migrator_progress",
				progressMetricURL:  "",
				progressEnabled:    true,
				maxBlockSize:       "500MB",
				numShards:          4,
			},
			failsValidation: true,
			errMessage:      `invalid input: minimum timestamp value (mint) cannot be greater than the maximum timestamp value (maxt)`,
		},
		{
			name:  "fail_progress_enabled_but_no_read_write_storage_url_provided",
			input: []string{"-mint=1000000000001", "-maxt=1000000000000", "-read-url=http://localhost:9090/api/v1/read", "-write-url=http://localhost:9201/write"},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               1000000000001000,
				mintSec:            1000000000001,
				maxt:               1000000000000000,
				maxtSec:            1000000000000,
				readURL:            "http://localhost:9090/api/v1/read",
				writeURL:           "http://localhost:9201/write",
				progressMetricName: "prom_migrator_progress",
				progressMetricURL:  "",
				progressEnabled:    true,
				maxBlockSize:       "500MB",
				numShards:          4,
			},
			failsValidation: true,
			errMessage:      `invalid input: minimum timestamp value (mint) cannot be greater than the maximum timestamp value (maxt)`,
		},
		{
			name:  "pass_progress_enabled_and_read_write_storage_url_provided",
			input: []string{"-mint=100000000000", "-maxt=1000000000000", "-read-url=http://localhost:9090/api/v1/read", "-write-url=http://localhost:9201/write", "-progress-metric-url=http://localhost:9201/read"},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               100000000000000,
				mintSec:            100000000000,
				maxt:               1000000000000000,
				maxtSec:            1000000000000,
				readURL:            "http://localhost:9090/api/v1/read",
				writeURL:           "http://localhost:9201/write",
				progressMetricName: "prom_migrator_progress",
				progressMetricURL:  "http://localhost:9201/read",
				progressEnabled:    true,
				maxBlockSize:       "500MB",
				numShards:          4,
			},
			failsValidation: false,
			errMessage:      `invalid input: minimum timestamp value (mint) cannot be greater than the maximum timestamp value (maxt)`,
		},
		// Mutual exclusive tests.
		{
			name:  "pass_normal_exclusive_password",
			input: []string{"-mint=1000", "-maxt=1001", "-read-url=http://localhost:9090/api/v1/read", "-write-url=http://localhost:9201/write", "-progress-enabled=false", "-read-auth-password=password"},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               1000000,
				mintSec:            1000,
				maxt:               1001000,
				maxtSec:            1001,
				readURL:            "http://localhost:9090/api/v1/read",
				writeURL:           "http://localhost:9201/write",
				progressMetricName: "prom_migrator_progress",
				progressMetricURL:  "",
				maxBlockSize:       "500MB",
				numShards:          4,
				progressEnabled:    false,
				readerAuth:         utils.Auth{Password: "password"},
			},
			failsValidation: false,
		},
		{
			name:  "pass_normal_exclusive_bearer_token",
			input: []string{"-mint=1000", "-maxt=1001", "-read-url=http://localhost:9090/api/v1/read", "-write-url=http://localhost:9201/write", "-progress-enabled=false", "-read-auth-bearer-token=token"},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               1000000,
				mintSec:            1000,
				maxt:               1001000,
				maxtSec:            1001,
				readURL:            "http://localhost:9090/api/v1/read",
				writeURL:           "http://localhost:9201/write",
				progressMetricName: "prom_migrator_progress",
				progressMetricURL:  "",
				maxBlockSize:       "500MB",
				numShards:          4,
				progressEnabled:    false,
				readerAuth:         utils.Auth{BearerToken: "token"},
			},
			failsValidation: false,
		},
		{
			name:  "fail_non_exclusive_bearer_token_and_password",
			input: []string{"-mint=1000", "-maxt=1001", "-read-url=http://localhost:9090/api/v1/read", "-write-url=http://localhost:9201/write", "-progress-enabled=false", "-read-auth-password=password", "-read-auth-bearer-token=token"},
			expectedConf: &config{
				name:               "prom-migrator",
				mint:               1000000,
				mintSec:            1000,
				maxt:               1001000,
				maxtSec:            1001,
				readURL:            "http://localhost:9090/api/v1/read",
				writeURL:           "http://localhost:9201/write",
				progressMetricName: "prom_migrator_progress",
				progressMetricURL:  "",
				maxBlockSize:       "500MB",
				numShards:          4,
				progressEnabled:    false,
				readerAuth:         utils.Auth{Password: "password", BearerToken: "token"},
			},
			failsValidation: true,
			errMessage:      `reader auth validation: at most one of basic_auth, bearer_token & bearer_token_file must be configured`,
		},
	}

	for _, c := range cases {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		config := new(config)
		parseFlags(config, c.input)
		convertSecFlagToMs(c.expectedConf)
		assert.Equal(t, c.expectedConf, config, fmt.Sprintf("parse-flags: %s", c.name))
		err := validateConf(config)
		if c.failsValidation {
			if err == nil {
				t.Fatalf(fmt.Sprintf("%s should have failed", c.name))
			}
			assert.Equal(t, c.errMessage, err.Error(), fmt.Sprintf("validation: %s", c.name))
		}
		if err != nil && !c.failsValidation {
			assert.NoError(t, err, c.name)
		}
	}
}
