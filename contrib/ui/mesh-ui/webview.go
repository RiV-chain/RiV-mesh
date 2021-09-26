package main

import (
	"github.com/RiV-chain/RiV-mesh/src/defaults"
	"github.com/RiV-chain/RiV-mesh/src/config"
	"golang.org/x/text/encoding/unicode"
	"github.com/webview/webview"
	"github.com/hjson/hjson-go"
	"encoding/json"
	"path/filepath"
	"io/ioutil"
	"os/exec"
	"net/url"
	"runtime"
	"strings"
	"log"
	"fmt"
	"os"	
)

func main() {
    debug := true
    w := webview.New(debug)
    defer w.Destroy()
    w.SetTitle("RiV-mesh")
    w.SetSize(470, 415, webview.HintNone)
    path, err := filepath.Abs(filepath.Dir(os.Args[0]))
    if err != nil {
            log.Fatal(err)
    }
    log.Println(path)
    w.Bind("onLoad", func() {
	log.Println("page loaded")
	
	go run(w)
    })
    w.Bind("savePeers", func(peer_list string) {
	//log.Println("peers saved ", peer_list)
	var peers []string
	_ = json.Unmarshal([]byte(peer_list), &peers)
	log.Printf("Unmarshaled: %v", peers)
	var conf []byte
	var err error
	conf, err = ioutil.ReadFile(defaults.GetDefaults().DefaultConfigFile)
	
	// If there's a byte order mark - which Windows 10 is now incredibly fond of
	// throwing everywhere when it's converting things into UTF-16 for the hell
	// of it - remove it and decode back down into UTF-8. This is necessary
	// because hjson doesn't know what to do with UTF-16 and will panic
	if bytes.Equal(conf[0:2], []byte{0xFF, 0xFE}) ||
		bytes.Equal(conf[0:2], []byte{0xFE, 0xFF}) {
		utf := unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
		decoder := utf.NewDecoder()
		conf, err = decoder.Bytes(conf)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
	}
	
	var dat map[string]interface{}
	if err := hjson.Unmarshal(conf, &dat); err != nil {
		log.Printf("Error: %v", err)
		return
	}
	dat["Peers"] = peers
	// Sanitise the config
	confHjson, err := hjson.Marshal(dat)
	if err != nil {
		panic(err)
	}
	//use meshctl to pass config into stdin as a new config
    })
    dat, err := ioutil.ReadFile(path+"/index.html")
    w.Navigate("data:text/html,"+url.QueryEscape(string(dat)))
    w.Run()
}

func run(w webview.WebView){
	if runtime.GOOS == "windows" {
		program_path := "programfiles"
		path, exists := os.LookupEnv(program_path)
		if exists {
			fmt.Println("Program path: %s", path)
			riv_ctrl_path := fmt.Sprintf("%s\\RiV-mesh\\meshctl.exe", path)
			get_self(w, riv_ctrl_path)
			get_peers(w, riv_ctrl_path)
		} else {
			fmt.Println("could not find Program Files path")
		}
	} else {
		riv_ctrl_path := fmt.Sprintf("meshctl")
		get_self(w, riv_ctrl_path)
		get_peers(w, riv_ctrl_path)
	}
}

func run_command(riv_ctrl_path string, command string) []string{
	cmd := exec.Command(riv_ctrl_path, command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
		return nil
	}
	lines := strings.Split(string(out), "\n")
	return lines
}

func get_self(w webview.WebView, riv_ctrl_path string){
	
	lines := run_command(riv_ctrl_path, "getSelf")
	m := make(map[string]string)
	for _, s := range lines {
		p := strings.SplitN(s, ":", 2)
		if len(p)>1 {
			m[p[0]]=strings.TrimSpace(p[1])
		}
	}
	if val, ok := m["IPv6 address"]; ok {
		//found ipv6
		fmt.Printf("IPv6: %s\n", val)
		go setFieldValue(w, "ipv6", val)
	}
	if val, ok := m["IPv6 subnet"]; ok {
		//found subnet
		fmt.Printf("Subnet: %s\n", val)
		go setFieldValue(w, "subnet", val)
	}	
}

func get_peers(w webview.WebView, riv_ctrl_path string){
	
	lines := run_command(riv_ctrl_path, "getPeers")
	lines = lines[1:] /*remove first element which is a header*/
	var m []string
	r:=""
	for _, s := range lines {
		p := strings.SplitN(s, " ", -1)
		if len(p)>1 {
			for _, t := range p {
				if len(strings.TrimSpace(t))>0 {
					r=strings.TrimSpace(t)
				}
			}
			index_p := strings.Index(r, "%")
			index_b := strings.Index(r, "]")
			if index_p>0 && index_b>0 {
				r = r[:index_p]+r[index_b:]
			}
			m=append(m, r)
		}
	}
	for k := range m {         
	    // Loop
	    fmt.Println(k)
	}
	inner_html := strings.Join(m[:], "<br>")
	go setFieldValue(w, "peers", inner_html)
}

func setFieldValue(p webview.WebView, id string, value string) {
	p.Dispatch(func() {
		p.Eval("setFieldValue('"+id+"','"+value+"');")
	})
}
