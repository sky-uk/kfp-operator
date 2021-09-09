import click
import importlib.util
import os
import sys

from tfx.components import Pusher, Trainer
from tfx.proto import pusher_pb2

def expand_components_with_pusher(tfx_components, serving_model_directory):
    if not any(isinstance(component, Pusher) for component in tfx_components):

        click.secho(
            "Could not find Pusher component. Trying to expand. Expanding", fg="green"
        )

        trainers = (
            comp
            for comp in tfx_components
            if isinstance(comp, Trainer) and "model" in comp.outputs
        )
        trainer = next(trainers, None)

        if next(trainers, None):
            click.secho("Found more than one Trainer component. aborting", fg="red")
            sys.exit(1)

        if trainer:
            model = trainer.outputs["model"]
        else:
            return tfx_components

        evaluators = (comp for comp in tfx_components if "blessing" in comp.outputs)
        evaluator = next(evaluators, None)

        if next(evaluators, None):
            click.secho("Found more than one Evaluator component. aborting", fg="red")
            sys.exit(1)

        blessing = evaluator.outputs["blessing"] if evaluator else None

        pusher = Pusher(
            model=model,
            model_blessing=blessing,
            push_destination=pusher_pb2.PushDestination(
                filesystem=pusher_pb2.PushDestination.Filesystem(
                    base_directory=serving_model_directory
                )
            ),
        )

        return tfx_components + [pusher]
    else:
        return tfx_components