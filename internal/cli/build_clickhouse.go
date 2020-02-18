// +build clickhouse

package cli

import (
	_ "github.com/ClickHouse/clickhouse-go"
	_ "github.com/zgo-migration/migrate/database/clickhouse"
)
