# Go实现antigen推理和api服务



## 测试



### 编译

```
make
```



### 启动 dispatcher

```
build/antigen-go server ../../nlp/qa_demo
```



### 启动 http

```
build/antigen-go http 5000
```



### 测试脚本

```
python3 test_api.py 127.0.0.1 _
```
