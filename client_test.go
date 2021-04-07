package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	mock "gopkg.in/h2non/gentleman-mock.v2"
	"gopkg.in/h2non/gentleman.v2"
)

func TestClient_TerminateActiveConnection(t *testing.T) {
	defer mock.Disable()

	mock.New(`http://localhost`).
		Delete(`/api/v2/connections/SFTP_001`).
		Reply(200)

	mock.New(`http://localhost`).
		Delete(`/api/v2/connections/SFTP_002`).
		Reply(401)

	cli := gentleman.New()
	cli.URL(`http://localhost`)
	cli.Use(mock.Plugin)

	client := client{cli: cli}
	err := client.TerminateActiveConnection(`SFTP_001`)
	assert.Nil(t, err)

	err = client.TerminateActiveConnection(`SFTP_002`)
	assert.Equal(t, ErrUnauthorized, err)
}
