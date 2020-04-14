package main

import (
	"fmt"
)

type Postgres struct {
	User     string
	Password string
	Metadata map[string]interface{}
}
type Config struct {
	Postgres Postgres
}

func main() {
	fmt.Println("Hello, playground")

	config := Config{
		Postgres: Postgres{
			User:     "name",
			Password: "pass",
			Metadata: map[string]interface{}{"test": "test"},
		},
	}

	var c interface{}
	c = config

	// b, err := toml.Marshal(c)
	// if err != nil {

	// 	fmt.Printf("err:%s\n", err.Error())
	// 	os.Exit(0)
	// }
	bc := c.(Config)
	fmt.Printf("unmarshal: %v\n", bc)
	fmt.Printf("user=%s", config.Postgres.User)
}
