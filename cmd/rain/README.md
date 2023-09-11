## Rain 🐶   

> based on cobra-cli 

#### 安装 

```bash
go install github.com/yrbb/rain/cmd/rain@latest
```

#### 使用 

```bash
mkdir newproject  

cd newproject

go mod init newproject

rain init

rain add test
```

#### 运行  

```bash
make run 
// or 
go run main.go test -c=config/config.toml
```