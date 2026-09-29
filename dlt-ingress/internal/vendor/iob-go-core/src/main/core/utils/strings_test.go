package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoreStringUtils_ToSnakeCase(t *testing.T) {
	assert.Equal(t, "camel_case", ToSnakeCase("camelCase"))
	assert.Equal(t, "database_timeout", ToSnakeCase("DatabaseTimeout"))
	assert.Equal(t, "https_request", ToSnakeCase("HTTPSRequest"))
	assert.Equal(t, "get_https_url", ToSnakeCase("getHTTPSUrl"))
	assert.Equal(t, "my_http_server", ToSnakeCase("MyHTTPServer"))
}
