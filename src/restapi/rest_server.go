package restapi

import (
	"bytes"
	"context"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"archive/zip"
	"time"

	"gerace.dev/zipfs"
	"golang.org/x/exp/slices"

	"github.com/RiV-chain/RiV-mesh/src/config"
	"github.com/RiV-chain/RiV-mesh/src/core"
	"github.com/RiV-chain/RiV-mesh/src/defaults"
	"github.com/RiV-chain/RiV-mesh/src/multicast"
	"github.com/RiV-chain/RiV-mesh/src/version"
	"github.com/ip2location/ip2location-go/v9"
	"github.com/slonm/tableprinter"
)

//	@title			RiV-mesh API
//	@version		0.1
//	@description	This is RiV-mesh client API documentation.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Development team
//	@contact.url	https://github.com/RiV-chain/RiV-mesh
//	@contact.email	support@rivchain.org

//	@license.name	LGPL3
//	@license.url	https://github.com/RiV-chain/RiV-mesh/blob/develop/LICENSE

//	@host		localhost:19019
//	@BasePath	/api

var _ embed.FS

//go:embed IP2LOCATION-LITE-DB1.BIN
var IP2LOCATION []byte

const ip2loc_not_supported string = "This parameter is unavailable for selected data file. Please upgrade the data file."
const ip2loc_invalid_ip_address string = "Invalid IP address."

type ServerEvent struct {
	Event string
	Data  []byte
}

type ApiHandler struct {
	Method  string `json:"method"`
	Pattern string `json:"pattern"` // Context path pattern
	Desc    string `json:"desc"`    // What does the endpoint do?
	//	args    []string            // List of human-readable argument names
	Handler func(w http.ResponseWriter, r *http.Request) // First is input map, second is output
}

type RestServerCfg struct {
	Core          *core.Core
	Multicast     *multicast.Multicast
	Log           core.Logger
	ListenAddress string
	WwwRoot       string
	ConfigFn      string
	handlers      []ApiHandler
	Domain        string
	Features      []string
}

type RestServer struct {
	server http.Server
	RestServerCfg
	listenUrl         *url.URL
	serverEvents      chan ServerEvent
	serverEventNextId int
	updateTimer       *time.Timer
	docFsType         string
	ip2locatinoDb     *ip2location.DB
}

