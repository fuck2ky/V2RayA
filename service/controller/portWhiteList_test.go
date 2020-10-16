package controller

import (
	"github.com/v2rayA/v2rayA/db/configure"
	"testing"
)

func TestCompressed(t *testing.T) {
	pwl := configure.PortWhiteList{
		TCP: []string{"2", "2", "5:10", "7", "8", "1", "-5", "64432", "65535", "65536"},
		UDP: []string{"3:65534", "65533:65534", "65533", "65534", "65535", "65536"},
	}
	t.Log(pwl.Compressed())
}
