package rain

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yrbb/rain/pkg/utils"
)

func showRainInfo() {
	fmt.Println("⛈️")
	fmt.Printf("Version: %s (%s)\n", Version, VERSION)
	fmt.Println("Build time: " + Compile)
	fmt.Println()
}

func getBinPath() string {
	binFile, _ := exec.LookPath(os.Args[0])
	binPath, _ := filepath.Abs(binFile)

	return binPath
}

func getConfigPath() string {
	cfgPath, _ := filepath.Abs(filepath.Dir(getBinPath()) + "/../config/")

	return cfgPath + "/"
}

func parseServerHostPort(listen string) (host string, port int, err error) {
	arr := strings.Split(listen, ":")
	switch len(arr) {
	case 1:
		port, err = strconv.Atoi(arr[0])
		if err != nil {
			return
		}

		host, err = utils.IP()
	case 2:
		host = arr[0]

		if tmp := strings.Split(host, "."); len(tmp) == 2 && tmp[1] == "*" {
			host, err = utils.IP(tmp[0])
			if err != nil {
				return
			}
		}

		port, err = strconv.Atoi(arr[1])
	default:
		err = fmt.Errorf("无法获取正确服务监听地址及端口")
	}

	if err != nil {
		return
	}

	if host == "" {
		err = fmt.Errorf("无法获取正确服务监听地址及端口")
	}

	return
}