func NewRestServer(cfg RestServerCfg) (*RestServer, error) {
	a := &RestServer{
		server:            http.Server{},
		RestServerCfg:     cfg,
		serverEvents:      make(chan ServerEvent, 10),
		serverEventNextId: 0,
	}
	if cfg.ListenAddress == "none" || cfg.ListenAddress == "" {
		return nil, errors.New("listening address isn't configured")
	}

	var err error
	a.listenUrl, err = url.Parse(cfg.ListenAddress)
	if err != nil {
		return nil, fmt.Errorf("an error occurred parsing http address: %w", err)
	}

	pakReader, err := zip.OpenReader(cfg.WwwRoot)

	//unregister http Handlers here
	http.DefaultServeMux = new(http.ServeMux)
	if err == nil {
		defer pakReader.Close()
		fs, err := zipfs.NewZipFileSystem(&pakReader.Reader, zipfs.ServeIndexForMissing())
		if err == nil {
			http.Handle("/", http.FileServer(fs))
			a.docFsType = "on zipfs"
		}
	}
	if a.docFsType == "" {
		a.docFsType = "not found"
		if _, err := os.Stat(cfg.WwwRoot); err == nil {
			var nocache = func(fs http.Handler) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					addNoCacheHeaders(w)
					fs.ServeHTTP(w, r)
				}
			}
			http.Handle("/", nocache(http.FileServer(http.Dir(cfg.WwwRoot))))
			a.docFsType = "on OS fs"
		} else {
			a.Log.Warnln("Document root get stat error: ", err)
		}
	}

	a.AddHandler(ApiHandler{Method: "GET", Pattern: "/api", Desc: "API documentation", Handler: a.getApiHandler})
	a.AddHandler(ApiHandler{Method: "GET", Pattern: "/api/self", Desc: "Show details about this node", Handler: a.getApiSelfHandler})
	a.AddHandler(ApiHandler{Method: "GET", Pattern: "/api/nodeinfo", Desc: "Request nodeinfo of this node", Handler: a.getApiNodeinfoHandler})
	a.AddHandler(ApiHandler{Method: "PUT", Pattern: "/api/nodeinfo", Desc: "Update nodeinfo of this node", Handler: a.putApiNodeinfoHandler})
	a.AddHandler(ApiHandler{Method: "GET", Pattern: "/api/peers", Desc: `Show directly connected peers`, Handler: a.getApiPeersHandler})
	a.AddHandler(ApiHandler{Method: "POST", Pattern: "/api/peers", Desc: `Append peers to the peers list. 
Request body [{ "uri":"tcp://xxx.xxx.xxx.xxx:yyyy", "interface":"eth0" }, ...], interface is optional
Request header "Riv-Save-Config: true" persists changes`, Handler: a.postApiPeersHandler})
	a.AddHandler(ApiHandler{Method: "PUT", Pattern: "/api/peers", Desc: `Set peers list. 
Request body [{ "uri":"tcp://xxx.xxx.xxx.xxx:yyyy", "interface":"eth0" }, ...], interface is optional.
Request header "Riv-Save-Config: true" persists changes`, Handler: a.putApiPeersHandler})
	a.AddHandler(ApiHandler{Method: "DELETE", Pattern: "/api/peers", Desc: `Remove all peers from this node
Request header "Riv-Save-Config: true" persists changes`, Handler: a.deleteApiPeersHandler})
	a.AddHandler(ApiHandler{Method: "GET", Pattern: "/api/publicpeers", Desc: "Show public peers loaded from URL which configured in mesh.conf file", Handler: a.getApiPublicPeersHandler})
	a.AddHandler(ApiHandler{Method: "GET", Pattern: "/api/paths", Desc: "Show established paths through this node", Handler: a.getApiPathsHandler})
	a.AddHandler(ApiHandler{Method: "POST", Pattern: "/api/health", Desc: "Run peers health check task", Handler: a.postApiHealthHandler})
	a.AddHandler(ApiHandler{Method: "GET", Pattern: "/api/sse", Desc: "Return server side events", Handler: a.getApiSseHandler})
	a.AddHandler(ApiHandler{Method: "GET", Pattern: "/api/dht", Desc: "Show known DHT entries", Handler: a.getApiDhtHandler})
	a.AddHandler(ApiHandler{Method: "GET", Pattern: "/api/sessions", Desc: "Show established traffic sessions with remote nodes", Handler: a.getApiSessionsHandler})
	a.AddHandler(ApiHandler{Method: "GET", Pattern: "/api/multicastinterfaces", Desc: "Show which interfaces multicast is enabled on", Handler: a.getApiMulticastinterfacesHandler})
	a.AddHandler(ApiHandler{Method: "GET", Pattern: "/api/remote/nodeinfo/{key}", Desc: "Request nodeinfo from a remote node by its public key", Handler: a.getApiRemoteNodeinfoHandler})
	a.AddHandler(ApiHandler{Method: "GET", Pattern: "/api/remote/self/{key}", Desc: "Request self from a remote node by its public key", Handler: a.getApiRemoteSelfHandler})
	a.AddHandler(ApiHandler{Method: "GET", Pattern: "/api/remote/peers/{key}", Desc: "Request peers from a remote node by its public key", Handler: a.getApiRemotePeersHandler})
	a.AddHandler(ApiHandler{Method: "GET", Pattern: "/api/remote/dht/{key}", Desc: "Request dht from a remote node by its public key", Handler: a.getApiRemoteDHTHandler})

	var _ = a.Core.PeersChangedSignal.Connect(func(data any) {
		b, err := json.Marshal(a.prepareGetPeers())
		if err != nil {
			a.Log.Errorf("get peers failed: %w", err)
			return
		}

		select {
		case a.serverEvents <- ServerEvent{Event: "peers", Data: b}:
		default:
		}
	})

	a.ip2locatinoDb, err = ip2location.OpenDBWithReader(nopCloser{bytes.NewReader(IP2LOCATION)})
	if err != nil {
		a.Log.Errorf("load ip2location DB failed: %w", err)
	}
	return a, nil
}

type nopCloser struct {
	*bytes.Reader
}

func (nopCloser) Close() error { return nil }

