module gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-http

go 1.26.1

require (
	github.com/go-viper/mapstructure/v2 v2.5.0
	github.com/google/uuid v1.6.0
	github.com/stretchr/testify v1.11.1
	gitlab.com/iobuilders/projects/eng/naryo-go/core v0.0.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gitlab.com/iobuilders/projects/eng/naryo-go/core => ../core
