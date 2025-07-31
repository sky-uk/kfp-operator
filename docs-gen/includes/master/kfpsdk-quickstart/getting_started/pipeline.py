from kfp import dsl
from kfp.dsl import Input, Output, Artifact
from typing import List

PIPELINE_NAME = "kfp-operator-kfpsdk-quickstart"
BASE_IMAGE = f"europe-docker.pkg.dev/ml-kfp-operator-sandbox-100/mlops/{PIPELINE_NAME}:latest"

@dsl.component(base_image=BASE_IMAGE)
def get_training_examples(
    input_dataset_uri: str,
    examples: Output[Artifact],
    columns: List[str] = None,
):
    import polars as pl
    from sklearn.model_selection import train_test_split

    df = pl.read_csv(input_dataset_uri, has_header=False, new_columns=columns).drop_nulls()
    n_cols = len(columns)

    # Split dataset X = features, y = labels
    X = df.select(df.columns[:(n_cols - 1)])
    y = df.get_column("species")

    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=2)

    # Save in Parquet format
    X_train.write_parquet(examples.path + "X_train.parquet")
    X_test.write_parquet(examples.path + "X_test.parquet")

    # Save labels as Parquet
    # Convert Series to DataFrame before saving as Parquet
    y_train.to_frame().write_parquet(examples.path + "y_train.parquet")
    y_test.to_frame().write_parquet(examples.path + "y_test.parquet")


@dsl.component(base_image=BASE_IMAGE)
def transform_examples(
    examples: Input[Artifact],
    transformed_examples: Output[Artifact],
):
    import polars as pl
    import numpy as np
    import torch
    from sklearn.preprocessing import StandardScaler, LabelEncoder

    # Load X from Parquet
    X_train = pl.read_parquet(examples.path + "X_train.parquet").to_numpy()
    X_test = pl.read_parquet(examples.path + "X_test.parquet").to_numpy()

    # Load from Parquet and convert back to NumPy
    y_train = pl.read_parquet(examples.path + "y_train.parquet").to_numpy().flatten()
    y_test = pl.read_parquet(examples.path + "y_test.parquet").to_numpy().flatten()


    # Transform X
    scaler = StandardScaler()
    X_scaled_train = scaler.fit_transform(X_train)
    X_scaled_test = scaler.transform(X_test)

    # Transform Y - Encode string labels to integers
    label_encoder = LabelEncoder()
    y_train_encoded = label_encoder.fit_transform(y_train)
    y_test_encoded = label_encoder.transform(y_test)

    # Convert to Torch Tensors
    X_train_tensor = torch.tensor(X_scaled_train, dtype=torch.float32)
    X_test_tensor = torch.tensor(X_scaled_test, dtype=torch.float32)
    y_train_tensor = torch.tensor(y_train_encoded, dtype=torch.long)
    y_test_tensor = torch.tensor(y_test_encoded, dtype=torch.long)

    # Save using Torch native format
    torch.save(X_train_tensor, transformed_examples.path + "X_train.pt")
    torch.save(X_test_tensor, transformed_examples.path + "X_test.pt")
    torch.save(y_train_tensor, transformed_examples.path + "y_train.pt")
    torch.save(y_test_tensor, transformed_examples.path + "y_test.pt")

@dsl.component(base_image=BASE_IMAGE)
def trainer(
    transformed_examples: Input[Artifact],
    model: Output[Artifact],
):
    import os
    os.environ["KERAS_BACKEND"] = "torch"  # Must be set before Keras import
    import keras
    import torch
    from getting_started.model import MyModel

    # Load tensors using torch
    x_train = torch.load(transformed_examples.path + "X_train.pt")
    y_train = torch.load(transformed_examples.path + "y_train.pt")

    model_obj = MyModel()

    model_obj.compile(
        optimizer=keras.optimizers.Adam(learning_rate=1e-3),
        loss=keras.losses.SparseCategoricalCrossentropy(),  # Integer labels
        metrics=[keras.metrics.SparseCategoricalAccuracy()],
    )

    model_obj.fit(
        x=x_train.numpy(),
        y=y_train.numpy(),
        batch_size=None,
        epochs=3,
        verbose="auto",
    )

    model_obj.summary()

    local_model_path = "./model_test.keras"

    # Save trained model
    model_obj.save(local_model_path)

@dsl.pipeline(
    description='A KFP SDK quickstart pipeline which performs training on the Iris dataset.',
)
def add_pipeline():

    examples_task = get_training_examples(
        input_dataset_uri= 'https://archive.ics.uci.edu/ml/machine-learning-databases/iris/iris.data',
        columns = [
            'sepal_length', 'sepal_width', 'petal_length', 'petal_width', 'species'
        ]
    )


    transform_examples_task = transform_examples(
        examples = examples_task.outputs['examples']

    )

    trainer_task = trainer(
        transformed_examples = transform_examples_task.outputs['transformed_examples'],
    )
