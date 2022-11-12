package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/cloudflare/cloudflare-go"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	path2 "path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Config struct {
	Interval time.Duration `json:"intervalTime"`
	Email    string        `json:"Email"`
	APIKey   string        `json:"APIKey"`
	Zones    []Zone        `json:"Zones"`
}

type Zone struct {
	Name    string   `json:"Name"`
	Records []Record `json:"Records"`
}

type Record struct {
	Name string `json:"Name"`
}

func getExcPath() string {
	file, _ := exec.LookPath(os.Args[0])
	// 获取包含可执行文件名称的路径
	path, _ := filepath.Abs(file)
	// 获取可执行文件所在目录
	index := strings.LastIndex(path, string(os.PathSeparator))
	ret := path[:index]
	return strings.Replace(ret, "\\", "/", -1)
}

func initConfig() Config {
	path := getExcPath()
	var config Config
	configFile, err := os.Open(path2.Join(path, "config.json"))
	if err != nil {
		config = Config{
			Interval: 60,
			Email:    "Example@outlook.com",
			APIKey:   "CLOUDFLARE_API_EMAIL",
			Zones: []Zone{
				{
					Name: "example.com",
					Records: []Record{
						{
							Name: "home.example.com",
						},
					},
				},
			},
		}
		configFile, err = os.Create(path2.Join(path, "config.json"))
		if err != nil {
			log.Fatal(err)
		}
		defer configFile.Close()
		configJson, err := json.Marshal(config)
		if err != nil {
			log.Fatal(err)
		}
		var out bytes.Buffer
		err = json.Indent(&out, configJson, "", "\t")
		_, err = configFile.Write(out.Bytes())

	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

func getPublicIPV4() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	return string(content)
}

func getMyIPV6() string {
	s, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range s {
		i := regexp.MustCompile(`(\w+:){7}\w+`).FindString(a.String())
		if strings.Count(i, ":") == 7 {
			return i
		}
	}
	return ""
}

func UpdateDNS(nowIp, id string, record cloudflare.DNSRecord, api *cloudflare.API) {
	if record.Content != nowIp {
		record.Content = nowIp
		err := api.UpdateDNSRecord(context.Background(), id, record.ID, record)
		log.Println(record.Name, "IP has been updated to", nowIp)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func DDNS(config Config) {
	api, err := cloudflare.New(config.APIKey, config.Email)
	if err != nil {
		log.Fatal(err)
	}
	for _, zone := range config.Zones {
		id, err := api.ZoneIDByName(zone.Name)
		if err != nil {
			log.Fatal(err)
		}
		records, err := api.DNSRecords(context.Background(), id, cloudflare.DNSRecord{})
		if err != nil {
			log.Fatal(err)
		}
		for _, record := range records {
			for _, zoneRecord := range zone.Records {
				if zoneRecord.Name == record.Name {
					var nowIp string
					switch record.Type {
					case "A":
						{
							nowIp = getPublicIPV4()
						}
					case "AAAA":
						{
							nowIp = getMyIPV6()
						}
					}
					UpdateDNS(nowIp, id, record, api)
				}
			}
		}
	}
}

func main() {
	log.Println("DDNS is running...by QingYu")
	config := initConfig()
	ticker := time.NewTicker(config.Interval * time.Second)
	for {
		<-ticker.C
		DDNS(config)
	}
	ticker.Stop()
}
