from kfp import dsl

@dsl.component(base_image='python:3.9')
def add(a: float, b: float) -> float:
    """Calculates sum of two arguments"""
    return a + b

@dsl.component(base_image="python:3.9")
def write_result(value: float, result: dsl.Output[dsl.Model]):
    """Writes the final result into an artifact file"""
    with open(result.path, "w") as f:
        f.write(str(value))

@dsl.pipeline(
    name="Addition pipeline",
    description="Pipeline that performs addition and outputs an artifact"
)
def add_pipeline(a: float = 1.0, b: float = 7.0):
    first = add(a=a, b=4.0)
    second = add(a=first.output, b=b)
    write_result(value=second.output)
