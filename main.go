/*
分布式配置服务端
数据为JSON格式, 相同KEY首次GET时保存为默认值, POST更新默认值
URL表单内NAME为配置名称, VALUE为JSON格式, 传入JSON属性数量不固定(目前仅支持一层JSON)
例如: http://localhost:8080/?conf={"k1":"v1","k2":"v2"}&conf1={"k":123}
功能单一，没有什么实用价值仅作为练手
*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

var kv sync.Map

//para key:name in url, val:value in url form, must be json; return r:default value
//i.e if url=host:port?conf={"k1":"v1"} then key=conf, val={"k1":"v1"}
func getDefaultValue(method string, key string, val []string) (r []string, e error) {
	uMethod := strings.ToUpper(method)
	if "GET" != uMethod && "POST" != uMethod {
		return nil, fmt.Errorf("wrong para: optional first para GET or POST")
	}

	var f interface{}

	for _, v := range val { //UNIQUE KEY, DO NOT SUPPORT ARRAY VAL
		var b = []byte(v)
		if err := json.Unmarshal(b, &f); err != nil {
			e = fmt.Errorf("invalid format json:%s %v", v, err)
			break
		} else {
			m := f.(map[string]interface{})
			if repo, ok := kv.Load(key); ok { //UPDATE CURRENT NODE
				mrepo := repo.(map[string]interface{})
				for k, v := range m {
					if mv, ok := mrepo[k]; ok {
						switch reflect.TypeOf(m[k]).Kind() {
						case reflect.String:
							if v != mv { //high performance?
								if "GET" == uMethod { //REVERT
									m[k] = mv
								} else if "POST" == uMethod { //SET
									mrepo[k] = v
								}
							}
						case reflect.Struct: //for depth more than 1 better to be detailed
							fallthrough
						default: //bug?
							if "GET" == uMethod { //REVERT
								m[k] = mv
							} else if "POST" == uMethod { //SET
								mrepo[k] = v
							}
						}
					} else { //INSERT NEW NODE
						reflect.ValueOf(&mv).Elem().Set(reflect.ValueOf(v))
						mrepo[k] = reflect.ValueOf(&mv).Interface()
					}
				}
			} else { //INSERT NEW NODE
				mn := make(map[string]interface{}, len(m))
				for k, v := range m {
					mn[k] = reflect.ValueOf(v).Interface()
				}
				kv.Store(key, mn)
			}

			if b, err := json.Marshal(f); err != nil {
				e = fmt.Errorf("error on process json:", err.Error())
			} else {
				r = append(r, string(b))
			}
		}
	}

	return
}

func onRouteMain(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if len(r.Form) == 0 {
		fmt.Fprintln(w, "wrong para: para should not be null!")
	} else {
		for k, v := range r.Form {
			if i, err := getDefaultValue(r.Method, k, v); err != nil {
				fmt.Fprintln(w, err.Error())
			} else {
				for _, mv := range i {
					fmt.Fprintln(w, k, mv)
				}
			}
		}
	}
}

func main() {
	var host *string = flag.String("h", "0.0.0.0", "host")
	var port *uint = flag.Uint("p", 8080, "port")

	flag.Parse()

	url := fmt.Sprintf("%s:%d", *host, *port)
	log.Println("LISTEN URL:", url)

	http.HandleFunc("/", onRouteMain)

	if err := http.ListenAndServe(url, nil); err != nil {
		log.Fatal("failed to open http service:", err.Error())
	}
}
