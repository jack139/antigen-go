# coding=utf-8

# 导出 locate 模型权重，可用于go tf.LoadSavedModel


import os, shutil
from keras.backend.tensorflow_backend import tf
from model_resnet_fpn import get_model as get_model_fpn

os.environ["CUDA_VISIBLE_DEVICES"] = '0'

# ------- locate model
input_size = (256,256,3)
model = get_model_fpn(input_size=input_size, weights=None) # fpn
model.load_weights("../../antigen/ckpt/locate_onebox_resnet-fpn_b128_e24_0.94362.h5")
model.summary()


config = tf.ConfigProto()
config.allow_soft_placement = True
config.gpu_options.allow_growth = True

save_model_path = "outputs/saved-model_locate"
if os.path.exists(save_model_path):
    shutil.rmtree(save_model_path) 
os.makedirs(save_model_path)

with tf.Session(config=config) as sess:
    sess.run(tf.global_variables_initializer())

    print('\n'.join([n.name for n in tf.get_default_graph().as_graph_def().node])) # 所有层的名字

    # save_model 输出 , for goland 测试
    builder = tf.saved_model.builder.SavedModelBuilder(save_model_path)
    builder.add_meta_graph_and_variables(sess, [tf.saved_model.tag_constants.TRAINING], clear_devices=True)
    builder.save()  
