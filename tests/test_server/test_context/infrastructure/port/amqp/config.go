package amqp

import ddd "github.com/paulvitic/ddd-go"

type AmqpConfig struct {
	Username    string `json:"amqpUsername"`
	Password    string `json:"amqpPassword"`
	Queue       string `json:"amqpQueue"`
	Host        string
	Port        int
	VirtualHost string
}

func NewAmqpConfig() *AmqpConfig {
	return &AmqpConfig{}
}

func (c *AmqpConfig) OnInit() {
	config, err := ddd.Configuration[AmqpConfig]("configs/properties.json")
	if err != nil {
		panic(err)
	}
	*c = *config
}
