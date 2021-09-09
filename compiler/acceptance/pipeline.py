from tfx.components import CsvExampleGen, Trainer
from tfx.proto import trainer_pb2

def create_components():
    example_gen = CsvExampleGen(input_base="/tmp")
    trainer = trainer = Trainer(
      examples=example_gen.outputs['examples'],
      run_fn='module.runFn',
      train_args=trainer_pb2.TrainArgs(num_steps=100),
      eval_args=trainer_pb2.EvalArgs(num_steps=5))

    return [example_gen, trainer]