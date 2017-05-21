# hack-sjtu-2017
API instructions and examples for hack-sjtu-2017

## 1)语音识别服务

目前支持英语的ASR，支持的音频格式是 16k 采样率，mono wav

语音数据使用HTTP请求，以stream方式上传到服务器，遵循的协议格式如下

### 输入格式

| Item     | Description                              |
| -------- | ---------------------------------------- |
| META_LEN | META_LEN是32bit整数(大端序),表示接下来的META的长度,用于传递给后面的计算服务. |
| META     | 内容如下. 以BASE64编码.                         |
| STREAM   | 二进制音频文件                                  |
| EOS      | EOS设定值是"0x450x4f0x53", 表示End Of Stream   |

META文件内容:

```json
{
    "quality":-1,
    "type":"asr"
}
```



### 输出格式

| Item     | Description                              |
| -------- | ---------------------------------------- |
| META_LEN | META_LEN是32bit整数,表示接下来的META的长度,用于返回给JS   |
| META     | META是一个JSON结构:{code:0, msg:"ok", key:"messageid", val:base64(json), extra: base64(json), flag:0},直接返回给JS即可 |

其中val是经过base64的json值,flag为0表示中间计算结果,为1表示最终计算结果(在输入音频的同时，也可以有中间结果返回), extra 是经过 base64 的 json 值，通常存放和业务相关的内容等

返回结果：
```
{
   "confidence": 84, //整体置信度
   "decoded": "hackers weekend ", //整体识别结果
   "details": [
      {
         "confidence": 97, //识别的置信度0~100，数值越大表明越可信
         "end": 171, //单词结束时间
         "start": 111, //单词起始时间
         "word": "hackers" //识别出的单词
      },
      {
         "confidence": 70,
         "end": 177,
         "start": 171,
         "word": "weekend"
      }
   ]
}
```

服务器/端口/URL: wss://rating.llsstaging.com/llcup/stream/upload



## 2)句子评分服务

语音数据的请求方式和语音识别一样，唯一不同的地方是传入的META文件内容不同
META文件内容
```
{
    "quality": -1,
    "type": "readaloud",
    "reftext": "hello nice to meet you" //要评分的句子的文本，去掉标点符号，全部转换成小写
}
```

服务器/端口/URL: wss://rating.llsstaging.com/llcup/stream/upload

[返回结果](./readaloud.json)

## Example

[Python](./Example/demo.py)

[C/C++ (websocket使用Poco library)](./Example/C.cpp)

[Go](./Example/ws_client.go)

[JS example](./Example/js-asr)