// Start http server
func (a *RestServer) Serve() error {
	sort.SliceStable(a.handlers, func(i, j int) bool {
		if len(a.handlers[i].Pattern) != len(a.handlers[j].Pattern) {
			return len(a.handlers[i].Pattern) < len(a.handlers[j].Pattern)
		}
		return strings.Compare(a.handlers[i].Pattern, a.handlers[j].Pattern) < 0
	})
	go func() {
		a.Log.Infof("Starting http server listening on %s. Document root %s %s\n", a.ListenAddress, a.WwwRoot, a.docFsType)
		localIp, err := net.LookupIP(a.listenUrl.Hostname())
		if err != nil {
			a.Log.Errorln(err)
			return
		}

		a.server.Addr = net.JoinHostPort(localIp[0].String(), a.listenUrl.Port())
		err = a.server.ListenAndServe()
		if err != nil {
			a.Log.Errorln(err)
		}
	}()
	return nil
}

// Shutdown http server
func (a *RestServer) Shutdown() error {
	err := a.server.Shutdown(context.Background())
	a.Log.Infof("Stop REST service")
	return err
}

// AddHandler is called for each admin function to add the handler and help documentation to the API.
func (a *RestServer) AddHandler(handler ApiHandler) {
	if idx := slices.IndexFunc(a.handlers, func(h ApiHandler) bool {
		return h.Method == handler.Method && h.Pattern == handler.Pattern
	}); idx >= 0 {
		a.Log.Debugln("handler " + handler.Pattern + " already exists")
		return
	}
	notRegistered := slices.IndexFunc(a.handlers, func(h ApiHandler) bool {
		return h.Pattern == handler.Pattern
	}) < 0
	a.handlers = append(a.handlers, handler)
	matchPattern := func(pattern string, value string) bool {
		p := strings.Split(pattern, "{")[0]
		return len(value) >= len(p) && p == value[:len(p)]
	}

	if notRegistered {
		http.HandleFunc(strings.Split(handler.Pattern, "{")[0], func(w http.ResponseWriter, r *http.Request) {
			{
				clientIp, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					http.Error(w, err.Error(), http.StatusForbidden)
					return
				}
				svrIp, _, err := net.SplitHostPort(a.server.Addr)
				if err != nil {
					http.Error(w, err.Error(), http.StatusForbidden)
					return
				}
				if clientIp != svrIp {
					http.Error(w, fmt.Sprintf("Forbidden access to '%s' from '%s'", svrIp, clientIp), http.StatusForbidden)
					return
				}
			}
			for i := range a.handlers {
				h := &a.handlers[len(a.handlers)-i-1]
				if h.Method == r.Method && matchPattern(h.Pattern, r.URL.Path) {
					//webauth module here
					for k, v := range r.Header {
						os.Setenv("HTTP_"+strings.ReplaceAll(strings.ToUpper(k), "-", "_"), strings.Join(v, ""))
						a.Log.Debugln("HTTP_" + strings.ReplaceAll(strings.ToUpper(k), "-", "_") + ":" + strings.Join(v, ""))
					}
					os.Setenv("REQUEST_METHOD", r.Method)
					os.Setenv("REQUEST_PATH", r.URL.Path)
					os.Setenv("QUERY_STRING", r.URL.RawQuery)
					os.Setenv("REMOTE_ADDR", r.RemoteAddr)
					os.Setenv("REMOTE_HOST", r.RemoteAddr)
					os.Setenv("SERVER_ADDR", r.Host)
					os.Setenv("SERVER_PROTOCOL", "HTTP/1.1")
					webauth := filepath.Join(filepath.Dir(a.WwwRoot), "var", "lib", "mesh", "hooks", "webauth")
					if _, err := os.Stat(webauth); err == nil {
						cmd := exec.Command(webauth)
						if err := cmd.Start(); err != nil {
							http.Error(w, err.Error(), http.StatusInternalServerError)
							return
						}
						if err := cmd.Wait(); err != nil {
							if exiterr, ok := err.(*exec.ExitError); ok {
								exitCode := exiterr.ExitCode()
								a.Log.Debugln("Auth failed. Exit code: ", exitCode)
								http.Error(w, "Authentication failed", http.StatusUnauthorized)
								return
							} else {
								http.Error(w, err.Error(), http.StatusInternalServerError)
								return
							}
						} else {
							a.Log.Debugln("Auth success")
						}
					} else {
						a.Log.Debugln("Auth module not found: ", webauth)
					}

					addNoCacheHeaders(w)
					h.Handler(w, r)
					return
				}
			}
			WriteError(w, http.StatusMethodNotAllowed)
		})
	}
}

func addNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Add("Pragma", "no-cache")
	w.Header().Add("Expires", "0")
}

func (a *RestServer) getApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.URL.Query().Has("fmt") && r.URL.Query()["fmt"][0] == "table" {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "Common query params: fmt=table|json - Response format\n\n")
		WriteJson(w, r, a.handlers)
		// for _, h := range a.handlers {
		// 	fmt.Fprintf(w, "%s %s\t\t%s\n\n", h.method, h.pattern, h.desc)
		// }
	} else {
		paths := map[string]map[string]map[string]string{}
		for _, h := range a.handlers {
			if _, exists := paths[h.Pattern]; !exists {
				paths[h.Pattern] = map[string]map[string]string{}
			}
			if _, exists := paths[h.Pattern][strings.ToLower(h.Method)]; !exists {
				paths[h.Pattern][strings.ToLower(h.Method)] = map[string]string{}
			}
			paths[h.Pattern][strings.ToLower(h.Method)]["description"] = h.Desc
		}
		swag := map[string]any{
			"openapi": "3.0.3",
			"info": map[string]string{
				"title":       "Riv mesh - OpenAPI 3.0",
				"description": "Common query params: fmt=table|json - response format",
			},
			"paths": paths,
		}
		b, err := json.Marshal(swag)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, string(b))
	}
}

// @Summary		Show details about this node. The output contains following fields: build name, build version, public key, private key, address, subnet, coords, features.
// @Produce		json
// @Success		200		{string}	string		"ok"
// @Failure		400		{error}		error		"Method not allowed"
// @Failure		401		{error}		error		"Authentication failed"
// @Router		/self [get]
func (a *RestServer) getApiSelfHandler(w http.ResponseWriter, r *http.Request) {
	self := a.Core.GetSelf()
	snet := a.Core.Subnet()
	var result = map[string]any{
		"build_name":    version.BuildName(),
		"build_version": version.BuildVersion(),
		"key":           hex.EncodeToString(self.Domain.Key[:]),
		"domain":        string(self.Domain.Name[:]),
		"private_key":   hex.EncodeToString(self.PrivateKey[:]),
		"address":       a.Core.Address().String(),
		"subnet":        snet.String(),
		"coords":        self.Coords,
		"features":      a.Features,
	}
	WriteJson(w, r, result)
}

// @Summary		Show node info of this node.
// @Produce		json
// @Success		200		{string}	string		"ok"
// @Failure		401		{error}		error		"Authentication failed"
// @Router		/nodeinfo [get]
func (a *RestServer) getApiNodeinfoHandler(w http.ResponseWriter, r *http.Request) {
	WriteJson(w, r, a.Core.GetThisNodeInfo())
}

// @Summary		Replace node info of this node.
// @Produce		json
// @Success		204		{string}	string		"No content"
// @Failure		400		{error}		error		"Bad request"
// @Failure		401		{error}		error		"Authentication failed"
// @Failure		500		{error}		error		"Internal error"
// @Router		/nodeinfo [put]
func (a *RestServer) putApiNodeinfoHandler(w http.ResponseWriter, r *http.Request) {
	var info core.NodeInfo
	err := json.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = a.Core.SetThisNodeInfo(info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	a.saveConfig(func(cfg *config.NodeConfig) {
		cfg.NodeInfo = info
	}, r)
}

// @Summary		Show known DHT entries. The output contains following fields: Address, Public Key, Port, Rest
// @Produce		json
// @Success		200		{string}	string		"ok"
// @Failure		400		{error}		error		"Method not allowed"
// @Failure		401		{error}		error		"Authentication failed"
// @Router		/dht [get]
func (a *RestServer) getApiDhtHandler(w http.ResponseWriter, r *http.Request) {
	dht := a.Core.GetDHT()
	result := make([]map[string]any, 0, len(dht))
	for _, d := range dht {
		addr := a.Core.AddrForDomain(d.Domain)
		entry := map[string]any{
			"address": net.IP(addr[:]).String(),
			"key":     hex.EncodeToString(d.Domain.Key),
			"Domain":  hex.EncodeToString(d.Domain.Name),
			"port":    d.Port,
			"rest":    d.Rest,
		}
		result = append(result, entry)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return strings.Compare(result[i]["key"].(string), result[j]["key"].(string)) < 0
	})
	WriteJson(w, r, result)
}

