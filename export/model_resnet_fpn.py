# coding=utf-8

# Panoptic FPN 1901.02446 + Resnet50 v2

from keras.applications import ResNet50V2
from keras.models import Model
from keras.layers import *


def get_backbone_ResNet50(input_shape, weights):
    """Builds ResNet50 with pre-trained imagenet weights"""
    backbone = ResNet50V2(include_top=False, input_shape=input_shape, weights=weights)
    c3_output, c4_output, c5_output = [
        backbone.get_layer(layer_name).output
        for layer_name in ["conv3_block4_out", "conv4_block6_out", "conv5_block3_out"]
    ]
    return backbone.input, [c3_output, c4_output, c5_output]


def get_model(input_size = (256,256,3), weights='imagenet'):
    base_inputs, base_outputs = get_backbone_ResNet50(input_shape=input_size, weights=weights)
    c3_output, c4_output, c5_output = base_outputs

    p3_output = Conv2D(256, 1, strides=1, activation='relu', padding="same")(c3_output)
    p4_output = Conv2D(256, 1, strides=1, activation='relu', padding="same")(c4_output)
    p5_output = Conv2D(256, 1, strides=1, activation='relu', padding="same")(c5_output)

    p4_output = Add()([p5_output, p4_output])
    p3_output = Add()([UpSampling2D(2)(p4_output), p3_output])

    p3_output = Conv2D(256, 3, strides=1, activation='relu', padding="same")(p3_output)
    p4_output = Conv2D(256, 3, strides=1, activation='relu', padding="same")(p4_output)
    p5_output = Conv2D(256, 3, strides=1, activation='relu', padding="same")(p5_output)

    p3_output = Flatten()(p3_output)
    p4_output = Flatten()(p4_output)
    p5_output = Flatten()(p5_output)

    m1_output = Concatenate(axis=1)([p3_output,
                                     p4_output,
                                     p5_output])

    m1_output = Dense(256, activation='relu')(m1_output)
    m1_output = Dense(64, activation='relu')(m1_output)
    m1_output = Dense(4, activation='sigmoid')(m1_output)

    model = Model(inputs=base_inputs, outputs=m1_output)

    return model
