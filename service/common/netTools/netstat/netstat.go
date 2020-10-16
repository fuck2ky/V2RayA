package netstat

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"v2ray.com/core/common/errors"
)

// Socket states
type SkState uint8

const (
	pathNet  = "/proc/net"
	pathProc = "/proc"

	ipv4StrLen = 8
	ipv6StrLen = 32
)

const (
	Established SkState = 0x01
	SynSent             = 0x02
	SynRecv             = 0x03
	FinWait1            = 0x04
	FinWait2            = 0x05
	TimeWait            = 0x06
	Close               = 0x07
	CloseWait           = 0x08
	LastAck             = 0x09
	Listen              = 0x0a
	Closing             = 0x0b
)

var skStates = [...]string{
	"UNKNOWN",
	"ESTABLISHED",
	"SYN_SENT",
	"SYN_RECV",
	"FIN_WAIT1",
	"FIN_WAIT2",
	"TIME_WAIT",
	"", // CLOSE
	"CLOSE_WAIT",
	"LAST_ACK",
	"LISTEN",
	"CLOSING",
}

func (sk SkState) String() string {
	return skStates[sk]
}

type Address struct {
	IP   net.IP
	Port int
}

func parseIPv4(s string) (net.IP, error) {
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return nil, newError().Base(err)
	}
	ip := make(net.IP, net.IPv4len)
	binary.LittleEndian.PutUint32(ip, uint32(v))
	return ip, nil
}

func parseIPv6(s string) (net.IP, error) {
	ip := make(net.IP, net.IPv6len)
	const grpLen = 4
	i, j := 0, 4
	for len(s) != 0 {
		grp := s[0:8]
		u, err := strconv.ParseUint(grp, 16, 32)
		binary.LittleEndian.PutUint32(ip[i:j], uint32(u))
		if err != nil {
			return nil, newError().Base(err)
		}
		i, j = i+grpLen, j+grpLen
		s = s[8:]
	}
	return ip, nil
}

func parseAddr(s string) (*Address, error) {
	fields := strings.Split(s, ":")
	if len(fields) < 2 {
		return nil, fmt.Errorf("netstat: not enough fields: %v", s)
	}
	var ip net.IP
	var err error
	switch len(fields[0]) {
	case ipv4StrLen:
		ip, err = parseIPv4(fields[0])
	case ipv6StrLen:
		ip, err = parseIPv6(fields[0])
	default:
		return nil, newError("Bad formatted string")
	}
	if err != nil {
		return nil, newError().Base(err)
	}
	v, err := strconv.ParseUint(fields[1], 16, 16)
	if err != nil {
		return nil, newError().Base(err)
	}
	return &Address{IP: ip, Port: int(v)}, nil
}

type Socket struct {
	LocalAddress  *Address
	RemoteAddress *Address
	State         SkState
	UID           string
	inode         string
	Proc          *Process
	processMutex  sync.Mutex
}

type Process struct {
	PID  string
	Name string
}

const (
	SocketFreed    = "process not found, correspond socket was freed"
	ProcOpenFailed = "cannot open the directory /proc"
)

func FillProcesses(sockets []*Socket) error {
	f, err := ioutil.ReadDir(pathProc)
	if err != nil {
		return newError().Base(errors.New(ProcOpenFailed))
	}
	mapInodeSocket := make(map[string]*Socket)
	iNodes := make(map[string]struct{})
	for i, s := range sockets {
		s.processMutex.Lock()
		defer s.processMutex.Unlock()
		mapInodeSocket[s.inode] = sockets[i]
		iNodes[s.inode] = struct{}{}
	}
loop1:
	for _, fi := range f {
		fn := fi.Name()
		if !fi.IsDir() {
			continue
		}
		for _, t := range fn {
			if t > '9' || t < '0' {
				continue loop1
			}
		}
		for _, s := range sockets {
			if s.Proc != nil {
				continue
			}
		}
		if inode := isProcessSocket(fn, iNodes); inode != "" {
			mapInodeSocket[inode].Proc = &Process{
				PID:  fn,
				Name: getProcessName(fn),
			}
			delete(iNodes, inode)
		}
	}
	return nil
}

