# hack-fdu-2016
API instructions and examples for hack-fdu-2016

### 1)语法检错服务
Dependency: 需要配置 [grpc](http://www.grpc.io/) 和 [protobuf](https://developers.google.com/protocol-buffers/)

输入一段文本，返回这句话可能有的语法错误。比如输入 i do not play the football，返回 冠词使用错误：在the football 中应删除冠词the

[proto file](https://github.com/yxf0605/hack-fdu-2016/blob/master/grammar_res/grammar_service.proto)

返回结果(样例)：
```
输入：There is a lot of book.
输出（Unicode）: 主谓一致错误:There is a中的is。

```

服务器/端口号：54.222.198.42:50054

参考代码：

```python
"""The Python implementation of the GRPC grammar_service.GrammarService client."""

import grpc
import grammar_service_pb2

_TIMEOUT_SECONDS = 10

def grammar_correct():
    channel = grpc.insecure_channel('%s:%d' % ('0.0.0.0', 50054))
    stub = grammar_service_pb2.GrammarServiceStub(channel)
    while True:
        txt = raw_input("Input>>>")
        if txt == 'exit': break
        response = stub.GrammarCorrect(grammar_service_pb2.GrammarCorrectRequest(content = txt, error_type = ""), _TIMEOUT_SECONDS)
        print "Correct result", response.message

if __name__ == '__main__':
    grammar_correct()

```

### 2)语义相似度计算服务
Dependency: 需要配置 [grpc](http://www.grpc.io/) 和 [protobuf](https://developers.google.com/protocol-buffers/)

输入两句话，返回这两句话在语义上的相似度。比如输入Obama speaks to the media in Illinois 和 The President addresses the press in Chicago，返回 0.777210439668(相似度值落在[0,1]区间内，值越大表示越相似)

[proto file](https://github.com/yxf0605/hack-fdu-2016/blob/master/semantic_res/semantic_sim.proto)


参考代码：
```python
"""The Python implementation of the GRPC semantic_sim.SemanticSim client."""

import grpc
import semantic_sim_pb2

_TIMEOUT_SECONDS = 10

def get_similarity(sent_pair):
    channel = grpc.insecure_channel('%s:%d' % ('0.0.0.0', 50067))
    stub = semantic_sim_pb2.SemanticSimStub(channel)
    response = stub.Communicate(semantic_sim_pb2.ASRRequest(message=sent_pair), _TIMEOUT_SECONDS)
    print "Similarity: " + response.message

if __name__ == '__main__':
    sent1 = raw_input('Sentence 1: ')
    sent2 = raw_input('Sentence 2: ')
    sent_pair = sent1+'###'+sent2
    get_similarity(sent_pair) 

```

服务器/端口号：54.222.198.42:50067

返回结果格式：
```
浮点数的 unicode 字符串格式
```
### 3)语音识别服务
目前支持英语的ASR，支持的音频格式是 16k 采样率，mono wav

语音数据使用HTTP请求，以stream方式上传到服务器，遵循的协议格式如下

输入格式为:

|  META_LEN	 |  META | STREAM  | EOS  |
|---|---|---|---|
META_LEN是32bit整数,表示接下来的META的长度,用于传递给后面的计算服务
META是后面的计算服务需要的metadata(base64(json)),比如说音频采样率之类的
EOS设定值是"0x450x4f0x53"，表示End Of Stream


输出格式为:

| META_LEN	 |  META |
|---|---|
META_LEN是32bit整数,表示接下来的META的长度,用于返回给JS
META是一个JSON结构,比如:{code:0, msg:"ok", key:"messageid", val:base64(json), extra: base64(json), flag:0},直接返回给JS即可
其中val是经过base64的json值,flag为0表示中间计算结果,为1表示最终计算结果(在输入音频的同时，也可以有中间结果返回), extra 是经过 base64 的 json 值，通常存放和业务相关的内容等
 
下面是C/C++ 请求的例子，websocket使用Poco library

```cpp
std::ifstream meta_ifs;
meta_ifs.open(meta_file.c_str());
std::stringstream ss;
ss << meta_ifs.rdbuf();
meta_ifs.close();
std::string meta_data_cpp = ss.str();
char const* meta_data = meta_data_cpp.c_str();
int data_length = strlen(meta_data);
unsigned int encoded_data_length = Base64encode_len(data_length); //include a null terminator
char* base64_meta_data = (char*)malloc(encoded_data_length);
Base64encode(base64_meta_data, meta_data, data_length);
WebSocket* m_psock = new WebSocket(cs, request, response);
m_psock->setReceiveTimeout(Poco::Timespan(600,0));
int bigendian = htonl(encoded_data_length - 1);
m_psock->sendFrame(&bigendian, 4, WebSocket::FRAME_BINARY);
int len = m_psock->sendFrame(base64_meta_data, encoded_data_length - 1, WebSocket::FRAME_BINARY);
//std::cout << "Sent bytes " << len << std::endl;
free(base64_meta_data);

if(! infilename.empty()) {
    short buf[CHUNK_SIZE];
    size_t out_size;
    FILE *fp = fopen(infilename.c_str(), "rb");
    while(1) {
        out_size = fread(buf, sizeof(*buf), CHUNK_SIZE, fp);
        if(out_size == 0) break;
        len = m_psock->sendFrame(buf, out_size * 2, WebSocket::FRAME_BINARY);
    }
    fclose(fp);
}

char eos[] = {0x45, 0x4f, 0x53};
// 0x45; 'E'
// 0x4f; 'O'
// 0x53; 'S'
len = m_psock->sendFrame(eos, 3, WebSocket::FRAME_BINARY);

int flags = 0;
int max_recv_size = 1024*1024;
char *recv_buf = (char*)malloc(max_recv_size);
int recv_len = m_psock->receiveFrame(recv_buf, max_recv_size, flags);
recv_buf[recv_len] = 0;
std::ofstream ofs;
ofs.open(outfilename.c_str());
//skip the first 4 bytes because it's header size
ofs << ExtractResultFromBase64Response(recv_buf + 4);
ofs.close();
m_psock->close();
free(recv_buf);
```
[JS example](https://github.com/lingochamp/hack-fdu-2016/tree/master/js-asr)

META文件内容:
```json
{
    "quality":-1,
    "type":"asr"
}
```

服务器/端口/URL: 54.223.187.43:8281 /llcup/stream/upload

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

### 4)句子评分服务
语音数据的请求方式和语音识别一样，唯一不同的地方是传入的META文件内容不同
META文件内容
```
{
    "quality": -1,
    "type": "readaloud",
    "reftext": "hello nice to meet you" //要评分的句子的文本，去掉标点符号，全部转换成小写
}
```

服务器/端口/URL: 54.223.187.43:8281 /llcup/stream/upload

[返回结果](https://github.com/yxf0605/hack-fdu-2016/blob/master/readaloud.json)
