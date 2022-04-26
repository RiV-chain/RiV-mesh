package main

import (
    "github.com/webview/webview"
    "github.com/hjson/hjson-go"
    "encoding/json"
    "path/filepath"
    "io/ioutil"
    "os/exec"
    "net/url"
    "runtime"
    "strings"
    "strconv"
    "time"
    "net"
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
    w.SetSize(690, 920, webview.HintFixed)
    /*1. Create ~/.riv-mesh folder if not existing
     *2. Create ~/.riv-mesh/mesh.conf if not existing
     *3. If the file exists read Peers. 
     *3.1 Invoke add peers for each record
     */
    mesh_folder := ".riv-mesh"
    mesh_conf := "mesh.conf"
    user_home := get_user_home_path()
    mesh_settings_folder := filepath.Join(user_home, mesh_folder)
    err := os.MkdirAll(mesh_settings_folder, os.ModePerm)
    if err != nil {
        fmt.Printf("Unable to create folder: %v", err)
    }
    mesh_settings_path := filepath.Join(user_home, mesh_folder, mesh_conf)
    if _, err := os.Stat(mesh_settings_path); os.IsNotExist(err) { 
        err := ioutil.WriteFile(mesh_settings_path, []byte(""), 0750)
        if err != nil {
            fmt.Printf("Unable to write file: %v", err)
        }
    } else {
        //read peers from mesh.conf
        conf, _ := ioutil.ReadFile(mesh_settings_path)
        var dat map[string]interface {}
       	if err := hjson.Unmarshal(conf, &dat); err != nil {
        	fmt.Printf("Unable to parse mesh.conf file: %v", err)
        } else {
            if dat["Peers"]!=nil {
                peers := dat["Peers"].([]interface{}) 
                remove_peers()
                for _, u := range peers {
                   log.Printf("Unmarshaled: %v", u.(string))
                   add_peers(u.(string))
                }
            } else {
                fmt.Printf("Warning: Peers array not loaded from mesh.conf file")
            }
        }
    }
    var path string

    if len(os.Args)>1 {
        path, err = filepath.Abs(filepath.Dir(os.Args[1]))
    } else {
        path, err = filepath.Abs(filepath.Dir(os.Args[0]))
    }
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
       remove_peers()
       for _, u := range peers {
            log.Printf("Unmarshaled: %v", u)
            add_peers(u)
       }
       //add peers to ~/mesh.conf
       dat := make(map[string]interface{})
       dat["Peers"] = peers
       bs, _ := hjson.Marshal(dat)
       e := ioutil.WriteFile(mesh_settings_path, bs, 0750)
       if e != nil {
           fmt.Printf("Unable to write file: %v", e)
       }
    })
    w.Bind("ping", func(peer_list string) {
        go ping(w, peer_list)
    })
    dat, err := ioutil.ReadFile(path+"/index.html")
    w.Navigate("data:text/html,"+url.QueryEscape(string(dat)))
    //w.Navigate("data:text/html,"+"<html>"+path+"</html>")
    w.Run()
}

func ping(w webview.WebView, peer_list string){
    var peers []string
    _ = json.Unmarshal([]byte(peer_list), &peers)
    log.Printf("Unmarshaled: %v", peers)
    for _, u := range peers {
        log.Printf("Unmarshaled: %v", u)
        ping_time := check(u);
        log.Printf("ping: %d", ping_time)
        setPingValue(w, u, strconv.FormatInt(ping_time, 10));
    }
}

func check(peer string) int64 {
    u, e := url.Parse(peer)
    if e!=nil {
        return -1
    }
    t := time.Now()
    _, err := net.DialTimeout("tcp", u.Host, 5*time.Second)
    if err!=nil {
        return -1
    }
    d := time.Since(t)
    return d.Milliseconds()
}

func get_user_home_path() string {
    if runtime.GOOS == "windows" {
        path, exists := os.LookupEnv("USERPROFILE")
        if exists {
            return path
        } else {
            return ""
        }
    } else {
        path, exists := os.LookupEnv("HOME")
        if exists {
            return path
        } else {
            return ""
        }
    }
}

func get_ctl_path() string{
    if runtime.GOOS == "windows" {
		program_path := "programfiles"
		path, exists := os.LookupEnv(program_path)
		if exists {
			fmt.Println("Program path: %s", path)
			riv_ctrl_path := fmt.Sprintf("%s\\RiV-mesh\\meshctl.exe", path)
			return riv_ctrl_path
		} else {
			fmt.Println("could not find Program Files path")
            return ""
		}
	} else {
		riv_ctrl_path := fmt.Sprintf("meshctl")
		return riv_ctrl_path
	}
}

func run(w webview.WebView){
    //var peersTimer *time.Timer
    riv_ctrl_path := get_ctl_path()
    if riv_ctrl_path != "" {
        get_self(w, riv_ctrl_path)
	get_peers(w, riv_ctrl_path)
    }
    _ = time.AfterFunc(10*time.Second, func() {
        run(w)
    })
}

func run_command(riv_ctrl_path string, command string) []byte{
	args := []string{"-json", command}
	cmd := exec.Command(riv_ctrl_path, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		//log.Fatalf("cmd.Run() failed with %s\n", err)
		return []byte(err.Error())
	}
	return out
}

func run_command_with_arg(riv_ctrl_path string, command string, arg string) []byte{
	args := []string{"-json", command, arg}
	cmd := exec.Command(riv_ctrl_path, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
        	//log.Fatalf("command failed: %s\n", riv_ctrl_path+" "+strings.Join(args, " "))
		return []byte(err.Error())
	}
	return out
}

func add_peers(uri string){
	riv_ctrl_path := get_ctl_path()
	run_command_with_arg(riv_ctrl_path, "addpeers", "uri="+uri)	
}

func remove_peers(){
	riv_ctrl_path := get_ctl_path()
	run_command(riv_ctrl_path, "removepeers")	
}

func get_self(w webview.WebView, riv_ctrl_path string){

	res := &admin.GetSelfResponse{}
	out := run_command(riv_ctrl_path, "getSelf")
	if err := json.Unmarshal(out, &res); err != nil {
		go setFieldValue(w, "ipv6", string(out))
		return
	}
	for ipv6, s := range res.Self {
		//found ipv6
		fmt.Printf("IPv6: %s\n", ipv6)		
		go setFieldValue(w, "ipv6", string(out))
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

func setPingValue(p webview.WebView, peer string, value string) {
	p.Dispatch(func() {
		p.Eval("setPingValue('"+peer+"','"+value+"');")
	})
}