/*
较为消耗资源
*/
func (s *Socket) Process() (*Process, error) {
	s.processMutex.Lock()
	s.processMutex.Unlock()
	if s.Proc != nil {
		return s.Proc, nil
	}
	f, err := ioutil.ReadDir(pathProc)
	if err != nil {
		return nil, nil
	}
loop1:
	for _, fi := range f {
		fn := fi.Name()
		if !fi.IsDir() {
			continue
		}
		for _, t := range fn {
			if t > '9' || t < '0' {
				continue loop1
			}
		}
		if isProcessSocket(fn, map[string]struct{}{s.inode: {}}) != "" {
			s.Proc = &Process{
				PID:  fn,
				Name: getProcessName(fn),
			}
			return s.Proc, nil
		}
	}
	return nil, newError(SocketFreed)
}

/*
没有做缓存，每次调用都会扫描，消耗资源
*/

var ErrorNotFound = newError("process not found")

func findProcessID(pname string) (pid string, err error) {
	f, err := ioutil.ReadDir(pathProc)
	if err != nil {
		err = newError().Base(err)
		return
	}
loop1:
	for _, fi := range f {
		if !fi.IsDir() {
			continue
		}
		fn := fi.Name()
		for _, t := range fn {
			if t > '9' || t < '0' {
				continue loop1
			}
		}
		if getProcessName(fn) == pname {
			return fn, nil
		}
	}
	return "", ErrorNotFound
}

func getProcName(s string) string {
	i := strings.Index(s, "(")
	if i < 0 {
		return ""
	}
	s = s[i+1:]
	j := strings.LastIndex(s, ")")
	if j < 0 {
		return ""
	}
	return s[:j]
}

func getProcessName(pid string) (pn string) {
	p := filepath.Join(pathProc, pid, "stat")
	b, err := ioutil.ReadFile(p)
	if err != nil {
		err = newError().Base(err)
		return
	}
	sp := bytes.SplitN(b, []byte(" "), 3)
	pn = string(sp[1])
	return getProcName(pn)
}

func isProcessSocket(pid string, socketInode map[string]struct{}) string {
	// link name is of the form socket:[5860846]
	p := filepath.Join(pathProc, pid, "fd")
	f, err := os.Open(p)
	fns, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return ""
	}
	for _, fn := range fns {
		lk, err := os.Readlink(filepath.Join(p, fn))
		if err != nil {
			continue
		}
		for inode := range socketInode {
			target := "socket:[" + inode + "]"
			if lk == target {
				return inode
			}
		}
	}
	return ""
}

func getProcessSocketSet(pid string) (set []string) {
	// link name is of the form socket:[5860846]
	p := filepath.Join(pathProc, pid, "fd")
	f, err := os.Open(p)
	fns, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		err = newError().Base(err)
		return
	}
	for _, fn := range fns {
		lk, err := os.Readlink(filepath.Join(p, fn))
		if err != nil {
			continue
		}
		if strings.HasPrefix(lk, "socket:[") {
			set = append(set, lk[8:len(lk)-1])
		}
	}
	return
}