// @Summary		Show public peers which is result of output PublicPeersURL in mesh.conf.
// @Produce		json
// @Success		200		{string}	string		"ok"
// @Failure		400		{error}		error		"Method not allowed"
// @Failure		401		{error}		error		"Authentication failed"
// @Failure		500		{error}		error		"Internal server error"
// @Router		/publicpeers [get]
func (a *RestServer) getApiPublicPeersHandler(w http.ResponseWriter, r *http.Request) {
	var response *http.Response
	var result []byte
	cfg, err := defaults.ReadConfig(a.ConfigFn)
	if err == nil {
		u := cfg.PublicPeersUrl
		response, err = http.Get(u)
		if err != nil {
			a.Log.Errorln("Error read public peers url:", u, " ", err)
			http.Error(w, "Error read public peers url", http.StatusInternalServerError)
			return
		}
		if response.StatusCode > 200 {
			a.Log.Errorln("Error read public peers url. Response code: ", response.StatusCode, ", Error message: ", response.Status)
			WriteError(w, response.StatusCode)
			return
		}
		result, err = io.ReadAll(response.Body)
		if err != nil {
			a.Log.Errorln("Error read public peers url:", u, " ", err)
			http.Error(w, "Error read public peers url", http.StatusInternalServerError)
			return
		}
	} else {
		a.Log.Errorln("Config file read error:", err)
		http.Error(w, "Error read public peers url", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, string(result))
}

// @Summary		Show established paths through this node. The output contains following fields: Public Key, Path
// @Produce		json
// @Success		200		{string}	string		"ok"
// @Failure		401		{error}		error		"Authentication failed"
// @Failure		400		{error}		error		"Method not allowed"
// @Router		/paths [get]
func (a *RestServer) getApiPathsHandler(w http.ResponseWriter, r *http.Request) {
	paths := a.Core.GetPaths()
	result := make([]map[string]any, 0, len(paths))
	for _, d := range paths {
		entry := map[string]any{
			"key":  hex.EncodeToString(d.Key),
			"path": d.Path,
		}
		result = append(result, entry)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return strings.Compare(result[i]["key"].(string), result[j]["key"].(string)) < 0
	})
	WriteJson(w, r, result)
}

// @Summary		Show established traffic sessions with remote nodes. The output contains following fields: Address, Byte received, Byte sent, Public Key, Uptime
// @Produce		json
// @Success		200		{string}	string		"ok"
// @Failure		400		{error}		error		"Method not allowed"
// @Failure		401		{error}		error		"Authentication failed"
// @Router		/sessions [get]
func (a *RestServer) getApiSessionsHandler(w http.ResponseWriter, r *http.Request) {
	sessions := a.Core.GetSessions()
	result := make([]map[string]any, 0, len(sessions))
	for _, s := range sessions {
		entry := map[string]any{
			"key":         s.Key,
			"domain":      s.Domain,
			"bytes_recvd": s.RXBytes,
			"bytes_sent":  s.TXBytes,
			"uptime":      s.Uptime.Seconds(),
		}
		result = append(result, entry)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return strings.Compare(result[i]["key"].(string), result[j]["key"].(string)) < 0
	})
	WriteJson(w, r, result)
}

// @Summary		Show which interfaces multicast is enabled on.
// @Produce		json
// @Success		200		{string}	string		"ok"
// @Failure		400		{error}		error		"Method not allowed"
// @Failure		401		{error}		error		"Authentication failed"
// @Failure		500		{error}		error		"Internal server error"
// @Router		/multicastinterfaces [get]
func (a *RestServer) getApiMulticastinterfacesHandler(w http.ResponseWriter, r *http.Request) {
	if a.Multicast == nil {
		http.Error(w, "Multicast module isn't started", http.StatusInternalServerError)
		return
	}
	res := []string{}
	for _, v := range a.Multicast.Interfaces() {
		res = append(res, v.Name)
	}
	WriteJson(w, r, res)
}

