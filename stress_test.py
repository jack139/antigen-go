# -*- coding: utf-8 -*-
#
import os
import time, hashlib, json, sys, random, base64
import threading
import urllib3

urllib3.disable_warnings()

TEST_SERVER = [
    '127.0.0.1:5000'
]

# 装入测试图片
img_path = 'data/stress'

img_data_pool = []

'''
filelist = os.listdir(img_path)
for file in filelist:
    with open(os.path.join(img_path, file), 'rb') as f:
        img_data = f.read()
        img_data_pool.append(base64.b64encode(img_data).decode('utf-8'))
'''

BODY = {
    'version'  : '1',
    'signType' : 'SHA256',
    'encType'  : 'plain',
    'data'     : {
        #'image'    : None, # 先置空
        'image'    : '',
        'corpus'   : "金字塔（英语：pyramid），在建筑学上是指锥体建筑物，著名的有埃及金字塔，还有玛雅卡斯蒂略金字塔、阿兹特克金字塔（太阳金字塔、月亮金字塔）等。",
        'question' : "金字塔是什么？",
    }
}


SCORE = {}

# 生成参数字符串
def gen_param_str(param1):
    param = param1.copy()
    name_list = sorted(param.keys())
    if 'data' in name_list: # data 按 key 排序, 中文不进行性转义，与go保持一致
        param['data'] = json.dumps(param['data'], sort_keys=True, ensure_ascii=False, separators=(',', ':'))
    return '&'.join(['%s=%s'%(str(i), str(param[i])) for i in name_list if str(param[i])!=''])


def main_loop(tname, test_server):
    global SCORE

    body = BODY.copy()

    appid = '66A095861BAE55F8735199DBC45D3E8E'
    unixtime = int(time.time())
    body['timestamp'] = unixtime
    body['appId'] = appid
    #body['data']['image'] = img_data_pool[random.randint(0,len(filelist)-1)]

    param_str = gen_param_str(body)
    sign_str = '%s&key=%s' % (param_str, '43E554621FF7BF4756F8C1ADF17F209C')
    signature_str =  base64.b64encode(hashlib.sha256(sign_str.encode('utf-8')).hexdigest().encode('utf-8')).decode('utf-8')

    #print(sign_str)

    body['signData'] = signature_str

    body = json.dumps(body)

    #print(body)

    pool = urllib3.PoolManager(num_pools=2, timeout=180, retries=False)
    url = 'http://%s/api/bert_qa'%test_server

    tick_start = time.time()

    try:
        print(url)
        #print header

        r = pool.urlopen('POST', url, body=body)
        #print(r.data)
        #print(r.status)

        tick_end = time.time()

        time_used =  int((tick_end-tick_start)*1000)
        SCORE[tname] += time_used

        if r.status!=200:
            print(tname, test_server, '!!!!!! HTTP ret=', r.status, 'time_used=', time_used)
        else:
            print(tname, test_server, '200', 'time_used=', time_used, 'in_data_len=', len(body), 'out_data_len=', len(r.data))
    except Exception as e:
        print("异常: %s : %s" % (e.__class__.__name__, e))

class MainLoop(threading.Thread):
    def __init__(self, rounds):
        threading.Thread.__init__(self)
        self._tname = None
        self._round = rounds

    def run(self):
        global count, mutex, SCORE
        self._tname = threading.currentThread().getName()
        SCORE[self._tname] = 0        

        print('Thread - %s started.' % self._tname)

        #while 1:
        for x in range(0, self._round):
            for y in TEST_SERVER:
                main_loop(self._tname, y)

            # 周期性打印日志
            time.sleep(random.randint(0,1))
            sys.stdout.flush()


if __name__=='__main__':
    if len(sys.argv)<3:
        print("usage: python stress_test.py <thread_num> <round_per_thread>")
        sys.exit(2)

    print("STRESS TEST started: " , time.ctime())

    thread_num = int(sys.argv[1])
    round_per_thread = int(sys.argv[2])

    #线程池
    threads = []
        
    # 创建线程对象
    for x in range(0, thread_num):
        threads.append(MainLoop(round_per_thread))
    
    # 启动线程
    for t in threads:
        t.start()

    # 等待子线程结束
    for t in threads:
        t.join()  

    total = 0
    for i in SCORE.keys():
        total += SCORE[i]
        print('%s - %.3f'%( i, SCORE[i]/round_per_thread ))

    print('Average: %.3f'%(total/(thread_num*round_per_thread)) )

    print("STRESS TEST exited: ", time.ctime())
