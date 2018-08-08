package openape

import (
	"fmt"

	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq" // used for db connection
	"github.com/spf13/viper"
)

// DatabaseConnect loads connection strings from the config file and connects to the specified DB
func DatabaseConnect() *xorm.Engine {
	engine, err := xorm.NewEngine(viper.GetString("database.type"), viper.GetString("database.conn"))
	if err != nil {
		panic(fmt.Errorf("Error connecting to database: %s", err))
	}
	return engine
}
