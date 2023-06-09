package gopapi

import (
	"github.com/gurkankaymak/hocon"
	"log"
	"plugin"
	"regexp"
	"strconv"
	"strings"
)

const (
	SOURCE_TIME     string = "tsrc"
	PROCESS_ID      string = "pid"
	LOG_SOURCE      string = "src"
	LEVEL           string = "lvl"
	ENTRY           string = "msg"
	RECEIVE_TIME    string = "trcv"
	SOURCE_IP       string = "ip"
	SOURCE_APP_NAME string = "app"
)

const ( // log levels
	DEBUG int = iota
	INFO
	LOG
	WARN
	ERROR
	SEVERE
	PANIC
)

var fre = regexp.MustCompile(`^[a-zA-Z]\w+$`)

type LogEntry struct {
	Ip      string
	SrcTime string
	RcvTime string
	Pid     string
	Source  string
	Entry   string
	AppName string
	Level   int

	fields map[string]string
}

func (e *LogEntry) LevelName() string {
	switch e.Level {
	case DEBUG:
		return `DEBUG`
	case INFO:
		return `INFO`
	case LOG:
		return `LOG`
	case WARN:
		return `WARN`
	case ERROR:
		return `ERROR`
	case SEVERE:
		return `SEVERE`
	default:
		return `PANIC`
	}
}

func NewEntry() LogEntry {
	return LogEntry{fields: make(map[string]string, 0)}
}

func (e *LogEntry) FieldsNames() []string {
	keys := make([]string, len(e.fields))
	i := 0
	for k := range e.fields {
		keys[i] = k
		i++
	}

	return keys
}

func (e *LogEntry) GetField(name string) string {
	return e.fields[name]
}

func (e *LogEntry) SetField(nameRaw string, value string) {
	name := strings.TrimSpace(nameRaw)
	if len(name) == 0 {
		return
	}

	switch name {
	case SOURCE_IP:
		e.Ip = value
		break
	case LOG_SOURCE:
		e.Source = value
		break
	case SOURCE_TIME:
		e.SrcTime = value
		break
	case RECEIVE_TIME:
		e.RcvTime = value
		break
	case PROCESS_ID:
		e.Pid = value
		break
	case ENTRY:
		e.Entry = value
		break
	case SOURCE_APP_NAME:
		e.AppName = value
		break
	case LEVEL:
		switch strings.ToUpper(strings.TrimSpace(value)) {
		case "DEBUG":
			e.Level = DEBUG
			return
		case "INFO":
			e.Level = INFO
			return
		case "LOG":
			e.Level = LOG
			return
		case "WARN":
			e.Level = WARN
			return
		case "ERROR":
			e.Level = ERROR
			return
		case "SEVERE":
			e.Level = SEVERE
			return
		case "PANIC":
			e.Level = PANIC
			return
		}

		l, er := strconv.Atoi(strings.TrimSpace(value))

		if er == nil && l >= DEBUG && l <= PANIC {
			e.Level = l
			break
		}
	default:
		if !fre.MatchString(name) {
			log.Printf("Field name doesnt match convention: '%s'", name)
			return
		}
		e.fields[name] = value
	}
}

type ConnectionType struct {
	Tcp  bool   // true - tcp, false - udp
	Name string // unique name of the type
}

type SinkPlugin struct {
	configure      func(config *hocon.Config, consumer func(entry LogEntry))
	supportedTypes func() []ConnectionType

	chunk func(chunk []byte, source string, tcp bool) []byte
}

func (plug *SinkPlugin) Configure(config *hocon.Config, consumer func(entry LogEntry)) {
	plug.configure(config, consumer)
}

func (plug *SinkPlugin) SupportedTypes() []ConnectionType {
	return plug.supportedTypes()
}

func (plug *SinkPlugin) Chunk(chunk []byte, source string, tcp bool) []byte {
	return plug.chunk(chunk, source, tcp)
}

func (plug *SinkPlugin) Init(lookup func(symName string) (plugin.Symbol, error)) error {
	cfgSym, err := lookup("Configure")
	if err != nil {
		return err
	}

	typesSym, err := lookup("SupportedTypes")
	if err != nil {
		return err
	}

	chunkSym, err := lookup("Chunk")
	if err != nil {
		return err
	}

	plug.configure = func(config *hocon.Config, consumer func(entry LogEntry)) {
		cfgSym.(func(config *hocon.Config, consumer func(entry LogEntry)))(config, consumer)
	}

	plug.supportedTypes = func() []ConnectionType {
		return typesSym.(func() []ConnectionType)()
	}

	plug.chunk = func(chunk []byte, source string, tcp bool) []byte {
		return chunkSym.(func(chunk []byte, source string, tcp bool) []byte)(chunk, source, tcp)
	}

	return nil
}
