from kfp import dsl


@dsl.component
def add(a: float, b: float) -> float:
    '''Calculates sum of two arguments'''
    return a + b


@dsl.pipeline(
    name='Addition pipeline',
    description='An example pipeline that performs addition calculations.')
def add_pipeline(
        a: float = 1.0,
        b: float = 7.0,
):
    first_add_task = add(a=a, b=4.0)
    second_add_task = add(a=first_add_task.output, b=b)
