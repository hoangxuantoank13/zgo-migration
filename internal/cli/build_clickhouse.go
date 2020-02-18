// +build clickhouse

package cli

import (
	_ "github.com/ClickHouse/clickhouse-go"
	_ "github.com/hoangxuantoank13/zgo-migration/database/clickhouse"
)
