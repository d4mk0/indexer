package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/cyberconnecthq/indexer/fetcher"
)

func TestProcessTwitterSybilList(t *testing.T) {
	identity, err := fetcher.NewFetcher().FetchIdentity("0x90C311e132Db3BBDd2d419062cCcc25A88972ad9")

	assert.Nil(t, err)
	assert.Equal(t, identity.Twitter[0].Handle, "smamn27")
	assert.Equal(t, identity.Twitter[0].DataSource, fetcher.SYBIL)
}
