# -*- coding: utf-8 -*-
"""
Version: 2.7
Date: 2017-5-6
author: Vast

语音识别服务 python实现
"""
import base64
import websocket
import struct
import wave
import audioop
def downsampleWav(src, dst, inrate=44100, outrate=16000, inchannels=2, outchannels=1):
    """
    改变音频文件
    :param src:源文件
    :param dst: 输出文件
    :param inrate: 原采样率
    :param outrate: 目标采样率
    :param inchannels: 原声道
    :param outchannels: 目标声道
    :return: Bool
    """
    import os
    if not os.path.exists(src):
        print 'Source not found!'
        return False

    # if not os.path.exists(os.path.dirname(dst)):
    #     os.makedirs(os.path.dirname(dst))
    try:
        s_read = wave.open(src, 'r')
        s_write = wave.open(dst, 'w')
    except:
        print 'Failed to open files!'
        return False

    s_read = wave.open(src, 'r')
    s_write = wave.open(dst, 'w')

    n_frames = s_read.getnframes()
    data = s_read.readframes(n_frames)

    # try:
    #     converted = audioop.ratecv(data, 2, inchannels, inrate, outrate, None)
    #     if outchannels == 1:
    #         converted = audioop.tomono(converted[0], 2, 1, 0)
    # except:
    #     print 'Failed to downsample wav'
    #     return False
    #     converted = audioop.ratecv(data, 2, inchannels, inrate, outrate, None)
    #     if outchannels == 1:
    #         converted = audioop.tomono(converted[0], 2, 1, 0)

    converted = audioop.ratecv(data, 2, inchannels, inrate, outrate, None)
    if outchannels == 1:
        converted = audioop.tomono(converted[0], 2, 1, 0)

    try:
        s_write.setparams((outchannels, 2, outrate, 0, 'NONE', 'Uncompressed'))
        s_write.writeframes(converted)
    except:
        print 'Failed to write wav'
        return False

    try:
        s_read.close()
        s_write.close()
    except:
        print 'Failed to close wav files'
        return False

    return True

def Voice2Text(file):
    """
    调用流利说API，把音频文件转换成字符串
    :param file: 音频文件,要求wav格式
    :return: 字符串
    """
    downsampleWav(file,'out.wav')
    先更改格式
    file='out.wav'
    META = """
    {
        "quality":-1,
        "type":"asr"
    }
        """
    META_BASE64 = base64.standard_b64encode(META)
    META_LEN = len(META_BASE64)
    EOS = 'EOS'
    url = 'wss://rating.llsstaging.com/llcup/stream/upload'
    ws = websocket.create_connection(url, subprotocols=["binary"])
    ws.send(struct.pack('>L', META_LEN))  # 发送网络序
    ws.send(META_BASE64)  # 发送META_BASE64

    with open(file, 'rb') as f:
        ws.send(f.read())
    ws.send(base64.standard_b64encode(EOS))

    data = ws.recv()
    data=data[4:]
    print data
    #{"status":0,"msg":"","reqId":"","key":"3415fadb2ddc6a8c435e589bafb8583e","result":"d2VsY29tZSB0byBiZWlqaW5n","flag":0}
    # print data,'\n',data[data.find('result')+9:data.find('flag":')-3]
    # print base64.b64decode(data[data.find('result')+9:data.find('flag":')-3])
    return base64.b64decode(data[data.find('result')+9:data.find('flag":')-3])




# if __name__ == '__main__':
#     Voice2Text("")