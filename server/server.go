package server

import (
	"fmt"
	_ "log"
	"net/http"
	"os"

	"github.com/dutchcoders/marija/server/datasources"

	"github.com/dutchcoders/marija/server/datasources/blockchain"
	"github.com/dutchcoders/marija/server/datasources/es5"
	"github.com/dutchcoders/marija/server/datasources/twitter"

	web "github.com/dutchcoders/marija-web"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/fatih/color"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("marija/server")

type Server struct {
	*config
}

type config struct {
	path    string
	address string
	debug   bool

	ListenerString string `toml:"listen"`

	Username string `toml:"username"`
	Password string `toml:"password"`
	Service  string `toml:"service"`

	Datasources Datasources `toml:"datasource"`

	Logging []struct {
		Output string `toml:"output"`
		Level  string `toml:"level"`
	} `toml:"logging"`
}

// type Datasource interface{}

type Datasources map[string]datasources.Index

func (d *Datasources) UnmarshalTOML(p interface{}) error {
	m := Datasources{}

	data, _ := p.(map[string]interface{})
	for n, v := range data {
		if d, ok := v.(map[string]interface{}); ok {
			if v, ok := d["type"]; !ok {
			} else if v, ok := v.(string); !ok {
			} else if v == "elasticsearch" {
				nd := &es5.Elasticsearch{}
				if err := nd.UnmarshalTOML(d); err != nil {
					return err
				}
				m[n] = nd
			} else if v == "twitter" {
				nd := &twitter.Twitter{}
				if err := nd.UnmarshalTOML(d); err != nil {
					return err
				}
				m[n] = nd
			} else if v == "blockchain" {
				nd := &btc.BTC{}
				if err := nd.UnmarshalTOML(d); err != nil {
					return err
				}
				m[n] = nd
			} else {
			}
		} else {
			return fmt.Errorf("not a dish")
		}
	}

	*d = m

	return nil
}

func (server *Server) Run() {
	go h.run()

	/*
		fileHandler := http.FileServer(http.Dir(path.Join(dir, "static")))
		http.Handle("/static/", fileHandler)
	*/

	staticHandler := http.FileServer(
		&assetfs.AssetFS{
			Asset:    web.Asset,
			AssetDir: web.AssetDir,
			AssetInfo: func(path string) (os.FileInfo, error) {
				return os.Stat(path)
			},
			Prefix: web.Prefix,
		})

	if server.path != "" {
		log.Debug("Using static file path: ", server.path)

		// check local css first
		staticHandler = http.FileServer(http.Dir(server.path))
	}

	http.Handle("/", staticHandler)

	http.HandleFunc("/ws", server.serveWs)

	fmt.Println(color.YellowString(fmt.Sprintf("Marija server started, listening on address %s.", server.address)))

	defer func() {
		fmt.Println(color.YellowString(fmt.Sprintf("Marija server stopped.")))
	}()

	if err := http.ListenAndServe(server.address, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func New(options ...func(*Server)) *Server {
	server := &Server{
		config: &config{
			debug: false,
		},
	}

	for _, optionFn := range options {
		optionFn(server)
	}

	return server
}
