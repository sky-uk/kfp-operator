import keras

@keras.saving.register_keras_serializable()
class MyModel(keras.Model):
    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        self.dense1 = keras.layers.Dense(32, activation="relu") # HL
        self.dense2 = keras.layers.Dense(3, activation="softmax") # Ouput

    def call(self, inputs):
        x = self.dense1(inputs)
        return self.dense2(x)
