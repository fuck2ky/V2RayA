package shadowsocksr

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
	"github.com/v2rayA/v2rayA/common/netTools/ports"
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"github.com/v2rayA/v2rayA/extra/proxy/socks5"
	"github.com/v2rayA/v2rayA/extra/proxy/ssr"
	"github.com/v2rayA/v2rayA/plugin"
)

type SSR struct {
	c         chan interface{}
	closed    chan interface{}
	localPort int
}
type Params struct {
	Cipher, Passwd, Address, Port, Obfs, ObfsParam, Protocol, ProtocolParam string
}

func init() {
	plugin.RegisterPlugin("ss", NewSSRPlugin)
	plugin.RegisterPlugin("ssr", NewSSRPlugin)
	plugin.RegisterPlugin("shadowsocks", NewSSRPlugin)
	plugin.RegisterPlugin("shadowsocksr", NewSSRPlugin)
}

func NewSSRPlugin(localPort int, v vmessInfo.VmessInfo) (plugin plugin.Plugin, err error) {
	plugin = new(SSR)
	err = plugin.Serve(localPort, v)
	return
}

func (self *SSR) Serve(localPort int, v vmessInfo.VmessInfo) (err error) {
	self.c = make(chan interface{}, 0)
	self.closed = make(chan interface{}, 0)
	self.localPort = localPort
	params := Params{
		Cipher:        v.Net,
		Passwd:        v.ID,
		Address:       v.Add,
		Port:          v.Port,
		Obfs:          v.TLS,
		ObfsParam:     v.Path,
		Protocol:      v.Type,
		ProtocolParam: v.Host,
	}
	u, err := url.Parse(fmt.Sprintf(
		"ssr://%v:%v@%v:%v",
		url.PathEscape(params.Cipher),
		url.PathEscape(params.Passwd),
		url.PathEscape(params.Address),
		url.PathEscape(params.Port),
	))
	if err != nil {
		log.Println(err)
		return
	}
	q := u.Query()
	if len(strings.TrimSpace(params.Obfs)) <= 0 {
		params.Obfs = "plain"
	}
	if len(strings.TrimSpace(params.Protocol)) <= 0 {
		params.Protocol = "origin"
	}
	q.Set("obfs", params.Obfs)
	q.Set("obfs_param", params.ObfsParam)
	q.Set("protocol", params.Protocol)
	q.Set("protocol_param", params.ProtocolParam)
	u.RawQuery = q.Encode()
	p, _ := ssr.NewProxy(u.String())
	local, err := socks5.NewSocks5Server("socks5://127.0.0.1:"+strconv.Itoa(localPort), p)
	if err != nil {
		return
	}
	go func() {
		go func() {
			e := local.ListenAndServe()
			if e != nil {
				err = e
			}
		}()
		<-self.c
		if local.(*socks5.Socks5).TcpListener != nil {
			close(self.closed)
			_ = local.(*socks5.Socks5).TcpListener.Close()
		}
	}()
	//等待100ms的error
	time.Sleep(100 * time.Millisecond)
	return err
}

func (self *SSR) Close() error {
	if self.c == nil {
		return newError("close fail: shadowsocksr not running")
	}
	if len(self.c) > 0 {
		return newError("close fail: duplicate close")
	}
	self.c <- nil
	self.c = nil
	time.Sleep(100 * time.Millisecond)
	start := time.Now()
	port := strconv.Itoa(self.localPort)
out:
	for {
		select {
		case <-self.closed:
			break out
		default:
		}
		var o bool
		o, _, err := ports.IsPortOccupied([]string{port + ":tcp"})
		if err != nil {
			return err
		}
		if !o {
			break
		}
		conn, e := net.Dial("tcp", ":"+port)
		if e == nil {
			conn.Close()
		}
		if time.Since(start) > 3*time.Second {
			log.Println("SSR.Close: timeout", self.localPort)
			return newError("SSR.Close: timeout")
		}
		time.Sleep(1000 * time.Millisecond)
	}
	return nil
}

func (self *SSR) SupportUDP() bool {
	return false
}
