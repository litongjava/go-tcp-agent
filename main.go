package main

import (
  "bufio"
  "fmt"
  "io"
  "net"
  "os"
  "strings"
)

type ProxyConfig struct {
  listenPort string
  targetHost string
  targetPort string
}

func main() {
  // 读取配置文件
  proxyConfigs := loadConfig("proxy.conf")

  // 启动代理服务
  for _, config := range proxyConfigs {
    go startProxy(config)
  }

  // 等待代理服务结束
  select {}
}

func loadConfig(filename string) []ProxyConfig {
  file, err := os.Open(filename)
  if err != nil {
    fmt.Println("Failed to open config file:", filename)
    os.Exit(1)
  }
  defer file.Close()

  var configs []ProxyConfig
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    line := scanner.Text()
    fields := strings.Split(line, " ")
    if len(fields) != 3 {
      fmt.Println("Invalid config line:", line)
      continue
    }
    config := ProxyConfig{
      listenPort: fields[0],
      targetHost: fields[1],
      targetPort: fields[2],
    }
    configs = append(configs, config)
  }

  if err := scanner.Err(); err != nil {
    fmt.Println("Failed to read config file:", err)
    os.Exit(1)
  }

  return configs
}

func startProxy(config ProxyConfig) {
  listener, err := net.Listen("tcp", ":"+config.listenPort)
  if err != nil {
    fmt.Println("Failed to start proxy on port", config.listenPort)
    os.Exit(1)
  }
  defer listener.Close()
  fmt.Println("Proxy started on port", config.listenPort)

  for {
    clientConn, err := listener.Accept()
    if err != nil {
      fmt.Println("Failed to accept client connection:", err)
      continue
    }

    targetConn, err := net.Dial("tcp", config.targetHost+":"+config.targetPort)
    if err != nil {
      fmt.Println("Failed to connect to target:", err)
      clientConn.Close()
      continue
    }

    // 将客户端的数据转发到目标服务器
    go func() {
      _, err := io.Copy(targetConn, clientConn)
      if err != nil {
        fmt.Println("Error copying data to target:", err)
      }
      targetConn.Close()
      clientConn.Close()
    }()

    // 将目标服务器的数据转发回客户端
    go func() {
      _, err := io.Copy(clientConn, targetConn)
      if err != nil {
        fmt.Println("Error copying data to client:", err)
      }
      targetConn.Close()
      clientConn.Close()
    }()
  }
}
