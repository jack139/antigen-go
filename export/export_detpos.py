# coding=utf-8

# 导出 locate 模型权重，可用于go tf.LoadSavedModel


import os, shutil
from keras import backend as K
import tensorflow as tf
from keras.applications import VGG16
from keras.layers import Dense, GlobalAveragePooling2D, Dropout
from keras.models import Model

os.environ["CUDA_VISIBLE_DEVICES"] = '0'

# ------- detpos model
input_size = (128,128,3) 
base_model = VGG16(weights=None, input_shape=input_size, include_top=False)
x = base_model.output
x = GlobalAveragePooling2D()(x)
predictions = Dense(5, activation='softmax')(x)
model = Model(inputs=base_model.input, outputs=predictions)
model.load_weights("../../antigen/ckpt/detpos_5labels_vgg16_b512_e10_1.0000.h5")
model.summary()


save_model_path = "outputs/saved-model_detpos_5labels_vgg16_b512_e10_1.0000"
if os.path.exists(save_model_path):
    shutil.rmtree(save_model_path) 
os.makedirs(save_model_path)

print('\n'.join([n.name for n in tf.get_default_graph().as_graph_def().node])) # 所有层的名字

# save_model 输出 , for goland 测试
builder = tf.saved_model.builder.SavedModelBuilder(save_model_path)
builder.add_meta_graph_and_variables(K.get_session(), [tf.saved_model.tag_constants.TRAINING], clear_devices=True)
builder.save()  
