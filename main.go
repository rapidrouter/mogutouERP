package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Allenxuxu/mogutouERP/models"
	"github.com/Allenxuxu/mogutouERP/pkg/token"
	"github.com/gin-gonic/gin"
	"github.com/pkg/browser"
)

func main() {
	path := flag.String("c", "/etc/conf", "配置文件夹路径")
	flag.Parse()

	token.InitConfig(*path+"/jwt.json", "jwt-key")

	// read config
	confData, err := os.ReadFile(*path + "/conf.json")
	if err != nil {
		log.Fatal(err)
	}

	var confRoot struct {
		Mysql models.DBInfo `json:"mysql"`
	}
	if err := json.Unmarshal(confData, &confRoot); err != nil {
		log.Fatal(err)
	}

	models.Init(&confRoot.Mysql)

	gin.DisableConsoleColor()
	r := initRouter()

	if os.Getenv("MOGUTOU_NO_BROWSER") == "" {
		go func() {
			time.Sleep(time.Second)
			_ = browser.OpenURL("http://127.0.0.1:1988/ui")
		}()
	}
	fmt.Println("Open: http://127.0.0.1:1988/ui")

	r.Run(":1988") // listen and serve on 0.0.0.0:1988
}
