/*
用KEY-VALUE形式存储JSON配置配置信息
传出时如果之前有则用之前存储的，如果没有则首先存储传入值再将传入值直接传出(即相同的KEY第二次以后的传出值都为第一次的传入值)
浏览器访问：key为配置名称, value为JSON格式, 传入JSON属性数量不固定(目前只支持一层JSON)
例如: http://127.0.0.1:13485/?config={"k1":"v1","k2":"v2"}&config1={"k":123}
功能单一，也可能没有什么实用价值，仅作为练手题目实现方式之一
*/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sync"
)

var kv sync.Map

//para jk:key in url, jv:value in url form, must be json; return r:default value
func getDefaultValue(jk string, jv []string) (r []string, e error) {
	var f interface{}

	for _, v := range jv {
		var b = []byte(v)
		if err := json.Unmarshal(b, &f); err != nil {
			e = fmt.Errorf("invalid format json:%s %s", v, err.Error())
			break
		} else {
			m := f.(map[string]interface{})
			for k, v := range m {
				if mv, ok := kv.LoadOrStore(k, v); ok {
					switch reflect.TypeOf(mv).Kind() {
					case reflect.String:
						if v != mv {
							m[k] = mv
						}
					default: //bug, for depth more than 1 better to be detailed
						reflect.ValueOf(&v).Elem().Set(reflect.ValueOf(mv))
					}
				}
			}

			if b, err := json.Marshal(f); err != nil {
				e = fmt.Errorf("error on process json:", err.Error())
			} else {
				r = append(r, string(b))
			}
		}
	}

	return r, e
}

func onRouteMain(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if len(r.Form) == 0 {
		fmt.Fprintln(w, "wrong para: para should not be null!")
	} else {
		for k, v := range r.Form {
			if i, err := getDefaultValue(k, v); err != nil {
				fmt.Fprintln(w, err.Error())
			} else {
				for _, mv := range i {
					fmt.Fprintln(w, k, mv)
				}
			}
		}
	}
}

func onRouteSet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "to to soon...")
}

func main() {
	http.HandleFunc("/", onRouteMain)   //GET
	http.HandleFunc("/set", onRouteSet) //SET
	if err := http.ListenAndServe(":13485", nil); err != nil {
		log.Fatal("faile to open http service:", err.Error())
	}
}
