package restapi

import (
	"bytes"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	"archive/zip"
	"time"

	"gerace.dev/zipfs"
	"golang.org/x/exp/slices"

	"github.com/RiV-chain/RiV-mesh/src/core"
	"github.com/RiV-chain/RiV-mesh/src/defaults"
	"github.com/RiV-chain/RiV-mesh/src/version"
	"github.com/ip2location/ip2location-go/v9"
)

var _ embed.FS

//go:embed IP2LOCATION-LITE-DB1.BIN
var IP2LOCATION []byte

const ip2loc_not_supported string = "This parameter is unavailable for selected data file. Please upgrade the data file."

type ServerEvent struct {
	Event string
	Data  []byte
}

type ApiHandler struct {
	pattern string // Context path pattern
	desc    string // What does the endpoint do?
	//	args    []string            // List of human-readable argument names
	handler func(w http.ResponseWriter, r *http.Request) // First is input map, second is output
}

type RestServerCfg struct {
	Core          *core.Core
	Log           core.Logger
	ListenAddress string
	WwwRoot       string
	ConfigFn      string
	handlers      []ApiHandler
}

type RestServer struct {
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

	a.AddHandler(ApiHandler{pattern: "/api", desc: "\tGET - API documentation", handler: a.apiHandler})
	a.AddHandler(ApiHandler{pattern: "/api/self", desc: "GET - Show details about this node", handler: a.apiSelfHandler})
	a.AddHandler(ApiHandler{pattern: "/api/peers", desc: `GET - Show directly connected peers 
			POST - Append peers to the peers list
			PUT - Set peers list
			DELETE - Remove all peers from this node
			Request header "Riv-Save-Config: true" persists changes`, handler: a.apiPeersHandler})
	a.AddHandler(ApiHandler{pattern: "/api/health", desc: "POST - Run peers health check task", handler: a.apiHealthHandler})
	a.AddHandler(ApiHandler{pattern: "/api/sse", desc: "GET - Return server side events", handler: a.apiSseHandler})
	a.AddHandler(ApiHandler{pattern: "/api/dht", desc: "GET - Show known DHT entries", handler: a.apiDhtHandler})
	a.AddHandler(ApiHandler{pattern: "/api/sessions", desc: "GET - Show established traffic sessions with remote nodes", handler: a.apiSessionsHandler})

	var _ = a.Core.PeersChangedSignal.Connect(func(data interface{}) {
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
		return strings.Compare(a.handlers[i].pattern, a.handlers[j].pattern) < 0
	})
	l, e := net.Listen("tcp4", a.listenUrl.Host)
	if e != nil {
		return fmt.Errorf("http server start error: %w", e)
	} else {
		a.Log.Infof("Started http server listening on %s. Document root %s %s\n", a.ListenAddress, a.WwwRoot, a.docFsType)
	}
	go func() {
		err := http.Serve(l, nil)
		if err != nil {
			a.Log.Errorln(err)
		}
	}()
	return nil
}

// AddHandler is called for each admin function to add the handler and help documentation to the API.
func (a *RestServer) AddHandler(handler ApiHandler) error {
	if idx := slices.IndexFunc(a.handlers, func(h ApiHandler) bool { return h.pattern == handler.pattern }); idx >= 0 {
		return errors.New("handler " + handler.pattern + " already exists")
	}
	a.handlers = append(a.handlers, handler)
	http.HandleFunc(handler.pattern, handler.handler)
	return nil
}

func addNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Add("Pragma", "no-cache")
	w.Header().Add("Expires", "0")
}

func (a *RestServer) apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Add("Content-Type", "text/plain")
	for _, h := range a.handlers {
		fmt.Fprintf(w, "%s\t\t%s\n\n", h.pattern, h.desc)
	}
}

func (a *RestServer) apiSelfHandler(w http.ResponseWriter, r *http.Request) {
	addNoCacheHeaders(w)
	switch r.Method {
	case "GET":
		self := a.Core.GetSelf()
		snet := a.Core.Subnet()
		var result = map[string]interface{}{
			"build_name":    a.Core.GetSelf(),
			"build_version": version.BuildVersion(),
			"key":           hex.EncodeToString(self.Key[:]),
			"private_key":   hex.EncodeToString(self.PrivateKey[:]),
			"address":       a.Core.Address().String(),
			"subnet":        snet.String(),
			"coords":        self.Coords,
		}
		writeJson(w, result)
	default:
		writeError(w, http.StatusMethodNotAllowed)
	}
}

func (a *RestServer) apiDhtHandler(w http.ResponseWriter, r *http.Request) {
	addNoCacheHeaders(w)
	switch r.Method {
	case "GET":
		dht := a.Core.GetDHT()
		result := make([]map[string]interface{}, 0, len(dht))
		for _, d := range dht {
			addr := a.Core.AddrForKey(d.Key)
			entry := map[string]interface{}{
				"address": net.IP(addr[:]).String(),
				"key":     hex.EncodeToString(d.Key),
				"port":    d.Port,
				"rest":    d.Rest,
			}
			result = append(result, entry)
		}
		sort.SliceStable(result, func(i, j int) bool {
			return strings.Compare(result[i]["key"].(string), result[j]["key"].(string)) < 0
		})
		writeJson(w, result)
	default:
		writeError(w, http.StatusMethodNotAllowed)
	}
}