func parseSocktab(r io.Reader) (map[int][]*Socket, error) {
	br := bufio.NewScanner(r)
	tab := make(map[int][]*Socket)

	// Discard title
	br.Scan()

	for br.Scan() {
		var s Socket
		line := br.Text()
		// Skip comments
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		fields := strings.Fields(line)
		if len(fields) < 12 {
			return tab, fmt.Errorf("netstat: not enough fields: %v, %v", len(fields), fields)
		}
		addr, err := parseAddr(fields[1])
		if err != nil {
			return tab, err
		}
		s.LocalAddress = addr
		addr, err = parseAddr(fields[2])
		if err != nil {
			return tab, err
		}
		s.RemoteAddress = addr
		u, err := strconv.ParseUint(fields[3], 16, 8)
		if err != nil {
			err = newError().Base(err)
			return tab, err
		}
		s.State = SkState(u)
		s.UID = fields[7]
		s.inode = fields[9]
		tab[s.LocalAddress.Port] = append(tab[s.LocalAddress.Port], &s)
	}
	if br.Err() != nil {
		return nil, newError(br.Err())
	}
	return tab, nil
}
func ToPortMap(protocols []string) (map[string]map[int][]*Socket, error) {
	m := make(map[string]map[int][]*Socket)
	for _, proto := range protocols {
		switch proto {
		case "tcp", "tcp6", "udp", "udp6":
			b, err := os.Open(filepath.Join(pathNet, proto))
			if err != nil {
				continue
			}
			m[proto], err = parseSocktab(b)
			if err != nil {
				return nil, err
			}
		default:
		}
	}

	return m, nil
}

func IsProcessListenPort(pname string, port int) (is bool, err error) {
	protocols := []string{"tcp", "tcp6"}
	m, err := ToPortMap(protocols)
	if err != nil {
		return
	}
	iNodes := make(map[string]struct{})
	for _, proto := range protocols {
		for _, v := range m[proto][port] {
			if v.State == Listen || v.State == Established {
				iNodes[v.inode] = struct{}{}
			}
		}
	}
	if len(iNodes) == 0 {
		return false, nil
	}
	pid, err := findProcessID(pname)
	if err != nil {
		if errors.Cause(err) == ErrorNotFound {
			return false, nil
		}
		return
	}
	return isProcessSocket(pid, iNodes) != "", nil
}

func FillAllProcess(sockets []*Socket) {
	mInodeSocket := make(map[string]*Socket)
	for _, v := range sockets {
		if v.Proc == nil {
			mInodeSocket[v.inode] = v
			v.processMutex.Lock()
			defer v.processMutex.Unlock()
		}
	}
	f, err := ioutil.ReadDir(pathProc)
	if err != nil {
		return
	}
loop1:
	for _, fi := range f {
		if !fi.IsDir() {
			continue
		}
		fn := fi.Name()
		for _, t := range fn {
			if t > '9' || t < '0' {
				continue loop1
			}
		}
		socketSet := getProcessSocketSet(fn)
		for _, s := range socketSet {
			if socket, ok := mInodeSocket[s]; ok {
				socket.Proc = &Process{
					PID:  fn,
					Name: getProcessName(fn),
				}
			}
			delete(mInodeSocket, s)
		}
	}
}

func Print(protocols []string) string {
	var buffer strings.Builder
	protos := make([]string, 0, 4)
	for _, proto := range protocols {
		switch proto {
		case "tcp", "tcp6", "udp", "udp6":
			protos = append(protos, proto)
		}
	}
	m, err := ToPortMap(protos)
	if err != nil {
		return ""
	}
	buffer.WriteString(fmt.Sprintf("%-6v%-25v%-25v%-15v%-6v%-9v%v\n", "Proto", "Local Address", "Foreign Address", "State", "User", "Inode", "PID/Program name"))
	var sockets []*Socket
	for _, proto := range protos {
		for _, v := range m[proto] {
			sockets = append(sockets, v...)
		}
	}
	FillAllProcess(sockets)
	for _, proto := range protos {
		for _, sockets := range m[proto] {
			for _, v := range sockets {
				process, err := v.Process()
				var pstr string
				if err != nil {
					pstr = ""
				} else {
					pstr = process.PID + "/" + process.Name
				}
				buffer.WriteString(fmt.Sprintf(
					"%-6v%-25v%-25v%-15v%-6v%-9v%v\n",
					proto,
					v.LocalAddress.IP.String()+"/"+strconv.Itoa(v.LocalAddress.Port),
					v.RemoteAddress.IP.String()+"/"+strconv.Itoa(v.RemoteAddress.Port),
					v.State.String(),
					v.UID,
					v.inode,
					pstr,
				))
			}
		}
	}
	return buffer.String()
}
