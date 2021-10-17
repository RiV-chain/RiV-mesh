package main

import (
	"github.com/webview/webview"
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
	
	"github.com/RiV-chain/RiV-mesh/src/admin"	
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

func run_command(riv_ctrl_path string, command string) []byte{
	args := []string{"-json", command}
	cmd := exec.Command(riv_ctrl_path, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
		return nil
	}
	return out
}

func get_self(w webview.WebView, riv_ctrl_path string){

	res := &admin.GetSelfResponse{}
	out := run_command(riv_ctrl_path, "getSelf")
	if err := json.Unmarshal(out, &res); err != nil {
		return
	}
	for ipv6, s := range res.Self {
		//found ipv6
		fmt.Printf("IPv6: %s\n", ipv6)		
		go setFieldValue(w, "ipv6", ipv6)
		//found subnet
		fmt.Printf("Subnet: %s\n", s.Subnet)
		go setFieldValue(w, "subnet", s.Subnet)
	}	
}

func get_peers(w webview.WebView, riv_ctrl_path string){

	res := &admin.GetPeersResponse{}
	out := run_command(riv_ctrl_path, "getPeers")
	if err := json.Unmarshal(out, &res); err != nil {
		return
	}

	var m []string
	for _, s := range res.Peers {
		m=append(m, s.Remote)
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
