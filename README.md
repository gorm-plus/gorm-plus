Gorm-Plus是Gorm的增强工具，它在保持Gorm原有特性的基础上，为开发者提供了开箱即用的增强功能。通过简化开发流程、提高效率，它为开发者带来了无与伦比的开发体验。如果您渴望尝试一款让开发变得轻松而高效的工具，Gorm-Plus将是您不可错过的选择。



## 特性

1. 无侵入、增强而非改变：Gorm-Plus以无侵入的方式对 Gorm 进行增强，保持原有特性的同时提供更多功能，让您无需担心兼容性问题。
2. 强大的 CRUD 操作和通用查询：Gorm-Plus内置了强大的 CRUD 操作功能，并提供了简便的通用查询功能，无需繁琐的配置即可轻松构建复杂条件查询。
3. 支持指针字段形式查询：通过支持指针字段形式查询，Gorm-Plus让编写各类查询条件变得更加便捷，再也不用担心因为字段写错而出现错误。
4. 内置泛型查询和灵活的返回类型封装：Gorm-Plus内置了泛型查询功能，让您能够灵活封装返回类型，轻松应对各种查询需求。
5. 内置分页插件：Gorm-Plus还提供了内置的分页插件，让分页操作变得简单而高效，让您能够轻松处理大量数据，并实现更好的用户体验。

## 快速上手

现有一张 `Users` 表，其表结构如下：

```SQL
CREATE TABLE `users` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `username` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `password` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `address` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `age` bigint DEFAULT NULL,
  `phone` varchar(11) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `score` bigint DEFAULT NULL,
  `dept` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=407 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

```



对应的数据如下：

```SQL
INSERT INTO `users` (`username`, `password`, `address`, `age`, `phone`, `score`, `dept`, `created_at`, `updated_at`)
VALUES
  ('张三', 'password1', '地址1', 25, '12345678901', 80, '部门1', NOW(), NOW()),
  ('李四', 'password2', '地址2', 30, '12345678902', 90, '部门2', NOW(), NOW()),
  ('王五', 'password3', '地址3', 35, '12345678903', 70, '部门1', NOW(), NOW()),
  ('赵六', 'password4', '地址4', 28, '12345678904', 85, '部门2', NOW(), NOW()),
  ('钱七', 'password5', '地址5', 32, '12345678905', 75, '部门1', NOW(), NOW());

```



## 开始使用

下载Gorm-Plus

```SQL
 go get github.com/acmestack/gorm-plus
```



```Go
package main

import (
  "github.com/acmestack/gorm-plus/gplus"
  "gorm.io/driver/mysql"
  "gorm.io/gorm"
  "gorm.io/gorm/logger"
  "log"
  "time"
)

type User struct {
  ID        int64
  Username  string
  Password  string
  Address   string
  Age       int
  Phone     string
  Score     int
  Dept      string
  CreatedAt time.Time
  UpdatedAt time.Time
}

var gormDb *gorm.DB

func init() {
  dsn := "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
  var err error
  gormDb, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
  })
  if err != nil {
    log.Println(err)
  }

  // 初始化gplus
  gplus.Init(gormDb)
}

func main() {
  users, resultDb := gplus.SelectList[User](nil)
  log.Println("error:", resultDb.Error)
  log.Println("RowsAffected:", resultDb.RowsAffected)
  for _, user := range users {
    log.Println("user:", user)
  }
}

```



控制台输出：

```Go
2023/06/01 17:48:19 error: <nil>
2023/06/01 17:48:19 RowsAffected: 5
2023/06/01 17:48:19 user: &{1 张三 password1 地址1 25 12345678901 80 部门1 2023-06-01 17:48:11 +0800 CST 2023-06-01 17:48:11 +0800 CST}
2023/06/01 17:48:19 user: &{2 李四 password2 地址2 30 12345678902 90 部门2 2023-06-01 17:48:11 +0800 CST 2023-06-01 17:48:11 +0800 CST}
2023/06/01 17:48:19 user: &{3 王五 password3 地址3 35 12345678903 70 部门1 2023-06-01 17:48:11 +0800 CST 2023-06-01 17:48:11 +0800 CST}
2023/06/01 17:48:19 user: &{4 赵六 password4 地址4 28 12345678904 85 部门2 2023-06-01 17:48:11 +0800 CST 2023-06-01 17:48:11 +0800 CST}
2023/06/01 17:48:19 user: &{5 钱七 password5 地址5 32 12345678905 75 部门1 2023-06-01 17:48:11 +0800 CST 2023-06-01 17:48:11 +0800 CST}

```

## 搜索工具

只需要下面一行代码即可完成单表的所有查询功能

```Bash
gplus.SelectList(gplus.BuildQuery[User](queryParams))
```



例子：

```Bash
func main() {
  http.HandleFunc("/", handleRequest)
  http.ListenAndServe(":8080", nil)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
  queryParams := r.URL.Query()
  list, _ := gplus.SelectList(gplus.BuildQuery[User](queryParams))
  marshal, _ := json.Marshal(list)
  w.Write(marshal)
}
```

假设我们要查询username为zhangsan的用户

```Bash
http://localhost:8080?q=username=zhangsan
```



假设我们要查询username姓zhang的用户

```Bash
http://localhost:8080?q=username~>=zhang
```



假设我们要查询age大于20的用户

```Bash
http://localhost:8080?q=age>20
```



假设我们要查询username等于zhagnsan，password等于123456的用户

```Bash
http://localhost:8080?q=username=zhangsan&q=password=123456
```



假设我们要查询username等于zhagnsan，password等于123456的用户

```Bash
http://localhost:8080?q=username=zhangsan&q=password=123456
```



假设我们要查询username等于zhagnsan，或者usename等于lisi的用户

可以增加一个分组和gcond的条件查询来实现

```Bash
http://localhost:8080?q=A.username=zhangsan&q=B.username=lisi&gcond=A|B
```



所有的单表查询我们都只需要一行代码即可。



## 总结

从上述步骤中，我们可以看到集成`Gorm-Plus`非常简单。只需在初始化`Gorm`之后添加一行代码`gplus.Init(gormDb)`

即可使用。不仅如此，使用`Gorm-Plus`也同样轻松，只需一行代码即可完成列表查询。

然而，`Gorm-Plus`的强大功能远不止于此。

更多文档请查看: [https://github.com/acmestack/gorm-plus/wiki](https://github.com/acmestack/gorm-plus/wiki)

## Star History
<a href="https://star-history.com/#acmestack/gorm-plus&Date">
  <picture>
    <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=acmestack/gorm-plus&type=Date" />
  </picture>
</a>

## 技术交流群

加微信：afumubit，注意备注：gplus




