### SIP部分包过滤规则

* INVITE包过滤规则

  ```
  1、UDP
  2、UDP数据以"INVITE sip:"开头
  ```

* INVITE的200 OK包过滤规则
  ```
  1、UDP
  2、UDP数据以"SIP/2.0 200 OK"开头
  3、UDP数据包含"\r\nCSeq: [0-9]* INVITE\r\n"（正则表达式）
  ```

* BYE的200 OK包过滤规则
  ```
  1、UDP
  2、UDP数据以"SIP/2.0 200 OK"开头
  3、UDP数据包含"\r\nCSeq: [0-9]* BYE\r\n"（正则表达式）
  ```

### 使用tcpreplay
```shell
    # -i: 指定网络接口
    # -l: 回放文件多少次
    # -p: 回放包速率，单位：pps
    # -M: 回放包速率，单位：Mbps
    tcpreplay -i en0 -l 1 -p 100 20201015.pcapng
```