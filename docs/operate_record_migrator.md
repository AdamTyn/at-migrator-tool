## operate-record-migrator

```text
- version v1.3
- latest_modify 2022.09.15
```

1. 链路分析

    ```bash
    [▼] main.go
    [▼] internal/conf/conf.pb.go # 解析配置
    [▼] internal/collector/{company_operation_log,delivery_operation_log,intern_operation_log}.go # 加载采集器
    [▼] internal/process/operate_record_migrator.go
    [▼] internal/application.go
    ```

2. 子文档
   - [hr-logger旧数据迁移字段对照表](https://mshare.feishu.cn/sheets/shtcnbP0hI3KHistaWdVeSYVLye)