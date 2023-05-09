## 1关于openapi定义
基础定义遵循restFul api定义
### 1.1通用openapi需要遵循的原则或规范
*   需要版本号，放入path，如http://api.example.com/v1
*   分页查询，请求方法GET，page_num(页码)，page_size(每页显示条数)
*   返回结构一致
*   公共错误码
### 1.2细化参考
#### 1.2.1为避免歧义，名词采用复数
#### 1.2.2路径名词采用snake编码规范，不采用驼峰
#### 1.2.3路径尽可能以单名词为主
如areas，services
```text
/v1/areas
/v1/services
/v1/instances
```
api示例参考 internal/apis/area/area.go
#### 1.2.4当出现多名词路由表达时，允许名词嵌套，如网卡绑定安全组、vip绑定网卡、实例绑定云盘
```text
/v1/vips/:vip_id/netcards/:netcard_id
```
api示例参考 **待添加**
#### 1.2.5当出现action类型的路由时,变成多名名词
```text
/v1/services/:id/operation/:name
/v1/services/:id/operation/start
/v1/services/:id/operation/stop
```
大部分场景下，为区分创建(POST)和查询列表(GET),请求方法采用PUT,DELETE,当然根据语义的不同，需采用不同的方法

#### 1.2.6返回数据格式
*   正常返回
```json
{
    "id": 404769890650566785,
    "name": "test",
    "created_at": "2022-04-24T09:07:35.482Z",
    "status": "creating"
}
```
*   错误返回
```json
{
    "code": "Devt.5000000000",
    "message":"服务器内部错误"
}
```
#### 1.2.7公共请求头
|字段名|释义|
|-----|-----|
|X-Account-ID|账户ID|
|X-User-ID| 用户ID|
|X-Source|请求来源|
#### 1.2.8错误码定义
10位字符串组成,由3位http状态码+3位服务编码+4位错误码
定义参考 internal/code/code.go
#### 1.2.9排序设计
请求参数增加sort字段，内容如下
```text
/v1/services?sort=created_at
/v1/services?sort=created_at asc
/v1/services?sort=created_at ASC
/v1/services?sort=created_at desc
/v1/services?sort=created_at DESC
/v1/services?sort=created_at,updated_at asc
/v1/services?sort=created_at DESC,updated_at asc
```
#### 1.2.10批量接口设计
基于单名词增加请求参数batch
```text
PUT /v1/instance?batch={"ids":["aa","bb","cc"]}
```
