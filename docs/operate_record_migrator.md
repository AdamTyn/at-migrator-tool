## operate-record-migrator

```go
- version v1.0
- latest_modify 2022.06.30
```

1. 链路分析

    ```bash
    [▼] cmd/main.go
    [▼] internal/conf/conf.pb.go # 解析配置
    [▼] internal/collector/*.go
    [▼] internal/process/operate_record_migrator.go
    [▼] internal/application.go
    ```

2. 流程图
![流程图](operate_record_migrator.流程图.png)