type Peer struct {
	Address       string   `json:"address"`
	Key           string   `json:"key"`
	Domain        string   `json:"domain"`
	Port          uint64   `json:"port"`
	Priority      uint64   `json:"priority"`
	Coords        []uint64 `json:"coords"`
	Remote        string   `json:"remote"`
	Remote_ip     string   `json:"remote_ip"`
	Bytes_recvd   uint64   `json:"bytes_recvd"`
	Bytes_sent    uint64   `json:"bytes_sent"`
	Uptime        float64  `json:"uptime"`
	Multicast     bool     `json:"multicast"`
	Country_short string   `json:"country_short"`
	Country_long  string   `json:"country_long"`
}

func (a *RestServer) prepareGetPeers() []Peer {
	peers := a.Core.GetPeers()
	response := make([]Peer, 0, len(peers))
	for _, p := range peers {
		addr := a.Core.AddrForDomain(p.Domain)
		entry := Peer{
			net.IP(addr[:]).String(),
			hex.EncodeToString(p.Domain.Key),
			string(p.Domain.Name),
			p.Port,
			uint64(p.Priority), // can't be uint8 thanks to gobind
			p.Coords,
			p.Remote,
			p.RemoteIp,
			p.RXBytes,
			p.TXBytes,
			p.Uptime.Seconds(),
			strings.Contains(p.Remote, "[fe80::"),
			"",
			"",
		}

		entry.Country_short, entry.Country_long = a.getCountry(p.RemoteIp)

		response = append(response, entry)
	}
	sort.Slice(response, func(i, j int) bool {
		if response[i].Multicast != response[j].Multicast {
			return (!response[i].Multicast && response[j].Multicast)
		}

		if response[i].Priority != response[j].Priority {
			return response[i].Priority < response[j].Priority
		}

		if cmp := strings.Compare(response[i].Address, response[j].Address); cmp != 0 {
			return cmp < 0
		}
		return response[i].Port < response[j].Port
	})
	return response
}

// @Summary		Get current peers list. The output contains following fields: address, public key, port, priority, coordinates, remote URL, remote IP, bytes received, bytes sent, uptime, multicast flag, country code, country.
// @Produce		json
// @Success		200		{string}	string		"ok"
// @Failure		401		{error}		error		"Authentication failed"
// @Failure		403		{error}		error		"Bad request"
// @Router		/peers [get]
func (a *RestServer) getApiPeersHandler(w http.ResponseWriter, r *http.Request) {
	WriteJson(w, r, a.prepareGetPeers())
}

// @Summary		Add new peers.
// @Produce		json
// @Success		200		{string}	string		"ok"
// @Failure		401		{error}		error		"Authentication failed"
// @Failure		403		{error}		error		"Bad request"
// @Router		/peers [post]
func (a *RestServer) postApiPeersHandler(w http.ResponseWriter, r *http.Request) {
	peers, err := a.doPostPeers(w, r)
	if err != nil {
		a.savePeers(peers, r)
	}
}

