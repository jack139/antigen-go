import sys, os, glob, json, base64
from test_api import request


if __name__ == '__main__':
    if len(sys.argv)<2:
        print("usage: python %s <host> <img_path>"%sys.argv[0])
        sys.exit(2)

    if os.path.isdir(sys.argv[2]):
        file_list = glob.glob(sys.argv[2]+'/*')
    else:
        file_list = [sys.argv[2]]

    for ff in file_list:
        if os.path.isdir(ff):
            continue

        with open(ff, 'rb') as f:
            img_data = f.read()

        body = {
            'version'  : '1',
            #'signType' : 'SHA256', 
            'signType' : 'SM2',
            'encType'  : 'plain',
            'data'     : {
                'image'    : base64.b64encode(img_data).decode('utf-8'),
            }
        }

        r = request(sys.argv[1], body)

        if r.status==200:
            j = json.loads(r.data.decode('utf-8'))
            #print(j)
            if j['success']:
                print(ff, "-->", j['data']['comment'])
            else:
                print("fail -->", ff)
        else:
            print(r.data)
