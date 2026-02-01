module github.com/1ncrease0/pixly/services/notification

go 1.25.5

require (
	github.com/1ncrease0/pixly/pkg/logger v0.0.0-20260118145235-7183e6521342
	github.com/ilyakaznacheev/cleanenv v1.5.0
	github.com/joho/godotenv v1.5.1
	github.com/rabbitmq/amqp091-go v1.10.0
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
)

replace (
	github.com/1ncrease0/pixly/pkg/logger => ../../pkg/logger
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)
