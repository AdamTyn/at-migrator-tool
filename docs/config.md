## config

```json
// 配置文件 config.json 备注
{
  "name": "at-migrator-tool", // app.name
  "server": {
    "http": {
      "endpoint": "0.0.0.0:9000",
      "timeout": 10
    }
  },
  "data": {
    "source": {
      "driver": "postgres",
      "dsn": "dbname=test user=test password=test host=test.com port=1234 sslmode=disable options='-c statement_timeout=5000'"
    },
    "target": {
      "driver": "postgres",
      "dsn": "dbname=test user=test password=test host=test.com port=1234 sslmode=disable options='-c statement_timeout=5000'"
    },
    "redis": {
      "addr": "test.cn:1234",
      "password": "123456",
      "db": 1
    }
  },
  "migrator": { // 迁移配置
    "operate_record": { // OperateRecordMigrator 配置
      "enable": true, // 是否启用
      "fetch_step": 40000, // 每次查询数据源的记录数
      "max_empty_fetch_num": 100 // 最大空查询次数
    },
    "intern_new_job": { // InternNewJobMigrator 配置
      "enable": true, // 是否启用
      "fetch_step": 10000, // 每次查询数据源的记录数
      "max_empty_fetch_num": 10 // 最大空查询次数
    },
    "deliver_uncheck": { // DeliverUncheckMigrator 配置
      "enable": true, // 是否启用
      "truncate_first": false, // 是否先清空历史数据(支持异常重刷)
      "fetch_step": 40000, // 每次查询数据源的记录数
      "max_empty_fetch_num": 10 // 最大空查询次数
    },
    "intern_uncheck": { // InternUncheckMigrator 配置
      "enable": true, // 是否启用
      "fetch_step": 40000, // 每次查询数据源的记录数
      "max_empty_fetch_num": 10 // 最大空查询次数
    }
  },
  "webhook": { // 外部api配置
    "fs_robot": "https://open.feishu.cn/hook/xxxxxxxxxxxxxxx" // 飞书机器人通知地址
  }
}
```