func (a *RestServer) apiSessionsHandler(w http.ResponseWriter, r *http.Request) {
	addNoCacheHeaders(w)
	switch r.Method {
	case "GET":
		sessions := a.Core.GetSessions()
		result := make([]map[string]interface{}, 0, len(sessions))
		for _, s := range sessions {
			addr := a.Core.AddrForKey(s.Key)
			entry := map[string]interface{}{
				"address":     net.IP(addr[:]).String(),
				"key":         hex.EncodeToString(s.Key),
				"bytes_recvd": s.RXBytes,
				"bytes_sent":  s.TXBytes,
				"uptime":      s.Uptime.Seconds(),
			}
			result = append(result, entry)
		}
		sort.SliceStable(result, func(i, j int) bool {
			return strings.Compare(result[i]["key"].(string), result[j]["key"].(string)) < 0
		})
		writeJson(w, result)
	default:
		writeError(w, http.StatusMethodNotAllowed)
	}
}

func (a *RestServer) prepareGetPeers() []map[string]interface{} {
	peers := a.Core.GetPeers()
	response := make([]map[string]interface{}, 0, len(peers))
	for _, p := range peers {
		addr := a.Core.AddrForKey(p.Key)
		entry := map[string]interface{}{
			"address":     net.IP(addr[:]).String(),
			"key":         hex.EncodeToString(p.Key),
			"port":        p.Port,
			"priority":    uint64(p.Priority), // can't be uint8 thanks to gobind
			"coords":      p.Coords,
			"remote":      p.Remote,
			"remote_ip":   p.RemoteIp,
			"bytes_recvd": p.RXBytes,
			"bytes_sent":  p.TXBytes,
			"uptime":      p.Uptime.Seconds(),
			"multicast":   strings.Contains(p.Remote, "[fe80::"),
		}

		a.fillCountry(&entry, p.RemoteIp)
		response = append(response, entry)
	}
	sort.Slice(response, func(i, j int) bool {
		if response[i]["multicast"].(bool) != response[j]["multicast"].(bool) {
			return (!response[i]["multicast"].(bool) && response[j]["multicast"].(bool))
		}

		if response[i]["priority"].(uint64) != response[j]["priority"].(uint64) {
			return response[i]["priority"].(uint64) < response[j]["priority"].(uint64)
		}

		if cmp := strings.Compare(response[i]["address"].(string), response[j]["address"].(string)); cmp != 0 {
			return cmp < 0
		}
		return response[i]["port"].(uint64) < response[j]["port"].(uint64)
	})
	return response
}

func (a *RestServer) apiPeersHandler(w http.ResponseWriter, r *http.Request) {
	var handleDelete = func() error {
		err := a.Core.RemovePeers()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return err
	}

	var handlePost = func() ([]string, error) {
		var peers []string
		err := json.NewDecoder(r.Body).Decode(&peers)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil, err
		}

		for _, peer := range peers {
			if err := a.Core.AddPeer(peer, ""); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return nil, err
			}
		}

		return peers, nil
	}

	var saveConfig = func(peers []string) {
		if len(a.ConfigFn) > 0 {
			saveHeaders := r.Header["Riv-Save-Config"]
			if len(saveHeaders) > 0 && saveHeaders[0] == "true" {
				cfg, err := defaults.ReadConfig(a.ConfigFn)
				if err == nil {
					cfg.Peers = peers
					err := defaults.WriteConfig(a.ConfigFn, cfg)
					if err != nil {
						a.Log.Errorln("Config file read error:", err)
					}
				} else {
					a.Log.Errorln("Config file read error:", err)
				}
			}
		}
	}
	addNoCacheHeaders(w)
	switch r.Method {
	case "GET":
		writeJson(w, a.prepareGetPeers())
	case "POST":
		peers, err := handlePost()
		if err != nil {
			saveConfig(peers)
		}
	case "PUT":
		if handleDelete() == nil {
			if peers, err := handlePost(); err == nil {
				saveConfig(peers)
				w.WriteHeader(http.StatusNoContent)
			}
		}
	case "DELETE":
		if handleDelete() == nil {
			saveConfig([]string{})
			w.WriteHeader(http.StatusNoContent)
		}
	default:
		writeError(w, http.StatusMethodNotAllowed)
	}
}

func (a *RestServer) apiHealthHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		peer_list := []string{}

		err := json.NewDecoder(r.Body).Decode(&peer_list)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		go a.testAllHealth(peer_list)
		w.WriteHeader(http.StatusAccepted)
	default:
		writeError(w, http.StatusMethodNotAllowed)
	}
}

func (a *RestServer) apiSseHandler(w http.ResponseWriter, r *http.Request) {
	addNoCacheHeaders(w)
	switch r.Method {
	case "GET":
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
	default:
		writeError(w, http.StatusMethodNotAllowed)
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

func (a *RestServer) testOneHealth(peer string) map[string]interface{} {
	result := map[string]interface{}{
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

	a.fillCountry(&result, ipaddr.String())

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

func (a *RestServer) fillCountry(entry *map[string]interface{}, ipaddr string) {
	if a.ip2locatinoDb != nil {
		ipLoc, err := a.ip2locatinoDb.Get_all(ipaddr)
		if err == nil {
			if ipLoc.Country_short != ip2loc_not_supported {
				(*entry)["country_short"] = ipLoc.Country_short
			}

			if ipLoc.Country_long != ip2loc_not_supported {
				(*entry)["country_long"] = ipLoc.Country_long
			}
		}
	}
}

func writeError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func writeJson(w http.ResponseWriter, data any) {
	b, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	fmt.Fprint(w, string(b))
}