// @Summary		Update peer list.
// @Produce		json
// @Success		204		{string}	string		"No content"
// @Failure		401		{error}		error		"Authentication failed"
// @Failure		403		{error}		error		"Bad request"
// @Router		/peers [put]
func (a *RestServer) putApiPeersHandler(w http.ResponseWriter, r *http.Request) {
	if a.doDeletePeers(w, r) == nil {
		if peers, err := a.doPostPeers(w, r); err == nil {
			a.savePeers(peers, r)
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// @Summary		Remove peers from list.
// @Produce		json
// @Success		204		{string}	string		"No content"
// @Failure		401		{error}		error		"Authentication failed"
// @Failure		403		{error}		error		"Bad request"
// @Router		/peers [delete]
func (a *RestServer) deleteApiPeersHandler(w http.ResponseWriter, r *http.Request) {
	if a.doDeletePeers(w, r) == nil {
		a.savePeers(nil, r)
		w.WriteHeader(http.StatusNoContent)
	}
}

func (a *RestServer) doDeletePeers(w http.ResponseWriter, r *http.Request) error {
	err := a.Core.RemovePeers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return err
}

func (a *RestServer) doPostPeers(w http.ResponseWriter, r *http.Request) (peers []map[string]string, err error) {
	err = json.NewDecoder(r.Body).Decode(&peers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, peer := range peers {
		if err = a.Core.AddPeer(peer["url"], peer["interface"]); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	return
}

func (a *RestServer) savePeers(peers []map[string]string, r *http.Request) {
	a.saveConfig(func(cfg *config.NodeConfig) {
		cfg.Peers = []string{}
		cfg.InterfacePeers = map[string][]string{}
		for _, peer := range peers {
			if peer["interface"] == "" {
				cfg.Peers = append(cfg.Peers, peer["url"])
			} else {
				cfg.InterfacePeers[peer["interface"]] = append(cfg.InterfacePeers[peer["interface"]], peer["url"])
			}
		}
	}, r)
}

func (a *RestServer) saveConfig(setConfigFields func(*config.NodeConfig), r *http.Request) {
	if len(a.ConfigFn) > 0 {
		saveHeaders := r.Header["Riv-Save-Config"]
		if len(saveHeaders) > 0 && saveHeaders[0] == "true" {
			cfg, err := defaults.ReadConfig(a.ConfigFn)
			if err == nil {
				if setConfigFields != nil {
					setConfigFields(cfg)
				}
				err := defaults.WriteConfig(a.ConfigFn, cfg)
				if err != nil {
					a.Log.Errorln("Config file write error:", err)
				}
			} else {
				a.Log.Errorln("Config file read error:", err)
			}
		}
	}
}

func applyKeyParameterized(w http.ResponseWriter, r *http.Request, fn func(key string) (map[string]any, error)) {
	cnt := strings.Split(r.URL.Path, "/")
	if len(cnt) != 5 || cnt[4] == "" {
		http.Error(w, "No remote public key supplied", http.StatusBadRequest)
		return
	}
	result, err := fn(cnt[4])
	if err == nil {
		WriteJson(w, r, result)
	} else if errors.As(err, &core.ErrTimeout) {
		http.Error(w, "Node inaccessible", http.StatusBadGateway)
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// @Summary		Show NodeInfo of a remote node.
// @Produce		json
// @Param		key	path			string				true	"Public key string"
// @Success		200		{string}	string		"ok"
// @Failure		400		{error}		error		"Method not allowed"
// @Failure		401		{error}		error		"Authentication failed"
// @Failure		404		{error}		error		"Not found"
// @Router		/remote/nodeinfo/{key} [get]
func (a *RestServer) getApiRemoteNodeinfoHandler(w http.ResponseWriter, r *http.Request) {
	applyKeyParameterized(w, r, a.Core.GetNodeInfo)
}

// @Summary		Show details about a remote node.
// @Produce		json
// @Param		key	path			string				true	"Public key string"
// @Success		200		{string}	string		"ok"
// @Failure		400		{error}		error		"Method not allowed"
// @Failure		401		{error}		error		"Authentication failed"
// @Failure		404		{error}		error		"Not found"
// @Router		/remote/self/{key} [get]
func (a *RestServer) getApiRemoteSelfHandler(w http.ResponseWriter, r *http.Request) {
	applyKeyParameterized(w, r, a.Core.RemoteGetSelf)
}

// @Summary		Show connected peers to a remote node.
// @Produce		json
// @Param		key	path			string				true	"Public key string"
// @Success		200		{string}	string		"ok"
// @Failure		400		{error}		error		"Method not allowed"
// @Failure		401		{error}		error		"Authentication failed"
// @Failure		404		{error}		error		"Not found"
// @Router		/remote/peers/{key} [get]
func (a *RestServer) getApiRemotePeersHandler(w http.ResponseWriter, r *http.Request) {
	applyKeyParameterized(w, r, a.Core.RemoteGetPeers)
}

// @Summary		Show DHT entries of a remote node.
// @Produce		json
// @Param		key	path			string				true	"Public key string"
// @Success		200		{string}	string		"ok"
// @Failure		400		{error}		error		"Method not allowed"
// @Failure		401		{error}		error		"Authentication failed"
// @Failure		404		{error}		error		"Not found"
// @Router		/remote/dht/{key} [get]
func (a *RestServer) getApiRemoteDHTHandler(w http.ResponseWriter, r *http.Request) {
	applyKeyParameterized(w, r, a.Core.RemoteGetDHT)
}

func (a *RestServer) postApiHealthHandler(w http.ResponseWriter, r *http.Request) {
	peer_list := []string{}

	err := json.NewDecoder(r.Body).Decode(&peer_list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	go a.testAllHealth(peer_list)
	w.WriteHeader(http.StatusAccepted)
}

// @Summary		Return server side events. The output contains following fields: id, event, data.
// @Produce		json
// @Success		200		{string}	string		"ok"
// @Failure		400		{error}		error		"Method not allowed"
// @Failure		401		{error}		error		"Authentication failed"
// @Router		/sse [get]
func (a *RestServer) getApiSseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/event-stream")
Loop:
	for {
		select {
		case v := <-a.serverEvents:
			fmt.Fprintln(w, "id:", a.serverEventNextId)
			fmt.Fprintln(w, "event:", v.Event)
			fmt.Fprintln(w, "data:", string(v.Data))
			fmt.Fprintln(w) //end of event
			a.serverEventNextId += 1
		default:
			break Loop
		}
	}
	if a.updateTimer != nil {
		select {
		case <-a.updateTimer.C:
			go a.sendSseUpdate()
			a.updateTimer.Reset(time.Second * 5)
		default:
		}
	} else {
		a.updateTimer = time.NewTimer(time.Second * 5)
	}
}

func (a *RestServer) sendSseUpdate() {
	rx, tx := a.getPeersRxTxBytes()
	a.serverEvents <- ServerEvent{Event: "rxtx", Data: []byte(fmt.Sprintf(`[{"bytes_recvd":%d,"bytes_sent":%d}]`, rx, tx))}
	data, _ := json.Marshal(a.Core.GetSelf().Coords)
	a.serverEvents <- ServerEvent{Event: "coord", Data: data}
}

func (a *RestServer) testAllHealth(peers []string) {
	for _, u := range peers {
		go func(u string) {
			health := a.testOneHealth(u)
			data, _ := json.Marshal(health)
			a.serverEvents <- ServerEvent{Event: "health", Data: data}
		}(u)
	}
}

func (a *RestServer) testOneHealth(peer string) map[string]any {
	result := map[string]any{
		"peer": peer,
	}
	u, err := url.Parse(peer)
	if err != nil {
		result["error"] = err.Error()
		return result
	}

	ipaddr, err := net.ResolveIPAddr("ip", u.Hostname())
	if err != nil {
		result["error"] = err.Error()
		return result
	}

	result["remote_ip"] = ipaddr.String()

	result["country_short"], result["country_long"] = a.getCountry(ipaddr.String())

	t := time.Now()
	address := ipaddr.String()
	intPort, err := strconv.Atoi(u.Port())
	if err == nil {
		tcpaddr := net.TCPAddr{
			IP:   ipaddr.IP,
			Port: intPort,
		}
		address = tcpaddr.String()
	}

	_, err = net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		result["error"] = err.Error()
		return result
	}
	d := time.Since(t)
	result["ping"] = d.Milliseconds()
	return result
}

func (a *RestServer) getPeersRxTxBytes() (uint64, uint64) {
	var rx uint64
	var tx uint64

	peers := a.Core.GetPeers()
	for _, p := range peers {
		rx += p.RXBytes
		tx += p.TXBytes
	}
	return rx, tx
}

func (a *RestServer) getCountry(ipaddr string) (country_short string, country_long string) {
	if a.ip2locatinoDb != nil {
		ipLoc, err := a.ip2locatinoDb.Get_all(ipaddr)
		if err == nil {
			if ipLoc.Country_short != ip2loc_not_supported && ipLoc.Country_short != ip2loc_invalid_ip_address {
				country_short = ipLoc.Country_short
			}

			if ipLoc.Country_long != ip2loc_not_supported && ipLoc.Country_long != ip2loc_invalid_ip_address {
				country_long = ipLoc.Country_long
			}
		}
	}
	return
}

func WriteError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func WriteJson(w http.ResponseWriter, r *http.Request, data any) {
	if r.URL.Query().Has("fmt") && r.URL.Query()["fmt"][0] == "table" {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		printer := tableprinter.New(w)
		printer.RowLengthTitle = func(int) bool { return false }
		printer.Print(data)
	} else {
		b, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, string(b))
	}
}
